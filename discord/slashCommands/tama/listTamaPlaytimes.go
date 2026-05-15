package tama

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

// Schedule a UTC time each day during which the Tamas will play with each other.
var ListTamaPlaytimes = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:                     "list-tama-playtimes",
		Description:              "List all recurring Tama playtimes.",
		Options:                  []*discordgo.ApplicationCommandOption{},
		DefaultMemberPermissions: &manageGuildPerms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverId := interaction.GuildID

		// Check if this is being executed in the registered channel.
		registeredChannelId := db.GetTamaChannel(serverId)
		errorMessage := ""
		var err error

		// Basic validations
		if registeredChannelId == "" {
			errorMessage = "No channel is registered as a Tama minigame yet (run `/register-tama-channel` first)."
		} else if registeredChannelId != interaction.ChannelID {
			errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
		}

		var playtimes []*models.ScheduledEvent
		if errorMessage == "" {
			playtimes, err = db.GetAllEvents("TamaPlaytime-" + serverId)
			if err != nil {
				errorMessage = fmt.Sprintf("Failed to fetch all playtime events from the DB: %s\n", err.Error())
			}
		}

		var message string
		if errorMessage == "" {
			if len(playtimes) > 0 {
				message = "All Tama playtime event IDs:"
				for _, event := range playtimes {
					message += fmt.Sprintf("\n- `%s`", event.ID)
				}
			} else {
				message = "No Tama playtime events are scheduled."
			}
		} else {
			message = errorMessage
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
