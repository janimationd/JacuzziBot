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
var Help = models.SlashCommand{
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

		var err error

		// Check if this is being executed in the registered channel.
		channelId := db.GetTamaChannel(interaction.GuildID)
		var errorMessage string
		if channelId != interaction.ChannelID {
			errorMessage = fmt.Sprintf("You must execute this command in <#%s>.", channelId)
		}

		// Check if the user is at their egg limit.
		var user models.User
		if errorMessage == "" {
			user, err = db.GetUser(interaction.GuildID, interaction.Member.User.ID)
			errorMessage = "Couldn't load user details" + constants.ErrorReportMessageSuffix
		}
		if errorMessage == "" && user.Tamas.Size() >= constants.TamaLimitPerUser {
			errorMessage = fmt.Sprintf("You're already at the limit of how many Tamas you can own: %d.",
				constants.TamaLimitPerUser)
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

		// Check if the user is allowed to claim the egg.
		if errorMessage == "" {
			canClaim, reason := canUserClaimEgg(interaction.GuildID, interaction.Member.User.ID, tama)
			if !canClaim {
				errorMessage = reason
			}
		}

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: utils.Help(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
