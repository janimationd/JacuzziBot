package tama

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
)

// Care for a Tama/egg
var CareTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "care-tama",
		Description: "Care for a Tama. Eggs get closer to hatching (24hr cooldown); Tamas get happier (8hr cooldown).",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama you want to care for.",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		intId := utils.GetCommandOption(interaction, "id").IntValue()
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
			if intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if !user.Tamas.Contains(id) {
				errorMessage = "You can only care for Tamas that you own."
			}
		}

		// Load the user's timezone
		var userTimezone *time.Location
		if errorMessage == "" {
			userTimezone, err = time.LoadLocation(user.Timezone)
			if err != nil {
				errorMessage = fmt.Sprintf("User's timezone %s is invalid: %s", user.Timezone, err.Error())
			}
		}

		// Attempt to care for this Tama
		var modified, hatched bool
		var tama *models.Tama
		if errorMessage == "" {
			tama, modified, hatched, err = db.CareForTama(serverId, id, userTimezone)
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
			var message string
			var flags discordgo.MessageFlags
			if hatched {
				message = fmt.Sprintf("Tama #%d has hatched! :tada: Congrats to <@%s>!\n\n", id, userId)
				message += fmt.Sprintf("You can now name it with `/name-tama %d` if you'd like.\n", id)
				message += tamas.GetTamaStatus(tama, userTimezone, "#")
			} else if tama.IsEgg() {
				careCountBeforeHatching := constants.EggCareHatchThreshold - tama.EggCareCount
				message = fmt.Sprintf("You have cared for egg #%d. Only %d days left before it hatches!",
					id, careCountBeforeHatching)
				flags = discordgo.MessageFlagsEphemeral
			} else if modified {
				message = fmt.Sprintf("You have played with Tama %s, improving its mood, and it's now %s!",
					tama.GetNameAndId(), tama.GetMoodString())
				flags = discordgo.MessageFlagsEphemeral
			} else {
				message = fmt.Sprintf("You have played with Tama %s, though its mood is already at maximum (+%d).",
					tama.GetNameAndId(), models.TamaMoodLimit)
				flags = discordgo.MessageFlagsEphemeral
			}
			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
					Flags:   flags,
				},
			})
		}
	},
}
