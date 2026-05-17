package tama

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
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
			if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if useSingleId && intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if !useSingleId {
				tamas, err := db.GetAllTamas(serverId, userId, false, false)
				if err != nil {
					errorMessage = fmt.Sprintf("Couldn't fetch list of Tamas owned by you: %s", err.Error())
				} else if len(tamas) == 0 {
					errorMessage = "You don't own any Tamas."
				}
			}
		}

		// Load the user's timezone
		var timezone *time.Location = nil
		if errorMessage == "" {
			timezone, err = time.LoadLocation(user.Timezone)
			if err != nil {
				errorMessage = fmt.Sprintf("User's timezone %s is invalid: %s", user.Timezone, err.Error())
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
					message = tamas.GetTamaStatus(tama, timezone, "#")
				}
			} else {
				message = "# All Tamas you own:\n"
				ownedTamas, err := db.GetAllTamas(serverId, userId, false, false)
				if err != nil {
					errorMessage = fmt.Sprintf("Couldn't fetch all your owned Tamas: %s", err.Error())
				} else {
					// Sort tamas by ID
					sortedTamas := make([]*models.Tama, len(ownedTamas))
					i := 0
					for _, tama := range ownedTamas {
						sortedTamas[i] = tama
						i++
					}
					slices.SortFunc(sortedTamas, func(a, b *models.Tama) int {
						return cmp.Compare(a.Id, b.Id)
					})
					for _, tama := range sortedTamas {
						if tama.IsAlive() { // Skip any dead Tamas
							message += tamas.GetTamaStatus(tama, timezone, "##")
						}
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
