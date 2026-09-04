package slashCommands

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Roll an X-sided die.
var Roll = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "roll",
		Description: "Roll X Y-sided dice.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "sides",
				Description: "How many sides does each die have?",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "How many dice are you rolling? Defaults to 1 if not provided.",
				Required:    false,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Extract options/parameters
		sides64 := utils.GetCommandOption(interaction, "sides").IntValue()
		countOpt := utils.GetCommandOption(interaction, "count")
		count := 1
		if countOpt != nil {
			count = int(countOpt.IntValue())
		}

		// Validations

		errorMessage := ""
		var sideLimit int64 = 1000
		countLimit := 10000

		if sides64 < 2 || sides64 > sideLimit {
			errorMessage = fmt.Sprintf("Sides must be between 2 and %d (inclusive).", sideLimit)
		} else if count < 1 || count > countLimit {
			errorMessage = fmt.Sprintf("Count must be between 1 and %d (inclusive).", countLimit)
		}
		sides := int(sides64)

		// If there was an error, privately show the feedback message to the user.
		var flags discordgo.MessageFlags
		var message string

		// Make the roll(s)
		if errorMessage == "" {
			if count == 1 {
				message = fmt.Sprintf("Rolling a %d-sided die...\n\n", sides)
			} else {
				message = fmt.Sprintf("Rolling %d %d-sided dice...\n\n", count, sides)
			}

			rolls := make([]int, count)
			for i := range count {
				rolls[i] = rand.Intn(sides) + 1
			}

			if count == 1 {
				message += fmt.Sprintf("**Rolled a `%d`!**", rolls[0])
			} else {
				message += "Rolled:\n```\n"
				total := 0
				displayLimit := 100
				for i := range count {
					total += rolls[i]
					// Only show the first 100 rolls so the message doesn't get too large.
					if i < displayLimit {
						if i < count-1 {
							message += fmt.Sprintf("%d + ", rolls[i])
						} else {
							message += fmt.Sprintf("%d\n", rolls[i])
						}
					}
				}
				if count > displayLimit {
					message += fmt.Sprintf("... [only showing first %d rolls]", displayLimit)
				}
				message += fmt.Sprintf("```\n**Total = `%d`!**", total)
			}
		} else {
			message = errorMessage
			flags = discordgo.MessageFlagsEphemeral
		}

		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   flags,
			},
		})
	},
}
