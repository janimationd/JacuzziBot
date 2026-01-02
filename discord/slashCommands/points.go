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

// Users can check their own points or the points of another user
var Points = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "points",
		Description: "Check the points of a user",
		// If you add new options there will likely be some handler validations you need to tweak
		// (e.g. checking to make sure the count of options is as expected)
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Whom to check the points of? Omit to check your own points.",
				Required:    false,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Extract options/parameters
		user := getCommandOption(interaction, "user")

		// Figure out target
		var target *discordgo.User
		if user != nil {
			target = user.UserValue(session)
		} else if len(interaction.ApplicationCommandData().Options) == 0 {
			// If user is omitted, default to the user who initated the command.
			target = interaction.Member.User
		}

		// Figure out the display name of the target user
		var targetName string
		var verb string
		switch target.ID {
		case session.State.User.ID:
			// recipient.DisplayName() is apparently an empty string when the targeted user is the bot itself,
			// so instead get its display name another way.
			targetName = session.State.User.DisplayName()
			verb = "has"
		case interaction.Member.User.ID:
			// The user is requesting their own point amount
			targetName = "You"
			verb = "have"
		default:
			// Another user
			target, err := session.GuildMember(interaction.GuildID, target.ID)
			if err != nil {
				log.Println("Failed to fetch user details", err)
				targetName = constants.ErrorFetchingDisplayNameMessage
			} else {
				targetName = target.DisplayName()
			}
			verb = "has"
		}

		// Fetch current points from database.
		targetDbUser, err := db.GetUser(interaction.GuildID, target.ID)
		if err != nil {
			str := fmt.Sprintf("Failed to fetch current points for user %s (%s)", target.ID, targetName)
			log.Printf("%s: %s\n", str, err.Error())
			// The user malformed their command; reject it
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: str + constants.ErrorReportMessageSuffix,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Respond with feedback message
		plural := utils.Plural(targetDbUser.Points)
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s %s %s point%s.", targetName, verb, utils.FormatUIFloat(targetDbUser.Points), plural),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
