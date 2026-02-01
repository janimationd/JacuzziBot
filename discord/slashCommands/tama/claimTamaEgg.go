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

// Returns whether or not the user can claim the egg, along with a reason if not.
func canUserClaimEgg(serverId string, userId string, tama *models.Tama) (bool, string) {
	eggAge := time.Now().Unix() - tama.EggLaidTime
	// Only owners of the parent Tamas can claim within this window
	if eggAge < constants.OnlyParentOwnersCanClaimSeconds {
		for parentId := range tama.Parents.All() {
			parent, err := db.GetTama(serverId, parentId)
			if err != nil {
				log.Printf("Failed to fetch parent info for Tama %d: %s\n", parentId, err.Error())
				continue
			}
			if parent.Owner == userId {
				return true, ""
			}
		}
		hoursRemaining := float64(constants.OnlyParentOwnersCanClaimSeconds-eggAge) / 60 / 60
		return false, fmt.Sprintf("Cannot claim egg yet. "+
			"For the next %.2f hour%s only the owners of the egg's parent Tamas can claim it.",
			hoursRemaining,
			utils.Plural(hoursRemaining),
		)
	} else {
		return true, ""
	}
}

// Print the help string
var ClaimTamaEgg = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "claim-tama-egg",
		Description: "Claim an unclaimed Tama egg.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the egg you want to claim.",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		id := models.JacuzziId(utils.GetCommandOption(interaction, "id").IntValue())
		userId := interaction.Member.User.ID

		var err error

		// Check if this is being executed in the registered channel.
		channelId := db.GetTamaChannel(interaction.GuildID)
		var errorMessage string

		// Basic validations
		var user models.User
		user, err = db.GetUser(interaction.GuildID, userId)
		if err != nil {
			errorMessage = "Couldn't load user details" + constants.ErrorReportMessageSuffix
		} else {
			if id <= models.NoId {
				errorMessage = "Tama IDs must be greater than 0."
			} else if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if channelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if channelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", channelId)
			} else if user.Tamas.Size() >= constants.TamaLimitPerUser {
				errorMessage = fmt.Sprintf("You're already at the limit of how many Tamas you can own: %d.",
					constants.TamaLimitPerUser)
			}
		}

		// Check if an unclaimed egg with that ID exists.
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(interaction.GuildID, id)
			if err != nil {
				errorMessage = err.Error()
			}
		}
		if errorMessage == "" {
			if tama.IsOwned() {
				errorMessage = fmt.Sprintf("This Tama is already owned by <@%s>.", tama.Owner)
			}
		}

		// Check if the user is allowed to claim the egg
		if errorMessage == "" {
			canClaim, reason := canUserClaimEgg(interaction.GuildID, userId, tama)
			if !canClaim {
				errorMessage = reason
			}
		}

		// Claim the egg
		if errorMessage == "" {
			tama.Owner = userId
			err = db.StoreTama(interaction.GuildID, channelId, tama)
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't claim egg: %s", err.Error())
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
			message := fmt.Sprintf("<@%s> has just claimed Tama egg #%d!", userId, id)
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
