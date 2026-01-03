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

// Users can give each other their points.
var Give = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "give",
		Description: "Give another user some of your points",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "recipient",
				Description: "Who will receive your points?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "amount",
				Description: "How many are you giving?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Send an optional message along with the points.",
				Required:    false,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		success := true

		// Extract options/parameters
		recipient := getCommandOption(interaction, "recipient").UserValue(nil)
		amount := getCommandOption(interaction, "amount").FloatValue()
		message := getCommandOption(interaction, "message")
		var messageStr string
		if message != nil {
			messageStr = message.StringValue()
		}

		// Validations
		if recipient.ID == interaction.Member.User.ID {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can't give yourself points, silly!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		if amount <= 0 {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can only give a value greater than zero, silly!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Default feedback messaging, assuming everything goes well.
		var responseMessage string
		var recipientName string
		if recipient.ID == session.State.User.ID {
			responseMessage = "I am honored, thank you!"
			// recipient.DisplayName() is apparently an empty string when the targeted recipient is the bot itself,
			// so instead get its display name another way.
			recipientName = session.State.User.DisplayName()
		} else {
			responseMessage = "Done!"
			member, err := session.GuildMember(interaction.GuildID, recipient.ID)
			if err != nil {
				log.Println("Unable to fetch member details:", err)
				recipientName = constants.ErrorFetchingDisplayNameMessage
			} else {
				recipientName = member.DisplayName()
			}
		}

		// Modify database. First try to subtract the user's own points, then add them to the recipient's.
		// Error handling is complicated since we want to gracefully fail when possible.
		var userAfter, recipientAfter models.User
		var err error
		userAfter, err = db.ModifyUserPoints(interaction.GuildID, interaction.Member.User.ID, -amount)
		if err != nil {
			responseMessage = err.Error()
			success = false
		} else {
			recipientAfter, err = db.ModifyUserPoints(interaction.GuildID, recipient.ID, amount)
			if err != nil {
				str := fmt.Sprintf(
					"Failed to grant given points to %s (%s)",
					recipient.ID,
					recipientName,
				)
				log.Println(str+":", err)
				responseMessage = str
				success = false
				// Refund the subtracted points since the give failed
				userAfter, err = db.ModifyUserPoints(interaction.GuildID, interaction.Member.User.ID, amount)
				if err != nil {
					str := fmt.Sprintf(
						"Also failed to refund points to %s (%s)",
						interaction.Member.User.ID,
						interaction.Member.User.DisplayName(),
					)
					log.Println(str+":", err)
					responseMessage += ". " + str
				}
				responseMessage += constants.ErrorReportMessageSuffix
			}
		}

		// Respond with feedback message
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if success {
			plural := utils.Plural(amount)
			line := fmt.Sprintf(
				"<@%s> has just given <@%s> %s point%s!\n",
				interaction.Member.User.ID,
				recipient.ID,
				utils.FormatUIFloat(amount),
				plural,
			)
			if messageStr != "" {
				line += fmt.Sprintf(
					"Message: \"%s\"\n",
					messageStr,
				)
			}
			plural = utils.Plural(userAfter.Points)
			line += fmt.Sprintf(
				"> *%s now has %s point%s*\n",
				interaction.Member.DisplayName(),
				utils.FormatUIFloat(userAfter.Points),
				plural,
			)
			plural = utils.Plural(recipientAfter.Points)
			line += fmt.Sprintf(
				"> *%s now has %s point%s*\n",
				recipientName,
				utils.FormatUIFloat(recipientAfter.Points),
				plural,
			)

			// Notify in the channel about what happened
			session.ChannelMessageSend(interaction.ChannelID, line)
		}
	},
}
