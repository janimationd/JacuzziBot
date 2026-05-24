package slashCommands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

// Create a prediction with possible outcomes that users can bet on.
var CreatePrediction = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "create-prediction",
		Description: "Create a prediction with possible outcomes that users can bet on.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverId := interaction.GuildID
		userId := interaction.Member.User.ID
		errorMessage := ""

		user, err := db.GetUser(serverId, userId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't fetch user when trying to create prediction: %s", err.Error())
			return
		}

		if errorMessage == "" && user.Timezone == "" {
			errorMessage = "You need to call `/set-timezone` before calling this command."
		}

		if errorMessage == "" {
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "predictionCreateModal",
					Title:    "Create a Prediction",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "question",
									Label:       "The question you're posing",
									Style:       discordgo.TextInputShort,
									Placeholder: "Will Half Life 3 be announced in 2026?",
									Required:    true,
									MinLength:   1,
									MaxLength:   200,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "outcomes",
									Label:       "One possible outcome per line (max 26)",
									Style:       discordgo.TextInputParagraph, // multiline
									Placeholder: "Yes\nIt won't be called Half Life 3, but yes\nNo",
									Required:    true,
									MaxLength:   1000,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "bettingDuration",
									Label:       "How long until betting closes (max 1 month)",
									Style:       discordgo.TextInputShort,
									Placeholder: "X minutes|X hours|X days|X weeks|[Month] [Day] [Year]",
									Required:    true,
									MaxLength:   32,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "expiration",
									Label:       "When it's auto-cancelled/refunded (max 1yr)",
									Style:       discordgo.TextInputShort,
									Placeholder: "X minutes|X hours|X days|X weeks|[Month] [Day] [Year]",
									Required:    true,
									MaxLength:   64,
								},
							},
						},
					},
				},
			})
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't create Prediction modal: %s", err.Error())
			} else {
				log.Println("Prediction modal created.")
			}
		}

		if errorMessage != "" {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: errorMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},
}
