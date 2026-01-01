package slashCommands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Print the help string
var Help = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "help",
		Description: "Display supported features and commands.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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
