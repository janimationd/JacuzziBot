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

const UpgradeFlairConfirmId = "UpgradeFlair:Confirm"
const UpgradeFlairCancelId = "UpgradeFlair:Cancel"

// Uppgrade the user's flair level using points.
var UpgradeFlair = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "upgrade-flair",
		Description: "Upgrade your cosmetic flair level, changing the color of your name in the server sidebar.",
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
		var nextFlairProps models.Flair
		var pointDiff float64

		// Validations
		if errorMessage == "" {
			flair = user.Flair
			flairProps = models.FlairProps[flair]
			if flair == models.FlairMax-1 {
				errorMessage = fmt.Sprintf("You are already at maximum flair level: %s", flairProps.Name)
			} else {
				nextFlairProps = models.FlairProps[flair+1]
				pointDiff = nextFlairProps.TotalPointCost - flairProps.TotalPointCost
				if pointDiff > user.Points {
					errorMessage = fmt.Sprintf("You need %s point%s to upgrade your flair, but you only have %s!",
						utils.FormatUIFloat(pointDiff), utils.Plural(pointDiff), utils.FormatUIFloat(user.Points))
				}
			}
		}

		if errorMessage == "" {
			message := fmt.Sprintf("- Your current flair level is **%s** (spent %s point%s so far, color = %s%s)\n",
				flairProps.Name, utils.FormatUIFloat(flairProps.TotalPointCost),
				utils.Plural(flairProps.TotalPointCost), flairProps.ColorName, flairProps.ColorEmoji)
			message += fmt.Sprintf("- Next flair level is **%s** (%s total point cost, color = %s%s)\n\n",
				nextFlairProps.Name, utils.FormatUIFloat(nextFlairProps.TotalPointCost), nextFlairProps.ColorName,
				nextFlairProps.ColorEmoji)

			message += fmt.Sprintf("Upgrading will cost you **%s point%s**.\n\nConfirm upgrade?",
				utils.FormatUIFloat(pointDiff), utils.Plural(pointDiff))

			confirmButtonCustomId := fmt.Sprintf("%s|%d", UpgradeFlairConfirmId, flair+1)

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
									CustomID: UpgradeFlairCancelId,
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

func HandleUpgradeFlairConfirm(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	serverId := interaction.GuildID
	userId := interaction.Member.User.ID
	customId := interaction.MessageComponentData().CustomID
	targetFlairLevel := models.FlairLevel(models.ParseCustomIdToInt(customId))
	currentFlairLevel := targetFlairLevel - 1

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
			errorMessage += fmt.Sprintf(" Also failed to rollback the flair upgrade: %s", err.Error())
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
		log.Printf("Couldn't send flair upgrade response: %s\n", err.Error())
	}

	if errorMessage == "" {
		targetFlair := models.FlairProps[targetFlairLevel]

		// Send channel message announcing the upgrade with details
		stars := ""
		for range targetFlairLevel {
			stars += ":star:"
		}
		announcement := fmt.Sprintf("# <@%s> has just upgraded their flair level! %s\n", userId, stars)
		announcement += fmt.Sprintf("Now flair level %d: **%s** (%s %s)\n\n",
			targetFlairLevel, targetFlair.Name, targetFlair.ColorName, targetFlair.ColorEmoji)
		announcement += fmt.Sprintf(
			"This cost them **%s point%s** (for a total of %s spent on flair). They have %s left.\n\n",
			utils.FormatUIFloat(-pointsDiff), utils.Plural(-pointsDiff), utils.FormatUIFloat(targetFlair.TotalPointCost),
			utils.FormatUIFloat(user.Points))
		announcement += "-# *Flair is a purely cosmetic role that changes the color of your name in the server sidebar. " +
			"You can modify your own using `/upgrade-flair` and `/downgrade-flair`.*"
		_, err = session.ChannelMessageSend(interaction.ChannelID, announcement)
		if err != nil {
			log.Printf("Couldn't send flair upgrade announcement: %s\n", err.Error())
		}
	}
}

func HandleUpgradeFlairCancel(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Cancelled.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err == nil {
		log.Println("Flair upgrade cancelled.")
	} else {
		log.Println("Failed to send cancel flair upgrade response:", err.Error())
	}
}
