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

// Claim an unclaimed Tama egg
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
			} else if user.Tamas.Size() >= constants.TamaLimitPerUser {
				errorMessage = fmt.Sprintf("You're already at the limit of how many Tamas you can own: %d.",
					constants.TamaLimitPerUser)
			}
		}

		// Check if an unclaimed egg with that ID exists.
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(serverId, id)
			if err != nil {
				errorMessage = err.Error()
			}
		}

		// Check if the user is allowed to claim the egg
		if errorMessage == "" {
			canClaim, reason := canUserClaimEgg(serverId, userId, tama)
			if !canClaim {
				errorMessage = reason
			}
		}

		// Claim the egg
		if errorMessage == "" {
			tama, err = db.ChangeTamaOwner(serverId, id, userId, false)
			if err != nil {
				errorMessage = err.Error()
			}
		}

		// Add ID to user's owned Tamas
		if errorMessage == "" {
			user, err = db.ModifyUserTamas(serverId, userId, db.Add, id)
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't claim egg: %s", err.Error())
				// Try to undo setting the egg's owner
				tama, err = db.ChangeTamaOwner(serverId, id, "", true)
				if err != nil {
					errorMessage += fmt.Sprintf(", also couldn't revert claim: %s", err.Error())
				}
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
