package tama

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

var perms = int64(discordgo.PermissionManageGuild)

func nextUTCTime(t time.Time, hour int, minute int) time.Time {
	candidate := time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, time.UTC)
	if !candidate.After(t) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}

// Schedule a UTC time each day during which the Tamas will play with each other.
var ScheduleTamaPlaytime = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "schedule-tama-playtime",
		Description: "Schedule a UTC time (hour:min) each day during which the Tamas will play with each other.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "hour",
				Description: "The UTC hour each day when the tamas will play with each other.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "minute",
				Description: "The UTC minute of the specified hour when the tamas will play with each other.",
				Required:    true,
			},
		},
		DefaultMemberPermissions: &perms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Extract options/parameters
		hour := utils.GetCommandOption(interaction, "hour").IntValue()
		minute := utils.GetCommandOption(interaction, "minute").IntValue()
		serverId := interaction.GuildID

		// Check if this is being executed in the registered channel.
		registeredChannelId := db.GetTamaChannel(serverId)
		errorMessage := ""
		var err error

		// Basic validations
		if hour < 0 || hour > 23 {
			errorMessage = "You must choose a valid UTC hour [0-23]."
		} else if minute < 0 || minute > 59 {
			errorMessage = "You must choose a valid UTC minute [0-59]."
		} else if registeredChannelId == "" {
			errorMessage = "No channel is registered as a Tama minigame yet (run `/register-tama-channel` first)."
		} else if registeredChannelId != interaction.ChannelID {
			errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
		}

		// Build event payload
		var payloadBytes []byte
		if errorMessage == "" {
			payload := models.TamaPlaytimePayload{
				ServerId: serverId,
			}
			payloadBytes, err = json.Marshal(payload)
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't marshall TamaPlaytimePayload to JSON: %s", err.Error())
			}
		}

		var eventId string

		// Schedule the recurring event
		if errorMessage == "" {
			// Events are non-server-specific, so to uniquely identify them we need to include the server ID.
			// We also use the hour and minute so that two events cannot be accidentally put onto the same time.
			eventId = fmt.Sprintf("TamaPlaytime-%s-%d:%d", serverId, hour, minute)
			nextTime := nextUTCTime(time.Now(), int(hour), int(minute))

			event := models.ScheduledEvent{
				ID:       eventId,
				NextTime: nextTime,
				// We explicitly ignore leap seconds, so as the years go on this may drift by a few seconds. Big whoop.
				Interval:            24 * time.Hour,
				Handler:             "TamaPlaytimeHandler",
				RestartGapTolerance: 1 * time.Hour,
				Payload:             payloadBytes,
			}

			// Overwrite any existing event at this time (true)
			_, err = db.ScheduleEvent(&event, true)
			if err != nil {
				errorMessage = fmt.Sprintf("Failed to register event %s: %s", eventId, err.Error())
			}
		}

		var message string
		if errorMessage == "" {
			message = fmt.Sprintf("Tama Playtime event scheduled with ID: `%s`", eventId)
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
