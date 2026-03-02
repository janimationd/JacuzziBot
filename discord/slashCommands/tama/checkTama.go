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

// Check the status of a Tama
var CheckTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "check-tama",
		Description: "Check the status of a Tama (or all your Tamas).",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama you want to check. Omit to check all your Tamas.",
				Required:    false,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		idOption := utils.GetCommandOption(interaction, "id")
		useSingleId := idOption != nil
		var intId int64
		var id models.JacuzziId
		if useSingleId {
			intId = idOption.IntValue()
			id = models.JacuzziId(intId)
		}
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
			if useSingleId && intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if !useSingleId && user.Tamas.Size() == 0 {
				errorMessage = "You don't own any Tamas."
			}
		}

		// Fetch Tama status(es)
		var message string
		if errorMessage == "" {
			if useSingleId {
				tama, err := db.GetTama(serverId, id)
				if err != nil {
					errorMessage = err.Error()
				} else {
					message = tama.StatusMessage()
				}
			} else {
				message = "Tamas you own:"
				for id := range user.Tamas.All() {
					tama, err := db.GetTama(serverId, id)
					if err != nil {
						// Swallow and log the error, continuing through the list of ones we can describe.
						log.Printf("Failed to load tama %d's details: %s\n", id, err.Error())
					} else {
						message += fmt.Sprintf("\n- %s", tama.StatusMessage())
					}
				}
			}
		}

		if errorMessage != "" {
			// Respond with error message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: errorMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},
}
