package tama

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Name a Tama
var NameTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "name-tama",
		Description: "Name a Tama. Tamas can only be named after hatching.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama you want to name.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "new-name",
				Description: "The new name you want to give to the Tama.",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		intId := utils.GetCommandOption(interaction, "id").IntValue()
		id := models.JacuzziId(intId)
		newName := utils.GetCommandOption(interaction, "new-name").StringValue()
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
			if intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if utf8.RuneCountInString(newName) > constants.MaxTamaNameLength {
				errorMessage = fmt.Sprintf("Tama names cannot exceed %d characters. Your name uses %d.",
					constants.MaxTamaNameLength, utf8.RuneCountInString(newName))
			} else {
				ownedTamas, err := db.GetAllTamas(serverId, userId, false, false)
				if err != nil {
					errorMessage = fmt.Sprintf("Couldn't fetch all your owned Tamas: %s", err.Error())
				} else if ownedTamas[id] == nil {
					errorMessage = "You can only name Tamas that you own."
				}
			}
		}

		// Make sure the name is not just a number
		if errorMessage == "" {
			_, err := strconv.ParseFloat(newName, 64)
			if err == nil {
				errorMessage = "A Tama name cannot be purely a number, so we can easily distinguish it from an ID."
			}
		}

		// Check if the Tama has already hatched.
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(serverId, id)
			if err != nil {
				errorMessage = err.Error()
			}
		}

		if errorMessage == "" && tama.IsEgg() {
			errorMessage = "You cannot name Tamas before they hatch."
		}

		// Rename the Tama
		if errorMessage == "" {
			tama, err = db.NameTama(serverId, id, newName)
			if err != nil {
				errorMessage = err.Error()
			}
		}

		if errorMessage != "" {
			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: errorMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			message := fmt.Sprintf("<@%s> has just named Tama #%d \"%s\"!", userId, id, newName)
			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
				},
			})
		}
	},
}
