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

const transferTamaCommandName string = "TransferTama"
const TransferTamaAcceptTransferId string = transferTamaCommandName + ":AcceptTransfer"
const TransferTamaCancelTransferId string = transferTamaCommandName + ":CancelTransfer"

// Transfer a Tama to another person
var TransferTama = models.SlashCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "transfer-tama",
		Description: "Transfer a Tama to another person.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The ID of the Tama you want to transfer.",
				Required:    true,
			},
			{
				Type: discordgo.ApplicationCommandOptionUser,
				Name: "new-owner",
				Description: "The person to transfer the Tama to. " +
					"They'll have to accept the transfer after you initiate it.",
				Required: true,
			},
		},
	},
	Handler: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		intId := utils.GetCommandOption(interaction, "id").IntValue()
		tamaId := models.JacuzziId(intId)
		newOwnerDiscordUser := utils.GetCommandOption(interaction, "new-owner").UserValue(session)
		newOwnerId := newOwnerDiscordUser.ID
		userId := interaction.Member.User.ID
		serverId := interaction.GuildID
		botUserId := session.State.User.ID

		registeredChannelId := db.GetTamaChannel(serverId)
		var errorMessage string

		_, err := db.GetUser(serverId, userId)
		if err != nil {
			errorMessage = fmt.Sprintf("Couldn't load user details: %s%s",
				err.Error(), constants.ErrorReportMessageSuffix)
		}
		newOwnerUser, err := db.GetUser(serverId, newOwnerId)
		if errorMessage == "" && err != nil {
			errorMessage = fmt.Sprintf("Couldn't load user details: %s%s",
				err.Error(), constants.ErrorReportMessageSuffix)
		}

		// Basic validations
		if errorMessage == "" {
			if intId <= int64(models.NoId) {
				errorMessage = fmt.Sprintf("Tama IDs must be greater than %d.", models.NoId)
			} else if userId == newOwnerId {
				errorMessage = "You can only transfer to someone else."
			} else if botUserId == newOwnerId {
				errorMessage = "Thanks, but I'm not playing the game :smile:"
			} else if registeredChannelId == "" {
				errorMessage = "No channel is registered as a Tama minigame yet (talk to an admin)."
			} else if registeredChannelId != interaction.ChannelID {
				errorMessage = fmt.Sprintf("You must run this command in the <#%s> channel.", registeredChannelId)
			} else if newOwnerUser.Timezone == "" {
				errorMessage = fmt.Sprintf("<@%s> must run `/set-timezone` before receiving a Tama.", newOwnerId)
			}
		}

		// Make sure the user owns the Tama
		var tama *models.Tama
		if errorMessage == "" {
			tama, err = db.GetTama(serverId, tamaId)
			if err != nil {
				errorMessage = err.Error()
			} else if tama.Owner != userId {
				errorMessage = "You can only transfer Tamas that you own."
			}
		}

		if errorMessage == "" {
			_, err = db.CreateTamaTransfer(serverId, tamaId, userId, newOwnerId)
			if err != nil {
				errorMessage = "Couldn't create Tama transfer: " + err.Error()
			}
		}

		if errorMessage == "" {
			message := fmt.Sprintf("<@%s> wants to transfer Tama #%d to <@%s>.", userId, tamaId, newOwnerId)
			acceptButtonIdWithArgs := fmt.Sprintf("%s|%d", TransferTamaAcceptTransferId, tamaId)
			cancelButtonIdWithArgs := fmt.Sprintf("%s|%d", TransferTamaCancelTransferId, tamaId)

			// Respond with feedback message
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Accept",
									Style:    discordgo.PrimaryButton,
									CustomID: acceptButtonIdWithArgs,
								},
								discordgo.Button{
									Label:    "Reject/Cancel",
									Style:    discordgo.DangerButton,
									CustomID: cancelButtonIdWithArgs,
								},
							},
						},
					},
				},
			})
		} else {
			// Respond with feedback message
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

func HandleTransferTamaAcceptTransfer(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var errorMessage string

	serverId := interaction.GuildID
	userId := interaction.Member.User.ID

	customId := interaction.MessageComponentData().CustomID
	tamaId, err := models.ExtractJacuzziIdFromCustomId(customId)
	if err != nil {
		errorMessage = err.Error()
	}

	var tamaTransfer *models.TamaTransfer
	if errorMessage == "" {
		var err error
		tamaTransfer, err = tamas.TransferTamaWorkflow(serverId, tamaId, userId, session, interaction)
		if err != nil {
			errorMessage = err.Error()
		}
	}

	if errorMessage != "" {
		log.Println(errorMessage)
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println("Failed to send error message:", err)
		}
	} else {
		message := fmt.Sprintf("> %s\nTransfer of Tama #%d from <@%s> to <@%s> complete!",
			interaction.Message.Content, tamaId, tamaTransfer.OldOwnerId, tamaTransfer.NewOwnerId)
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    message,
				Components: []discordgo.MessageComponent{},
			},
		})
		if err != nil {
			log.Println("Failed to update original transfer message:", err)
		} else {
			log.Println(message)
		}
	}
}

func HandleTransferTamaCancelTransfer(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var errorMessage string

	serverId := interaction.GuildID
	userId := interaction.Member.User.ID

	customId := interaction.MessageComponentData().CustomID
	tamaId, err := models.ExtractJacuzziIdFromCustomId(customId)
	if err != nil {
		errorMessage = err.Error()
	}

	var tamaTransfer *models.TamaTransfer
	if errorMessage == "" {
		var err error
		tamaTransfer, err = db.GetTamaTransfer(serverId, tamaId)
		if err != nil {
			errorMessage = err.Error()
		}
	}

	if errorMessage == "" {
		if tamaTransfer.OldOwnerId != userId && tamaTransfer.NewOwnerId != userId {
			errorMessage = "Only the people involved in the transfer can cancel it."
		}
	}

	if errorMessage == "" {
		err := db.DeleteTamaTransfer(serverId, tamaId)
		if err != nil {
			errorMessage = err.Error()
		}
	}

	if errorMessage != "" {
		log.Println(errorMessage)
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println("Failed to send error message:", err)
		}
	} else {
		message := fmt.Sprintf("> %s\nTama transfer cancelled.", interaction.Message.Content)
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    message,
				Components: []discordgo.MessageComponent{},
			},
		})
		if err != nil {
			log.Println("Failed to update original transfer message:", err)
		} else {
			log.Println(message)
		}
	}
}
