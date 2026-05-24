package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

func parseOutcomes(input string) (map[rune]string, error) {
	lines := strings.Split(input, "\n")
	result := make(map[rune]string)
	index := 'A'
	alreadySeenOutcomes := utils.NewSet[string]()
	for _, line := range lines {
		if index > 'Z' {
			return result, fmt.Errorf("Can only have a maximum of 26 possible outcomes.")
		}
		if alreadySeenOutcomes.Contains(line) {
			return result, fmt.Errorf("Cannot have duplicate outcomes (%s).", line)
		}
		if line != "" {
			result[index] = line
			alreadySeenOutcomes.Add(line)
			index++
		}
	}
	return result, nil
}

func atoi(input string) int {
	num, err := strconv.Atoi(input)
	if err != nil {
		panic(err)
	}
	return num
}

func extractFirstCaptureAsInt(regex *regexp.Regexp, input string) int {
	captures := regex.FindStringSubmatch(input)
	numString := captures[1]
	return atoi(numString)
}

func extractCapturesAsDate(regex *regexp.Regexp, input string, timezone *time.Location) (time.Time, error) {
	captures := regex.FindStringSubmatch(input)
	monthString := captures[1]
	var month time.Month
	monthTime, err := time.Parse("January", monthString)
	if err != nil {
		monthTime, err = time.Parse("Jan", monthString)
		if err != nil {
			monthInt, err := strconv.Atoi(monthString)
			if err != nil || monthInt < 1 || monthInt > 12 {
				return time.Unix(0, 0), fmt.Errorf("Couldn't parse month \"%s\": %w", monthString, err)
			}
			month = time.Month(monthInt)
		}
	}
	if month == 0 {
		month = monthTime.Month()
	}
	dayString := captures[2]
	day := atoi(dayString)
	if day < 1 || day > 31 {
		return time.Unix(0, 0), fmt.Errorf("Day \"%s\" is invalid", dayString)
	}
	yearString := captures[3]
	year := atoi(yearString)
	return time.Date(year, month, day, 0, 0, 0, 0, timezone), nil
}

var minutesRegex = regexp.MustCompile(`^(\d+) min(ute)?s?$`)
var hoursRegex = regexp.MustCompile(`^(\d+) h(ou)?rs?$`)
var daysRegex = regexp.MustCompile(`^(\d+) days?$`)
var weeksRegex = regexp.MustCompile(`^(\d+) w(ee)?ks?$`)
var dateRegex = regexp.MustCompile(`^(\w+)[ -/](\d+)[\w,]*[ -/](\d+)$`)

func parseFutureTime(input string, now time.Time, timezone *time.Location) (time.Time, error) {
	now = now.In(timezone)

	switch {
	case minutesRegex.MatchString(input):
		minutes := extractFirstCaptureAsInt(minutesRegex, input)
		return now.Add(time.Duration(minutes) * time.Minute), nil
	case hoursRegex.MatchString(input):
		hours := extractFirstCaptureAsInt(hoursRegex, input)
		return now.Add(time.Duration(hours) * time.Hour), nil
	case daysRegex.MatchString(input):
		days := extractFirstCaptureAsInt(daysRegex, input)
		return now.Add(time.Duration(24*days) * time.Hour), nil
	case weeksRegex.MatchString(input):
		weeks := extractFirstCaptureAsInt(weeksRegex, input)
		return now.Add(time.Duration(24*7*weeks) * time.Hour), nil
	case dateRegex.MatchString(input):
		return extractCapturesAsDate(dateRegex, input, timezone)
	}

	return now, fmt.Errorf("Couldn't parse time value `%s`", input)
}

func PredicationCreateSubmitHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	data := interaction.ModalSubmitData()
	serverId := interaction.GuildID
	userId := interaction.Member.User.ID

	questionInput := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	outcomesInput := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	bettingDurationInput := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	expirationInput := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	log.Printf("questionInput = \"%s\"\n", questionInput)
	log.Printf("outcomesInput = \"%s\"\n", outcomesInput)
	log.Printf("bettingDurationInput = \"%s\"\n", bettingDurationInput)
	log.Printf("expirationInput = \"%s\"\n", expirationInput)

	errorMessage := ""

	// Sanitize
	outcomesInput = strings.ReplaceAll(outcomesInput, "\r\n", "\n")

	outcomes, err := parseOutcomes(outcomesInput)
	if err != nil {
		errorMessage = fmt.Sprintf("Couldn't parse outcomes: %s", err.Error())
	} else if len(outcomes) < 2 {
		errorMessage = "You must specify at least 2 outcomes (one on each line)"
	}

	var id models.JacuzziId
	if errorMessage == "" {
		id, err = db.ClaimNextJacuzziId(serverId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't reserve next JacuzziId: %s", err.Error())
		}
	}

	now := time.Now()

	user, err := db.GetUser(serverId, userId)
	if err != nil {
		errorMessage = fmt.Sprintf("Couldn't fetch user when trying to create prediction: %s", err.Error())
		return
	}

	var timezone *time.Location
	if errorMessage == "" {
		if user.Timezone == "" {
			errorMessage = "You need to call `/set-timezone` before calling this command."
		} else {
			timezone, err = time.LoadLocation(user.Timezone)
			if err != nil {
				errorMessage = fmt.Sprintf("User's timezone \"%s\" was invalid: %s", user.Timezone, err.Error())
			}
		}
	}

	var bettingCloseTime time.Time
	if errorMessage == "" {
		bettingCloseTime, err = parseFutureTime(bettingDurationInput, now, timezone)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't parse betting close time: %s", err.Error())
		} else if time.Until(bettingCloseTime) > time.Hour*24*31 {
			// If the betting close time is more than 1 month in the future
			errorMessage = "Betting close time needs to be within 1 month (31 days)."
		}
	}
	var expirationTime time.Time
	if errorMessage == "" {
		expirationTime, err = parseFutureTime(expirationInput, now, timezone)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't parse expiration time: %s", err.Error())
		} else if time.Until(expirationTime) > time.Hour*24*366 {
			// If the betting close time is more than 1 month in the future
			errorMessage = "Expiration time needs to be within 1 year (366 days)."
		}
	}
	if errorMessage == "" {
		resolutionDuration := expirationTime.Sub(bettingCloseTime)
		if resolutionDuration < 0 {
			errorMessage = "Expiration time must be after betting close time."
		} else if resolutionDuration < 10*time.Minute {
			errorMessage = "Betting close time and expiration time are too close together. " +
				"Please give your admins 10 minutes at *a bare minimum* to resolve the prediction. :smile:"
		}
	}

	// TODO: schedule betting close and expiration events

	var prediction models.Prediction
	if errorMessage == "" {
		prediction = models.Prediction{
			Id:               id,
			Question:         questionInput,
			CreationTime:     now,
			BettingCloseTime: bettingCloseTime,
			ExpirationTime:   expirationTime,
			Outcomes:         outcomes,
			Bets:             make(map[string]models.PredictionBet),
			State:            models.AcceptingBets,
		}
		err = db.StorePrediction(serverId, &prediction)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't store prediction in DB: %s", err.Error())
		}
	}

	var message string
	var flags discordgo.MessageFlags
	if errorMessage != "" {
		message = errorMessage
		flags = discordgo.MessageFlagsEphemeral
	} else {
		message = prediction.DisplayString()
	}

	// Must respond, or else the modal won't go away.
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   flags,
		},
	})
}
