package workflows

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
)

// Get the list of message components that should be attached to the main prediction message. If initialSend is true,
// it'll create a message for the first time. Otherwise it'll update the one that's identified in the prediction.
// The prediction p will be updated with the channel message ID if there was no error.
func CreateOrUpdatePredictionMessage(p *models.Prediction, channelId string, now time.Time) error {
	innerComps := []discordgo.MessageComponent{}
	state := p.GetState(now)
	if state == models.BettingOpen {
		innerComps = append(innerComps, discordgo.Button{
			Label:    "Bet",
			Style:    discordgo.PrimaryButton,
			CustomID: "PredictionBetButton",
		})
	}
	if state != models.PredictionResolved && state != models.PredictionCancelled {
		innerComps = append(innerComps, discordgo.Button{
			Label:    "Resolve",
			Style:    discordgo.SuccessButton,
			CustomID: "PredictionResolveButton",
		})
		innerComps = append(innerComps, discordgo.Button{
			Label:    "Cancel",
			Style:    discordgo.DangerButton,
			CustomID: "PredictionCancelButton",
		})
	}

	comps := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: innerComps,
		},
	}

	if p.MessageId == "" {
		messageSend := &discordgo.MessageSend{
			Content:    p.DisplayString(),
			Components: comps,
		}
		message, err := session.Handle.ChannelMessageSendComplex(channelId, messageSend)
		if err != nil {
			return fmt.Errorf("Couldn't send initial prediction message: %w", err)
		}
		p.MessageId = message.ID
	} else {
		content := p.DisplayString()
		messageEdit := &discordgo.MessageEdit{
			ID:         p.MessageId,
			Channel:    channelId,
			Content:    &content,
			Components: &comps,
		}
		_, err := session.Handle.ChannelMessageEditComplex(messageEdit)
		if err != nil {
			return fmt.Errorf("Couldn't edit prediction message: %w", err)
		}
	}
	return nil
}
