package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
)

func isUserAnAdmin(i *discordgo.InteractionCreate) bool {
	userPerms := i.Member.Permissions
	return (userPerms & discordgo.PermissionManageGuild) != 0
}

func handleBet(
	i *discordgo.InteractionCreate,
	p *models.Prediction,
	state models.PredictionState,
) error {
	userId := i.Member.User.ID
	log.Printf("User %s clicked Bet button for prediction %d\n", userId, p.Id)

	if state != models.BettingOpen {
		return fmt.Errorf("Bets are closed.")
	}

	// TODO: show bet modal, collecting info on the target outcome and wager

	return nil
}

func handleResolve(
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

	// Can only resolve when betting is open or closed
	if state != models.PredictionResolved && state != models.PredictionCancelled {
		return fmt.Errorf("You can only resolve a prediction that hasn't been resolved or cancelled yet.")
	}

	// TODO: show resolve modal, collecting info on the winning outcome

	return nil
}

func handleCancel(
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

	// Can only resolve when betting is open or closed
	if state != models.PredictionResolved && state != models.PredictionCancelled {
		return fmt.Errorf("You can only cancel a prediction that hasn't been resolved or cancelled yet.")
	}

	// TODO: show a cancel modal which asks for a reason for cancelling the prediction

	return nil
}

func PredictionButtonHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	errorMessage := ""
	serverId := i.GuildID
	customId := i.MessageComponentData().CustomID

	predictionId, err := models.ExtractJacuzziIdFromCustomId(i)
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
			err = handleBet(i, p, state)
		} else if strings.HasPrefix(customId, "PredictionResolveButton|") {
			err = handleResolve(i, p, state)
		} else if strings.HasPrefix(customId, "PredictionCancelButton|") {
			err = handleCancel(i, p, state)
		} else {
			panic(fmt.Errorf("Unexpected prediction button custom ID: %s", customId))
		}
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't handle button press: %s", err.Error())
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
