package tama

import (
	"fmt"
	"log"
	"time"

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
			} else {
				ownedTamas, err := db.GetAllTamas(serverId, userId, false, false)
				if err != nil {
					errorMessage = fmt.Sprintf("Couldn't fetch all your owned Tamas: %s", err.Error())
				} else if ownedTamas[id] == nil {
					errorMessage = "You can only feed Tamas that you own."
				}
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

		// If the tama is an egg
		if errorMessage == "" && tama.IsEgg() {
			errorMessage = "This Tama is still an egg. You can only feed it after it hatches."
		}

		var timezone *time.Location
		if errorMessage == "" {
			timezone, err = time.LoadLocation(user.Timezone)
			if err != nil {
				panic(fmt.Sprintf("Existing user timezone is invalid: %s", user.Timezone))
			}
		}

		// See if the Tama needs to be fed
		if errorMessage == "" && tama.Hunger(timezone) == 0 {
			errorMessage = "This Tama is already full. You don't need to feed it again until tomorrow (your local time)."
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
				tama, err = db.FeedTama(serverId, id, timezone)
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
		if errorMessage != "" {
			message = errorMessage
		} else {
			message = fmt.Sprintf("You have fed %s one food, ", tama.GetNameAndId())
			newHunger := tama.Hunger(timezone)
			if newHunger > 0 {
				message += fmt.Sprintf("and it's still hungry! Feed it %d more time%s to satisfy its hunger.",
					newHunger, utils.Plural(newHunger))
			} else {
				message += "and now it's full!"
			}
			message += fmt.Sprintf("\n\nThe food cost %s point%s, so you now have %s point%s.",
				utils.FormatUIFloat(cost), utils.Plural(cost),
				utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
		}

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
