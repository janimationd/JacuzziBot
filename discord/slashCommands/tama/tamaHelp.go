package tama

import (
	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Print the help string for the Tama minigame
var TamaHelp = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "tama-help",
		Description: "Display supported features and commands for the Tama minigame.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: utils.TamaHelp(),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
