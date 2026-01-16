package slashCommands

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

var timezoneRegions = map[string][]string{
	"Africa": {
		"Abidjan",
		"Accra",
		"Addis_Ababa",
		"Algiers",
		"Cairo",
		"Casablanca",
		"Johannesburg",
		"Lagos",
		"Nairobi",
	},

	"America": {
		"Anchorage",
		"Argentina/Buenos_Aires",
		"Bogota",
		"Chicago",
		"Denver",
		"Halifax",
		"Los_Angeles",
		"Mexico_City",
		"New_York",
		"Phoenix",
		"Sao_Paulo",
		"St_Johns",
		"Toronto",
		"Vancouver",
	},

	"Asia": {
		"Bangkok",
		"Dubai",
		"Hong_Kong",
		"Jakarta",
		"Jerusalem",
		"Kolkata",
		"Seoul",
		"Shanghai",
		"Singapore",
		"Tokyo",
	},

	"Australia": {
		"Adelaide",
		"Brisbane",
		"Darwin",
		"Melbourne",
		"Perth",
		"Sydney",
	},

	"Europe": {
		"Amsterdam",
		"Athens",
		"Berlin",
		"Brussels",
		"Dublin",
		"Helsinki",
		"Istanbul",
		"Lisbon",
		"London",
		"Madrid",
		"Paris",
		"Rome",
		"Stockholm",
		"Vienna",
		"Warsaw",
	},

	"Pacific": {
		"Auckland",
		"Chatham",
		"Fiji",
		"Guam",
		"Honolulu",
		"Noumea",
		"Port_Moresby",
	},
}

var regionChoices []*discordgo.ApplicationCommandOptionChoice

// Cache the list of keys in the map to use as choices in the command for the "region" option.
func getRegionChoices() []*discordgo.ApplicationCommandOptionChoice {
	if len(regionChoices) == 0 {
		for region := range timezoneRegions {
			regionChoices = append(regionChoices, &discordgo.ApplicationCommandOptionChoice{
				Name:  region,
				Value: region,
			})
		}
	}

	sort.Slice(regionChoices, func(i, j int) bool {
		return regionChoices[i].Name < regionChoices[j].Name
	})

	return regionChoices
}

// Allow users to set their timezone with the bot, which can be used for other features.
var SetTimezone = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name: "set-timezone",
		Description: fmt.Sprintf("Tell %s what your local timezone is (used by other features).",
			constants.BotName),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "region",
				Description: "Your region",
				Required:    true,
				Choices:     getRegionChoices(),
			},
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "city",
				Description:  "After selecting your \"region\", type any character to see the valid city options.",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		region := interaction.ApplicationCommandData().GetOption("region").StringValue()
		city := interaction.ApplicationCommandData().GetOption("city").StringValue()
		timezone := fmt.Sprintf("%s/%s", region, city)
		_, err := db.SetUserTimezone(interaction.GuildID, interaction.Member.User.ID, timezone)

		var message string
		if err == nil {
			message = fmt.Sprintf("Successfully configured user timezone as `%s`.", timezone)
		} else {
			message = fmt.Sprintf("Failed to configure user timezone: %s", err.Error())
		}
		log.Println(message)

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}

func TimezoneCityAutoComplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	regionOption := i.ApplicationCommandData().GetOption("region")
	// If region hasn't been provided yet, we can't suggest anything.
	if regionOption == nil {
		return
	}
	region := regionOption.StringValue()
	if region == "" {
		return
	}
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, tz := range timezoneRegions[region] {
		// Replace underscores with spaces for displaying to the end user.
		displayTz := strings.ReplaceAll(tz, "_", " ")
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  displayTz,
			Value: tz,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

func TimezoneTests() {
	log.Println("Testing timezone options for correctness:")
	tests := 0
	failures := 0
	for region, cities := range timezoneRegions {
		for _, city := range cities {
			tests += 1
			timezone := fmt.Sprintf("%s/%s", region, city)
			loc, err := time.LoadLocation(timezone)
			if err != nil {
				log.Printf("> BAD: %s did not convert to a location properly: %s\n", timezone, err.Error())
				failures += 1
			} else {
				log.Printf("> Good: %s\n", loc.String())
			}
		}
	}
	log.Printf("%d tests completed, %d failures.\n", tests, failures)
}
