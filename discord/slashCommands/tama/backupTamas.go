package tama

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

// Backup the current state of the Tamas minigame to a file.
var BackupTamas = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:                     "backup-tamas",
		Description:              "Backup the current state of the Tamas minigame to a file.",
		DefaultMemberPermissions: &manageGuildPerms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverId := interaction.GuildID

		backupTime, err := db.BackupTamaBuckets(serverId)
		var message string
		if err != nil {
			message = fmt.Sprintf("Couldn't backup Tamas minigame: %s", err.Error())
		} else {
			message = fmt.Sprintf("Backup created at time: `%s`", backupTime)
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
