package tamas

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/constants"
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

	var newOwnerUser models.User
	if errorMessage == "" {
		newOwnerUser, err = db.GetUser(serverId, newOwnerId)
		if err != nil {
			errorMessage = "Couldn't fetch user details: " + err.Error()
		}
	}

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
		} else if newOwnerUser.Tamas.Size() >= constants.TamaLimitPerUser {
			errorMessage = fmt.Sprintf("You are already at the limit of how many Tamas you can own: %d.",
				constants.TamaLimitPerUser)
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

	// Update both users' lists of owned Tamas, being careful to rollback all state changes so far on error.
	if errorMessage == "" {
		_, err = db.ModifyUserTamas(serverId, oldOwnerId, db.Remove, tamaId)
		if err != nil {
			errorMessage = "Couldn't remove Tama from original owner: " + err.Error()
		} else {
			_, err = db.ModifyUserTamas(serverId, newOwnerId, db.Add, tamaId)
			if err != nil {
				errorMessage = "Couldn't add Tama to new owner: " + err.Error()

				_, err = db.ModifyUserTamas(serverId, oldOwnerId, db.Add, tamaId)
				if err != nil {
					errorMessage += ", and couldn't rollback Tama removal from original owner: " + err.Error()
				}
			}
		}

		if errorMessage != "" {
			_, err = db.ChangeTamaOwner(serverId, tamaId, oldOwnerId, true)
			if err != nil {
				errorMessage += ", and couldn't rollback Tama owner change: " + err.Error()
			}
		}
	}

	if errorMessage == "" {
		err := AddUserToMinigameRole(session, serverId, newOwnerId)
		if err != nil {
			// Don't fail the whole command if this fails, just print it.
			log.Println("Failed to add user to Tama minigame role:", err)
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
