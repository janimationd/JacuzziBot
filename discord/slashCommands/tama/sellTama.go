package tama

import (
	"fmt"
	"log"
	"regexp"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
)

const sellTamaCommandName string = "SellTama"
const SellTamaConfirmSaleId string = sellTamaCommandName + ":Y"
const SellTamaCancelSaleId string = sellTamaCommandName + ":N"

var SellTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "sell-tama",
		Description: "Sell a Tama, removing it from the game permanently. Value increases with age and mood.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama to sell.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "flavor-text",
				Description: "Optional text describing how your Tama spends the rest of its life.",
				Required:    false,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		serverId := interaction.GuildID
		userId := interaction.Member.User.ID

		// Argument parsing
		intId := utils.GetCommandOption(interaction, "id").IntValue()
		flavorTextOption := utils.GetCommandOption(interaction, "flavor-text")
		flavorText := ""
		if flavorTextOption != nil {
			flavorText = flavorTextOption.StringValue()
			log.Printf("Flavor text: %s\n", flavorText)
		}

		invalidName, err := regexp.MatchString(constants.InputDenylistRegex, flavorText)
		// Regex compile error... we have bug
		if err != nil {
			panic(err)
		}

		user, err := db.GetUser(interaction.GuildID, userId)
		errorMessage := ""
		if err != nil {
			errorMessage = fmt.Sprintf("Failed to check purchase eligibility: %s", err.Error())
		}

		// Basic validations
		var id models.JacuzziId
		confirmButtonCustomId := ""
		if errorMessage == "" {
			tamaChannel := db.GetTamaChannel(interaction.GuildID)
			if user.Timezone == "" {
				errorMessage = "You must run `/set-timezone` before running this command."
			} else if tamaChannel == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if tamaChannel != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", tamaChannel)
			} else if intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if invalidName {
				errorMessage = fmt.Sprintf(
					"Flavor text must not contain any forbidden characters (``%s``) or newlines.",
					constants.InputDenylistForPrinting)
			} else {
				// Make sure the Discord API limit for CustomIds on the Confirm button is respected.
				prefix := fmt.Sprintf("%s|%d|", SellTamaConfirmSaleId, intId)
				confirmButtonCustomId = prefix + flavorText
				if utf8.RuneCountInString(confirmButtonCustomId) > constants.MaxCustomIdRuneLength {
					max := constants.MaxCustomIdRuneLength - utf8.RuneCountInString(prefix)
					errorMessage = fmt.Sprintf("Flavor text must be %d characters or less.", max)
				} else {
					// All these validations passed
					id = models.JacuzziId(intId)
				}
			}
		}

		// Tama validations
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(serverId, id)
			if err != nil {
				errorMessage = fmt.Sprintf("Couldn't fetch Tama info: %s", err.Error())
			} else if tama.IsDead() {
				errorMessage = "You can't sell a dead Tama!"
			} else if tama.IsEgg() {
				errorMessage = fmt.Sprintf(
					"You can't sell a Tama before it hatches! (`/care-tama %d` until it hatches)", id)
			} else if userId != tama.Owner {
				errorMessage = "You can only sell Tamas you own!"
			}
		}

		if errorMessage != "" {
			// Respond with error message
			err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: errorMessage,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})

			if err != nil {
				log.Println("InteractionRespond error:", err)
			}
		} else {
			// Confirmation prompt
			buttons := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Confirm",
							Style:    discordgo.DangerButton,
							CustomID: confirmButtonCustomId,
						},
						discordgo.Button{
							Label:    "Cancel",
							Style:    discordgo.PrimaryButton,
							CustomID: SellTamaCancelSaleId,
						},
					},
				},
			}

			sellValue, equation := tama.SellValueAndEquation()

			message := fmt.Sprintf("Tama %s would sell for **%s point%s** right now.\n\n",
				tama.GetNameAndId(), utils.FormatUIFloat(sellValue), utils.Plural(sellValue))
			message += fmt.Sprintf("`%s`\n\n", equation)
			// Describe consequences for its love target if it has one
			if tama.LoveTarget != models.NoId {
				loveTarget, err := db.GetTama(serverId, tama.LoveTarget)
				if err != nil {
					log.Printf("Couldn't fetch tama love target %d's details: %s\n", tama.LoveTarget, err.Error())
				} else {
					newMood := max(loveTarget.Mood-constants.TamaSaleLoveTargetMoodDamage, -models.TamaMoodLimit+1)
					message += fmt.Sprintf("Since %s loves your Tama, **selling will break its heart**, "+
						"causing it to lose %d mood (stopping short of death). This will leave it at %s.\n\n",
						loveTarget.GetNameAndId(), constants.TamaSaleLoveTargetMoodDamage,
						models.PreviewMoodString(newMood, loveTarget.Id))
				}
			}
			message += "Your Tama will leave the game **PERMANENTLY**. Are you sure you want to sell it?"

			// Respond with confirmation prompt
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
		}
	},
}

func HandleSellTamaConfirmSale(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	serverId := interaction.GuildID
	userId := interaction.Member.User.ID
	customId := interaction.MessageComponentData().CustomID
	tamaId, flavorText := models.ParseCustomIdToJacuzziIdAndString(customId)

	result, err := tamas.SellTamaWorkflow(serverId, userId, tamaId)
	var responseMessage string
	if err != nil {
		responseMessage = fmt.Sprintf("Couldn't sell Tama: %s", err.Error())
	}
	tama := result.DeletedTama
	user := result.Owner
	loveTarget := result.LoveTargetTama

	var channelMessage string
	if responseMessage == "" {
		sellValue, equation := tama.SellValueAndEquation()
		channelMessage = fmt.Sprintf("# <@%s> has sold Tama %s! :wave::face_holding_back_tears::moneybag:",
			userId, tama.GetNameAndId())
		channelMessage += fmt.Sprintf("\nSell value: **%s point%s**",
			utils.FormatUIFloat(sellValue), utils.Plural(sellValue))
		channelMessage += fmt.Sprintf("\n\n`%s`", equation)
		if flavorText != "" {
			channelMessage += fmt.Sprintf("\n\nProvided flavor text:\n> \"%s\"", flavorText)
		}
		if loveTarget != nil {
			channelMessage += fmt.Sprintf("\n\nTama %s (owned by <@%s>) was in love with this Tama, "+
				"**and its heart broke :broken_heart:**, taking %d mood damage (now at %s).",
				loveTarget.GetNameAndId(), loveTarget.Owner, constants.TamaSaleLoveTargetMoodDamage,
				loveTarget.GetMoodString())
		}

		responseMessage = fmt.Sprintf("**Success!** You now have %s point%s.\n",
			utils.FormatUIFloat(user.Points), utils.Plural(user.Points))
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: responseMessage,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("Couldn't respond to tama sale:", err)
	}
	if channelMessage != "" {
		_, err = session.ChannelMessageSend(interaction.ChannelID, channelMessage)
		if err != nil {
			log.Println("Couldn't send channel message:", err)
		}
	}

	log.Println("Tama sale completed.")
}

func HandleSellTamaCancelSale(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Cancelled.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err == nil {
		log.Println("Tama sale cancelled.")
	} else {
		log.Println("Failed to send tama sale cancel response:", err.Error())
	}
}
