package slashCommands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows"
)

const DowngradeFlairConfirmId = "DowngradeFlair:Confirm"
const DowngradeFlairCancelId = "DowngradeFlair:Cancel"

// Uppgrade the user's flair level using points.
var DowngradeFlair = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "downgrade-flair",
		Description: "Downgrade your cosmetic flair level, getting back the points you used to purchase it.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var err error
		userId := interaction.Member.User.ID
		serverId := interaction.GuildID

		errorMessage := ""

		user, err := db.GetUser(serverId, userId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't load user details: %s\n", err.Error())
		}

		var flair models.FlairLevel
		var flairProps models.Flair
		var prevFlairProps models.Flair
		var pointDiff float64

		// Validations
		if errorMessage == "" {
			flair = user.Flair
			flairProps = models.FlairProps[flair]
			if flair == models.FlairNone {
				errorMessage = "You don't have any flair to downgrade"
			} else {
				prevFlairProps = models.FlairProps[flair-1]
				pointDiff = flairProps.TotalPointCost - prevFlairProps.TotalPointCost
			}
		}

		if errorMessage == "" {
			message := fmt.Sprintf("- Your current flair level is **%s** (spent %s point%s so far, color = %s%s)\n",
				flairProps.Name, utils.FormatUIFloat(flairProps.TotalPointCost),
				utils.Plural(flairProps.TotalPointCost), flairProps.ColorName, flairProps.ColorEmoji)
			message += fmt.Sprintf("- Previous flair level is **%s** (%s total point cost, color = %s%s)\n\n",
				prevFlairProps.Name, utils.FormatUIFloat(prevFlairProps.TotalPointCost), prevFlairProps.ColorName,
				prevFlairProps.ColorEmoji)

			message += fmt.Sprintf("Downgrading will return **%s point%s** to you.\n\nConfirm downgrade?",
				utils.FormatUIFloat(pointDiff), utils.Plural(pointDiff))

			confirmButtonCustomId := fmt.Sprintf("%s|%d", DowngradeFlairConfirmId, flair-1)

			// Embed current or target flair level into button custom ID
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
					Flags:   discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Confirm",
									Style:    discordgo.PrimaryButton,
									CustomID: confirmButtonCustomId,
								},
								discordgo.Button{
									Label:    "Cancel",
									Style:    discordgo.DangerButton,
									CustomID: DowngradeFlairCancelId,
								},
							},
						},
					},
				},
			})
		} else {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: errorMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},
}

func HandleDowngradeFlairConfirm(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	serverId := interaction.GuildID
	userId := interaction.Member.User.ID
	customId := interaction.MessageComponentData().CustomID
	targetFlairLevel := models.FlairLevel(models.ParseCustomIdToInt(customId))
	currentFlairLevel := targetFlairLevel + 1

	errorMessage := ""

	user, pointsDiff, err := db.ModifyUserFlairLevel(serverId, userId, currentFlairLevel, targetFlairLevel)
	if err != nil {
		errorMessage = fmt.Sprintf("Couldn't modify user flair level: %s", err.Error())
	}

	if errorMessage == "" {
		err = workflows.EnsureFlairRoles(session, serverId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't ensure flair roles existed: %s", err.Error())
		}
	}

	if errorMessage == "" {
		err = workflows.ChangeUserFlairRole(session, serverId, userId, targetFlairLevel)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't update user's flair role: %s", err.Error())
		}
	}

	// If we need to rollback the flair update
	if pointsDiff != 0 && errorMessage != "" {
		_, _, err = db.ModifyUserFlairLevel(serverId, userId, targetFlairLevel, currentFlairLevel)
		if err != nil {
			errorMessage += fmt.Sprintf(" Also failed to rollback the flair downgrade: %s", err.Error())
		}
	}

	var message string
	if errorMessage == "" {
		message = "Done!"
	} else {
		message = errorMessage
	}

	// Respond
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Couldn't send flair downgrade response: %s\n", err.Error())
	}

	if errorMessage == "" {
		targetFlair := models.FlairProps[targetFlairLevel]

		// Send channel message announcing the downgrade with details
		stars := ""
		for range targetFlairLevel {
			stars += ":star:"
		}
		announcement := fmt.Sprintf("# <@%s> has just downgraded their flair level! %s\n", userId, stars)
		announcement += fmt.Sprintf("Now flair level %d: **%s** (%s%s)\n\n",
			targetFlairLevel, targetFlair.Name, targetFlair.ColorName, targetFlair.ColorEmoji)
		announcement += fmt.Sprintf(
			"This returned **%s point%s** to them (%s still spent on flair). They now have %s.\n\n",
			utils.FormatUIFloat(pointsDiff), utils.Plural(pointsDiff), utils.FormatUIFloat(targetFlair.TotalPointCost),
			utils.FormatUIFloat(user.Points))
		announcement += "-# *Flair is a purely cosmetic role that changes the color of your name in the server sidebar. " +
			"You can modify your own using `/upgrade-flair` and `/downgrade-flair`.*"
		_, err = session.ChannelMessageSend(interaction.ChannelID, announcement)
		if err != nil {
			log.Printf("Couldn't send flair downgrade announcement: %s\n", err.Error())
		}
	}
}

func HandleDowngradeFlairCancel(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Cancelled.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err == nil {
		log.Println("Flair downgrade cancelled.")
	} else {
		log.Println("Failed to send cancel flair downgrade response:", err.Error())
	}
}
