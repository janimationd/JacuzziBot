package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows"
)

// Show bet modal, collecting info on the target outcome and wager
func showBetModal(i *discordgo.InteractionCreate, p *models.Prediction) error {
	return session.Handle.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("PredictionBetModal|%d", p.Id),
			Title:    fmt.Sprintf("Bet On Prediction #%d", p.Id),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "outcome",
							Label:       "The outcome ID you're betting on (A-Z)",
							Style:       discordgo.TextInputShort,
							Placeholder: "A",
							Required:    true,
							MinLength:   1,
							MaxLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "wager",
							Label:       "The point amount you're wagering (number)",
							Style:       discordgo.TextInputShort,
							Placeholder: "100",
							Required:    true,
							MinLength:   1,
							MaxLength:   50,
						},
					},
				},
			},
		},
	})
}

// Show resolve modal, collecting info on the winning outcome
func showResolveModal(i *discordgo.InteractionCreate, p *models.Prediction) error {
	return session.Handle.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("PredictionResolveModal|%d", p.Id),
			Title:    fmt.Sprintf("Resolve Prediction #%d", p.Id),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "outcome",
							Label:       "The ID of the winning outcome (A-Z)",
							Style:       discordgo.TextInputShort,
							Placeholder: "A",
							Required:    true,
							MinLength:   1,
							MaxLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "proof",
							Label:       "Any relevant proof of this outcome",
							Style:       discordgo.TextInputShort,
							Placeholder: "Perhaps a link?",
							Required:    false,
							MaxLength:   512,
						},
					},
				},
			},
		},
	})
}

// Show cancel modal, collecting info on the reason
func showCancelModal(i *discordgo.InteractionCreate, p *models.Prediction) error {
	return session.Handle.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("PredictionCancelModal|%d", p.Id),
			Title:    fmt.Sprintf("Cancel Prediction #%d", p.Id),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "reason",
							Label:       "Why are you cancelling the prediction?",
							Style:       discordgo.TextInputShort,
							Placeholder: "Misconfigured the prediction",
							Required:    true,
							MinLength:   1,
							MaxLength:   256,
						},
					},
				},
			},
		},
	})
}

func isUserAnAdmin(i *discordgo.InteractionCreate) bool {
	userPerms := i.Member.Permissions
	return (userPerms & discordgo.PermissionManageGuild) != 0
}

func handleBetButton(
	i *discordgo.InteractionCreate,
	p *models.Prediction,
	state models.PredictionState,
) error {
	userId := i.Member.User.ID
	log.Printf("User %s clicked Bet button for prediction %d\n", userId, p.Id)

	if state != models.BettingOpen {
		return fmt.Errorf("Bets are closed.")
	}

	if p.Bets[userId].Wager != 0 {
		return fmt.Errorf("You've already bet on this prediction.")
	}

	return showBetModal(i, p)
}

func handleResolveButton(
	i *discordgo.InteractionCreate,
	p *models.Prediction,
	state models.PredictionState,
) error {
	userId := i.Member.User.ID
	log.Printf("User %s clicked Resolve button for prediction %d\n", userId, p.Id)

	userIsAdmin := isUserAnAdmin(i)
	if !userIsAdmin {
		return fmt.Errorf("Only admins can resolve predictions.")
	}

	if state == models.PredictionResolved || state == models.PredictionCancelled {
		return fmt.Errorf("You can only resolve a prediction that hasn't been resolved or cancelled yet.")
	}

	return showResolveModal(i, p)
}

func handleCancelButton(
	i *discordgo.InteractionCreate,
	p *models.Prediction,
	state models.PredictionState,
) error {
	userId := i.Member.User.ID
	log.Printf("User %s clicked Cancel button for prediction %d\n", userId, p.Id)

	// Only the creator (before betting closes) or an admin can cancel.
	userIsCreatorAndBettingIsOpen := userId == p.Creator && state == models.BettingOpen
	if !userIsCreatorAndBettingIsOpen {
		userIsAdmin := isUserAnAdmin(i)
		if !userIsAdmin {
			return fmt.Errorf(
				"You can only cancel a prediction if you're its creator (and betting is still open) or you're an admin.")
		}
	}

	if state == models.PredictionResolved || state == models.PredictionCancelled {
		return fmt.Errorf("You can only cancel a prediction that hasn't been resolved or cancelled yet.")
	}

	return showCancelModal(i, p)
}

func handleBetModal(
	i *discordgo.InteractionCreate,
	p *models.Prediction,
	state models.PredictionState,
	now time.Time,
) error {
	serverId := i.GuildID
	userId := i.Member.User.ID
	log.Printf("User %s submitted Bet modal for prediction %d\n", userId, p.Id)

	if state != models.BettingOpen {
		return fmt.Errorf("Sorry, betting is now closed for this prediction.")
	}

	outcomeId := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	outcomeIdRune := []rune(outcomeId)[0]
	if outcomeIdRune < 'A' || outcomeIdRune >= rune('A'+len(p.Outcomes)) {
		return fmt.Errorf("Outcome \"%c\" is invalid for this prediction.", outcomeIdRune)
	}

	wager := i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	wagerFloat, err := strconv.ParseFloat(wager, 64)
	if err != nil || wagerFloat <= 0 {
		return fmt.Errorf("Wager \"%s\" is invalid. Wagers must be positive numbers.", wager)
	}

	// Try to deduct the wager from the user's points.
	user, err := db.ModifyUserPoints(serverId, userId, -wagerFloat)
	if err != nil {
		return fmt.Errorf("Couldn't place wager: %w", err)
	}

	bet := models.PredictionBet{
		OutcomeId: outcomeIdRune,
		Wager:     wagerFloat,
	}
	p, err = db.PlaceBetOnPrediction(serverId, p.Id, userId, bet)
	if err != nil {
		// Refund their wager
		db.ModifyUserPoints(serverId, userId, wagerFloat)
		return err
	}
	workflows.CreateOrUpdatePredictionMessage(p, "", now)

	message := fmt.Sprintf("<@%s> has wagered **%s point%s** on Prediction %d's outcome %c: \"%s\". *They now have %s point%s.*",
		userId, utils.FormatUIFloat(wagerFloat), utils.Plural(wagerFloat), p.Id, outcomeIdRune,
		p.GetOutcome(outcomeIdRune), utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
	messageSend := discordgo.MessageSend{
		Content: message,
		Reference: &discordgo.MessageReference{
			MessageID: p.MessageId,
			ChannelID: p.ChannelId,
			GuildID:   serverId,
		},
	}
	_, err = session.Handle.ChannelMessageSendComplex(p.ChannelId, &messageSend)
	if err != nil {
		log.Printf("Couldn't send bet channel message: %s", err.Error())
	}

	// Respond to the interaction.
	return session.Handle.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Bet placed!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func PredictionInteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	errorMessage := ""
	serverId := i.GuildID
	var customId string
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		customId = i.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		customId = i.ModalSubmitData().CustomID
	}
	log.Printf("Prediction interaction with ID %s received.\n", customId)

	predictionId, err := models.ExtractJacuzziIdFromCustomId(customId)
	if err != nil {
		errorMessage = fmt.Sprintf("Couldn't extract prediction Id from button ID: %s", err.Error())
	}

	var p *models.Prediction
	if errorMessage == "" {
		p, err = db.GetPrediction(serverId, predictionId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't fetch prediction from DB: %s", err.Error())
		}
	}

	now := time.Now()

	if errorMessage == "" {
		state := p.GetState(now)
		if strings.HasPrefix(customId, "PredictionBetButton|") {
			err = handleBetButton(i, p, state)
		} else if strings.HasPrefix(customId, "PredictionResolveButton|") {
			err = handleResolveButton(i, p, state)
		} else if strings.HasPrefix(customId, "PredictionCancelButton|") {
			err = handleCancelButton(i, p, state)
		} else if strings.HasPrefix(customId, "PredictionBetModal|") {
			err = handleBetModal(i, p, state, now)
			// } else if strings.HasPrefix(customId, "PredictionResolveModal|") {
			// 	err = handleResolveModal(i, p, state)
			// } else if strings.HasPrefix(customId, "PredictionCancelModal|") {
			// 	err = handleCancelModal(i, p, state)
		} else {
			panic(fmt.Errorf("Unexpected prediction button custom ID: %s", customId))
		}
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't handle prediction interaction: %s", err.Error())
		}
	}

	if errorMessage != "" {
		// Respond with the error message. We assume no other interaction response has succeeded.
		session.Handle.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
