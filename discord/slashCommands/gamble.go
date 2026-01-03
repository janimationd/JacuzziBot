package slashCommands

import (
	"fmt"
	"log"
	"math"
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
		Description: "Gamble some of your points away",
		// If you add new options there will likely be some handler validations you need to tweak
		// (e.g. checking to make sure the count of options is as expected)
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "odds",
				Description: "What are the odds of winning? (e.g. 2 for 1 in 2 chance, yields double your points)",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "How many points are you gambling?",
				Required:    true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var err error

		// Extract options/parameters
		odds := getCommandOption(interaction, "odds").IntValue()
		amount := getCommandOption(interaction, "amount").IntValue()

		// Validations		
		
		// Amount is invalid
		if amount <= 0 {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can't gamble nothing!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// User picks invalid odds
		if odds <= 0 {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You can't gamble with invalid odds! Pick a number greater than 0.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// User doesn't have enough points to gamble
		user, err := db.GetUser(interaction.GuildID, interaction.Member.User.ID)

		if err != nil {
			log.Println("Unable to fetch user details:", err)
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unable to fetch user details.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		if user.Points < float64(amount) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You don't have enough points to gamble that much!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Determine win or lose
		// If user picks 1:n odds, they have a 1 in n chance to actually win.
		win := rand.Intn(int(odds)) == 0
		pointsDelta := amount * odds
		if !win {
			pointsDelta = -amount
		} 

		var userNewState models.User
		userNewState, err = db.ModifyUserPoints(interaction.GuildID, interaction.Member.User.ID, float64(pointsDelta))
	
		if err != nil {
			str := fmt.Sprintf(
				"Failed to modify points for %s (%s)",
				interaction.Member.User.ID,
				interaction.Member.User.DisplayName(),
			)
			log.Println(str+":", err)
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: str + constants.ErrorReportMessageSuffix,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Respond with the outcome
		outcomeStr := "lost"
		if win {
			outcomeStr = "won"
		}

		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf(
					"You %s %s points! You now have %s points.",
					outcomeStr,
					utils.FormatUIFloat(math.Abs(float64(pointsDelta))),
					utils.FormatUIFloat(userNewState.Points),
				),
			},
		})
	},
}
