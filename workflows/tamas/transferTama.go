package tamas

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

func TransferTamaWorkflow(
	serverId string,
	tamaId models.JacuzziId,
	userId string,
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) (*models.TamaTransfer, error) {
	errorMessage := ""
	var err error

	tamaTransfer, err := db.GetTamaTransfer(serverId, tamaId)
	if err != nil {
		errorMessage = "Couldn't fetch Tama transfer details: " + err.Error()
	} else if tamaTransfer == nil {
		errorMessage = fmt.Sprintf("There is no pending transfer for Tama #%d.", tamaId)
	}
	oldOwnerId := tamaTransfer.OldOwnerId
	newOwnerId := tamaTransfer.NewOwnerId

	var tama *models.Tama
	if errorMessage == "" {
		tama, err = db.GetTama(serverId, tamaId)
		if err != nil {
			errorMessage = "Couldn't fetch Tama details: " + err.Error()
		}
	}

	// Basic validations
	if errorMessage == "" {
		if userId != newOwnerId {
			errorMessage = "Only the recipient can accept the transfer."
		} else if tama.Owner == newOwnerId {
			errorMessage = "The Tama has already been transferred to you."
		} else if tama.Owner != oldOwnerId {
			errorMessage = fmt.Sprintf("The tama is no longer owned by <@%s>, "+
				"so they can no longer transfer it to you.", oldOwnerId)
		}
	}

	// Update the Tama's owner
	if errorMessage == "" {
		_, err = db.ChangeTamaOwner(serverId, tamaId, newOwnerId, true)
		if err != nil {
			errorMessage = "Couldn't update Tama owner: " + err.Error()
		}
	}

	if errorMessage == "" {
		err := AddUserToMinigameRole(session, serverId, newOwnerId)
		if err != nil {
			// Don't fail the whole command if this fails, just print it.
			log.Printf("Failed to add user %s to server %s's Tama minigame role: %s\n", userId, serverId, err.Error())
		}
	}

	if errorMessage == "" {
		err = db.DeleteTamaTransfer(serverId, tamaId)
		if err != nil {
			errorMessage += "Transfer completed, but couldn't cleanup transfer state: " + err.Error()
		}
	}

	if errorMessage != "" {
		log.Println(errorMessage)
		return nil, fmt.Errorf(errorMessage)
	}

	return tamaTransfer, nil
}
