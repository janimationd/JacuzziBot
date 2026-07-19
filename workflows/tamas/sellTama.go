package tamas

import (
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
)

type SellTamaWorkflowResult struct {
	// The final state of the deleted Tama
	DeletedTama *models.Tama
	// The full user details of its owner
	Owner models.User
	// If the Tama was in love, the full updated details of the target of its love
	LoveTargetTama *models.Tama
}

// Sell a Tama. Deletes it from the game database and pays out the owner based on its sell value.
func SellTamaWorkflow(
	serverId string,
	userId string,
	tamaId models.JacuzziId,
) (SellTamaWorkflowResult, error) {
	result := SellTamaWorkflowResult{}
	var err error
	result.DeletedTama, err = db.GetTama(serverId, tamaId)
	if err != nil {
		log.Printf("Couldn't fetch tama details to sell: %s", err.Error())
		return result, err
	}
	tama := result.DeletedTama

	sellValue, _ := tama.SellValueAndEquation()

	// Pay the owner the sell value. Lower risk operations first!
	result.Owner, err = db.ModifyUserPoints(serverId, userId, sellValue)
	if err != nil {
		log.Println("Could not modify user points:", err)
		return result, err
	}

	// Delete the Tama from the DB
	deleteTamaResult, err := db.DeleteTama(serverId, tamaId)
	if err != nil {
		// Deduct the payout since the Tama wasn't actually sold. Allow debt just in case.
		result.Owner, err = db.ModifyUserPointsWithDebt(serverId, userId, -sellValue, true)
		if err != nil {
			log.Printf("Couldn't reverse Tama sale payout: %s", err.Error())
		}
		return result, err
	}
	result.LoveTargetTama = deleteTamaResult.LoveTarget

	statusString := GetTamaStatus(tama, time.UTC, "")
	log.Printf("Sold tama:\n\n%s\n", statusString)

	return result, nil
}
