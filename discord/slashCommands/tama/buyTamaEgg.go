package tama

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows"
)

const commandName string = "BuyTamaEgg"
const BuyTamaEggConfirmPurchaseId string = commandName + ":ConfirmPurchase"
const BuyTamaEggCancelPurchaseId string = commandName + ":CancelPurchase"

var BuyTamaEgg = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "buy-tama-egg",
		Description: "Buy a Tama egg. Cost increases the more you buy; you'll be asked to confirm purchase cost.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var cost float64
		// Calculate the cost of buying another egg
		user, err := db.GetUser(interaction.GuildID, interaction.Member.User.ID)
		if err == nil {
			cost = workflows.CalculateTamaEggCost(&user)
		}

		var message string
		allowPurchase := false
		if err == nil {
			message = fmt.Sprintf("The cost to buy your next egg is %s point%s, doubling thereafter.",
				utils.FormatUIFloat(cost), utils.Plural(cost))
			if user.Points >= cost {
				allowPurchase = true
				message += fmt.Sprintf("\nYou currently have %s point%s, which is enough. Confirm purchase?",
					utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
			} else {
				message += fmt.Sprintf("\nYou currently have %s point%s, **which isn't enough**.",
					utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
			}
		} else {
			message = fmt.Sprintf("Failed to check purchase eligibility: %s", err.Error())
		}

		// Only add the confirm/cancel buttons if the user is allowed to complete the purchase.
		buttons := []discordgo.MessageComponent{}
		if allowPurchase {
			buttons = append(buttons, discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Confirm",
						Style:    discordgo.PrimaryButton,
						CustomID: BuyTamaEggConfirmPurchaseId,
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.DangerButton,
						CustomID: BuyTamaEggCancelPurchaseId,
					},
				},
			})
		}

		// Respond with feedback message
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:    message,
				Flags:      discordgo.MessageFlagsEphemeral,
				Components: buttons,
			},
		})

		if err != nil {
			log.Println("InteractionRespond error:", err)
		}
	},
}

func HandleBuyTamaEggConfirmPurchase(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	tama, err := workflows.BuyTamaEgg(interaction.GuildID, interaction.ChannelID, interaction.Member.User.ID)

	var message string
	if err == nil {
		message += "**Tama egg purchased!**\n"
		message += fmt.Sprintf("Its ID is: `%d`\n\n", tama.Id)
		message += fmt.Sprintf("Make sure to take care of it once a day using `/care-tama-egg %d`. ", tama.Id)
		message += fmt.Sprintf("You can check on the status of your egg with `/check-tama %d`.", tama.Id)
	} else {
		message = fmt.Sprintf("Couldn't complete egg purchase: %s", err.Error())
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    message,
			Components: []discordgo.MessageComponent{},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("Couldn't respond to egg purchase:", err)
	} else {
		log.Println("Tama egg purchase completed.")
	}
}

func HandleBuyTamaEggCancelPurchase(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "Cancelled.",
			Components: []discordgo.MessageComponent{},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
	if err == nil {
		log.Println("Tama egg purchase cancelled.")
	} else {
		log.Println("Failed to cleanup egg purchase message:", err.Error())
	}
}
