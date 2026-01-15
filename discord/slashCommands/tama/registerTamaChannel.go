package tama

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

// Only allow users with "Manage Guild/Server" permission to run this command.
var perm = int64(discordgo.PermissionManageGuild)

var RegisterTamaChannel = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name: "register-tama-channel",
		Description: "Registers a channel as a Tama minigame. " +
			"Can only be run by someone with 'Manage Server' permissions.",
		DefaultMemberPermissions: &perm,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		err := db.RegisterTamaChannel(interaction.GuildID, interaction.ChannelID)

		var message string
		if err == nil {
			message = fmt.Sprintf("Successfully registered <#%s> as a Tama minigame channel.", interaction.ChannelID)
		} else {
			message = fmt.Sprintf("Failed to register <#%s> as a Tama minigame channel: %s",
				interaction.ChannelID, err.Error())
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
