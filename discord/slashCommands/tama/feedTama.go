package tama

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Feed a Tama to reduce its hunger by 1.
var FeedTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name: "feed-tama",
		Description: fmt.Sprintf("Pay %s point%s to Feed a Tama, reducing its hunger by 1.",
			utils.FormatUIFloat(constants.TamaFeedCost), utils.Plural(constants.TamaFeedCost)),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama to feed.",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		idOption := utils.GetCommandOption(interaction, "id")
		intId := idOption.IntValue()
		id := models.JacuzziId(intId)

		userId := interaction.Member.User.ID
		serverId := interaction.GuildID

		// Check if this is being executed in the registered channel.
		registeredChannelId := db.GetTamaChannel(serverId)
		var errorMessage string

		// Basic validations
		user, err := db.GetUser(serverId, userId)
		if err != nil {
			errorMessage = "Couldn't load user details" + constants.ErrorReportMessageSuffix
		} else {
			if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if !user.Tamas.Contains(id) {
				errorMessage = "You can only feed Tamas that you own."
			}
		}

		// Load the tama's details
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(serverId, id)
			if err != nil {
				errorMessage = err.Error()
			}
		}

		// See if the Tama needs to be fed
		if errorMessage == "" {
			if tama.Hunger == 0 {
				errorMessage = "This Tama is already full. You don't need to feed it again until tomorrow (your local time)."
			}
		}

		const cost = constants.TamaFeedCost

		// Deduct the points to feed it, then feed it.
		if errorMessage == "" {
			user, err = db.ModifyUserPoints(serverId, userId, -cost)
			if err != nil {
				errorMessage = err.Error()
			} else {
				// Since now we have deducted points, we need to invert error handling to properly undo the point deduction if
				// something goes wrong.
				tama, err = db.FeedTama(serverId, id)
				if err != nil {
					log.Printf("Failed to feed Tama: %s\n", err.Error())
					// Refund the cost of the food
					var refundErr error
					user, refundErr = db.ModifyUserPoints(serverId, userId, cost)
					if refundErr != nil {
						log.Printf("And also failed to refund user points: %s\n", err.Error())
						errorMessage = fmt.Sprintf("Failed to feed Tama (%s), and also failed to refund your points (%s)%s",
							err.Error(), refundErr.Error(), constants.ErrorReportMessageSuffix)
					} else {
						errorMessage = fmt.Sprintf("Failed to feed Tama (%s)%s",
							err.Error(), constants.ErrorReportMessageSuffix)
					}
				}
			}
		}

		// Figure out response message
		var message string
		var flags discordgo.MessageFlags
		if errorMessage != "" {
			message = errorMessage
			flags = discordgo.MessageFlagsEphemeral
		} else {
			message = fmt.Sprintf("<@%s> has fed %s one food, purchased with %s point%s.\n\nNow they have %s point%s.",
				userId, tama.GetNameAndId(), utils.FormatUIFloat(cost), utils.Plural(cost),
				utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
		}

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   flags,
			},
		})
	},
}
