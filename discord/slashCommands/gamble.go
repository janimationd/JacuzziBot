package slashCommands

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Users can gamble their points away.
var Gamble = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "gamble",
		Description: "Wager some of your points, gambling to maybe earn more!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "chance",
				Description: "What is the chance of winning? (e.g. \"3\" for 1 in 3 chance to make back triple your wager)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "wager",
				Description: "How many points are you wagering?",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var err error
		userId := interaction.Member.User.ID

		// Extract options/parameters
		chance := utils.GetCommandOption(interaction, "chance").IntValue()
		wager := utils.GetCommandOption(interaction, "wager").FloatValue()

		// Validations

		// Amount is invalid
		if wager <= 0 {
			err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your wager must be more than 0.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Println("Failed to respond to interaction:", err)
			}
			return
		}

		// User chose an invalid chance. Chance of 1 in 1 doesn't make sense, so must be at least 2.
		if chance < 2 {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Chance must be at least 2.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Println("Failed to respond to interaction:", err)
			}
			return
		}

		// Subtract their wager now to make sure they have the required amount of points.
		user, err := db.ModifyUserPoints(interaction.GuildID, userId, -wager)
		if err != nil {
			// The error will have a user-facing message in the case where they don't have enough points.
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to deduct wager from your balance:\n\n" + err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		message := fmt.Sprintf("<@%s> wagered %s point%s with a 1 in %d chance of winning...",
			userId, utils.FormatUIFloat(wager), utils.Plural(wager), chance)

		// Determine win or loss. They have a 1 in N chance to actually win.
		win := rand.Intn(int(chance)) == 0

		if win {
			winnings := wager * float64(chance)
			net := winnings - wager
			message += fmt.Sprintf("\n\n**And won %s point%s back** (+%s point%s net) :money_mouth:",
				utils.FormatUIFloat(winnings), utils.Plural(winnings), utils.FormatUIFloat(net), utils.Plural(net))

			// Give them their winnings
			user, err = db.ModifyUserPoints(interaction.GuildID, userId, winnings)
			if err != nil {
				log.Printf("Failed to modify points for %s (%s): %s\n",
					userId, interaction.Member.User.DisplayName(), err.Error())
				message += fmt.Sprintf("\n\nThough we failed to award your points: %s", err.Error())

				// Try to refund their wager
				user, err = db.ModifyUserPoints(interaction.GuildID, userId, wager)
				if err != nil {
					log.Println("Also failed to refund their wager:", err)
					message += fmt.Sprintf("\n\nWe also failed to refund your wager%s: %s",
						constants.ErrorReportMessageSuffix, err.Error())
				} else {
					message += "\n\nWe have refunded their wager."
				}
			}
		} else {
			message += "\n\n**And lost the gamble** :frowning2:"
		}

		// If there was an error, privately show the feedback message to the user.
		var flags discordgo.MessageFlags
		if err != nil {
			flags = discordgo.MessageFlagsEphemeral
		}

		message += fmt.Sprintf(" They now have %s point%s.",
			utils.FormatUIFloat(user.Points), utils.Plural(user.Points))

		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   flags,
			},
		})
	},
}
