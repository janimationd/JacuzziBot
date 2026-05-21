package tama

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// DANGEROUS!!! Restore the Tamas minigame from a previous file backup.
var RestoreTamas = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "restore-tamas",
		Description: "DANGEROUS!!! Restore the Tamas minigame from a previous file backup.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "timestamp",
				Description: "The time that the backup was made at.",
				Required:    true,
			}, {
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "confirm",
				Description: "You must type CONFIRM here.",
				Required:    true,
			},
		},
		DefaultMemberPermissions: &manageGuildPerms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverId := interaction.GuildID
		userId := interaction.Member.User.ID
		timestamp := utils.GetCommandOption(interaction, "timestamp").StringValue()
		confirm := utils.GetCommandOption(interaction, "confirm").StringValue()

		var message string
		flags := discordgo.MessageFlagsEphemeral
		if confirm == "CONFIRM" {
			err := db.RestoreTamaBucketsFromBackup(serverId, timestamp)
			if err != nil {
				message = fmt.Sprintf("Couldn't restore Tamas minigame from backup: %s", err.Error())
			} else {
				message = fmt.Sprintf("Tamas minigame restored from backup `%s` by <@%s>!", timestamp, userId)
				flags = 0
			}
		} else {
			message = "You didn't `CONFIRM` the command, so assuming it was an accident."
		}

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   flags,
			},
		})
	},
}
