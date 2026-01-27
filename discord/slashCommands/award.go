package slashCommands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

var perms = int64(discordgo.PermissionManageGuild)

// Admins can award users points.
var Award = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "award",
		Description: "Award a user some points",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "recipient",
				Description: "Who will receive the points?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "amount",
				Description: "How many are you awarding?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Reason for awarding the points.",
				Required:    true,
			},
		},
		DefaultMemberPermissions: &perms,
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Extract options/parameters
		recipient := getCommandOption(interaction, "recipient").UserValue(nil)
		amount := getCommandOption(interaction, "amount").FloatValue()
		reasonStr := getCommandOption(interaction, "reason").StringValue()

		// We explicitly allow negative amounts so this can be used to undo incorrectly awarded points too.

		// Default feedback messaging, assuming everything goes well.
		var recipientName string
		if recipient.ID == session.State.User.ID {
			// recipient.DisplayName() is apparently an empty string when the targeted recipient is the bot itself,
			// so instead get its display name another way.
			recipientName = session.State.User.DisplayName()
		} else {
			member, err := session.GuildMember(interaction.GuildID, recipient.ID)
			if err != nil {
				log.Println("Unable to fetch member details:", err)
				recipientName = constants.ErrorFetchingDisplayNameMessage
			} else {
				recipientName = member.DisplayName()
			}
		}

		// Modify database. Error handling is complicated since we want to gracefully fail when possible.
		user, err := db.ModifyUserPoints(interaction.GuildID, recipient.ID, amount)
		if err != nil {
			str := fmt.Sprintf(
				"Failed to grant awarded points to %s (%s)",
				recipient.ID,
				recipientName,
			)
			log.Println(str+":", err)
			responseMessage := str + constants.ErrorReportMessageSuffix
			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: responseMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		responseMessage := fmt.Sprintf(
			"<@%s> has been awarded **%s point%s**!",
			recipient.ID,
			utils.FormatUIFloat(amount),
			utils.Plural(amount),
		)
		responseMessage += fmt.Sprintf(
			" Reason: \"%s\".\n",
			reasonStr,
		)
		responseMessage += fmt.Sprintf(
			"\nThey now have %s point%s.\n",
			utils.FormatUIFloat(user.Points),
			utils.Plural(user.Points),
		)

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseMessage,
			},
		})
	},
}
