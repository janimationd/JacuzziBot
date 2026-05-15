package tama

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

const invalidEventIdMessage = "You must provide a valid ID for an existing playtime event this server."

// Cancel an existing scheduled playtime by ID.
var CancelTamaPlaytime = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "cancel-tama-playtime",
		Description: "Cancel an existing scheduled playtime by ID.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "id",
				Description: "The ID of the playtime event. You can see the IDs of all events with /list-tama-playtimes.",
				Required:    true,
			},
		},
		DefaultMemberPermissions: &manageGuildPerms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Extract options/parameters
		eventId := utils.GetCommandOption(interaction, "id").StringValue()
		serverId := interaction.GuildID

		// Check if this is being executed in the registered channel.
		registeredChannelId := db.GetTamaChannel(serverId)
		errorMessage := ""

		// Basic validations
		if registeredChannelId == "" {
			errorMessage = "No channel is registered as a Tama minigame yet (run `/register-tama-channel` first)."
		} else if registeredChannelId != interaction.ChannelID {
			errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
		}

		// Make sure the ID only applies to this server before trying to use it.
		if errorMessage == "" {
			matched, err := regexp.MatchString(fmt.Sprintf("TamaPlaytime-%s-.*", serverId), eventId)
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't validate event ID: %w", err)
			} else if !matched {
				errorMessage = invalidEventIdMessage
			}
		}

		// Cancel the event.
		if errorMessage == "" && !db.CancelEvent(eventId) {
			errorMessage = invalidEventIdMessage
		}

		var message string
		if errorMessage == "" {
			message = "Tama Playtime event cancelled."
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
