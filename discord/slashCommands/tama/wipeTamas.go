package tama

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// DANGEROUS!!! Wipe the Tamas minigame state back to nothing.
var WipeTamas = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "wipe-tamas",
		Description: "DANGEROUS!!! Wipe the Tamas minigame state back to nothing.",
		Options: []*discordgo.ApplicationCommandOption{
			{
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
		confirm := utils.GetCommandOption(interaction, "confirm").StringValue()

		var message string
		flags := discordgo.MessageFlagsEphemeral
		if confirm == "CONFIRM" {
			err := db.WipeTamaBuckets(serverId)
			if err != nil {
				message = fmt.Sprintf("Couldn't wipe Tamas minigame: %s", err.Error())
			} else {
				message = fmt.Sprintf("Tamas minigame state was completely wiped by <@%s>!", userId)
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
