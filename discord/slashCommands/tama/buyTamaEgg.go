package tama

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
)

const buyTamaEggCommandName string = "BuyTamaEgg"
const BuyTamaEggConfirmPurchaseId string = buyTamaEggCommandName + ":ConfirmPurchase"
const BuyTamaEggCancelPurchaseId string = buyTamaEggCommandName + ":CancelPurchase"

var BuyTamaEgg = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "buy-tama-egg",
		Description: "Buy a Tama egg. Cost increases the more you buy; you'll be asked to confirm purchase cost.",
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		const cost float64 = constants.TamaEggPurchaseCost
		user, err := db.GetUser(interaction.GuildID, interaction.Member.User.ID)

		var message string
		allowPurchase := false
		if err == nil {
			// Other validations and messaging
			tamaChannel := db.GetTamaChannel(interaction.GuildID)
			if user.Timezone == "" {
				message = "You must run `/set-timezone` before running this command."
			} else if tamaChannel == "" {
				message = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if tamaChannel != interaction.ChannelID {
				message = fmt.Sprintf("You must run this command in the <#%s> channel.", tamaChannel)
			} else {
				message = fmt.Sprintf("The cost to buy an egg is %s point%s.",
					utils.FormatUIFloat(cost), utils.Plural(cost))
				if user.Points >= cost {
					allowPurchase = true
					message += fmt.Sprintf("\nYou currently have %s point%s, which is enough. Confirm purchase?",
						utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
				} else {
					message += fmt.Sprintf("\nYou currently have %s point%s, **which isn't enough**.",
						utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
				}
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
	serverId := interaction.GuildID
	userId := interaction.Member.User.ID
	tama, user, err := tamas.BuyTamaEggWorkflow(serverId, interaction.ChannelID, userId)

	var channelMessage string
	var responseMessage string
	if err == nil {
		channelMessage += fmt.Sprintf("# <@%s> has bought a Tama egg! :moneybag::arrow_right::egg:\n"+
			"- Egg ID is #%d.\n"+
			"- Purchased for %s point%s *(%s remaining)*",
			userId, tama.Id, utils.FormatUIFloat(constants.TamaEggPurchaseCost),
			utils.Plural(constants.TamaEggPurchaseCost), utils.FormatUIFloat(user.Points))

		responseMessage += fmt.Sprintf("**Success!** You now have %s point%s.\n",
			utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
		responseMessage += fmt.Sprintf("- Make sure to take care of it using `/care-tama %d`.\n", tama.Id)
		responseMessage += fmt.Sprintf("- You can check on the status of your egg with `/check-tama %d`.", tama.Id)
		// Add the user to the minigame role for later ease of @-ing them.
		err = tamas.AddUserToMinigameRole(session, serverId, userId)
		if err != nil {
			// Don't fail the whole command if this fails, just print it.
			log.Printf("Failed to add user %s to server %s's Tama minigame role: %s\n", userId, serverId, err.Error())
		}
	} else {
		responseMessage = fmt.Sprintf("Couldn't complete egg purchase: %s", err.Error())
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: responseMessage,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("Couldn't respond to egg purchase:", err)
	}
	if channelMessage != "" {
		_, err = session.ChannelMessageSend(interaction.ChannelID, channelMessage)
		if err != nil {
			log.Println("Couldn't send channel message:", err)
		}
	}

	log.Println("Tama egg purchase completed.")
}

func HandleBuyTamaEggCancelPurchase(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Cancelled.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err == nil {
		log.Println("Tama egg purchase cancelled.")
	} else {
		log.Println("Failed to send egg purchase response:", err.Error())
	}
}
