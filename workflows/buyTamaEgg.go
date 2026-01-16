package workflows

import (
	"log"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/errors"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

func makeTamaEgg(serverId string) (models.Tama, error) {
	var tama models.Tama

	id, err := db.ClaimNextJacuzziId(serverId)
	if err != nil {
		log.Println("Error creating Tama egg:", err)
		return tama, err
	}

	return models.Tama{
		Id:             id,
		IsEgg:          true,
		PositiveTraits: &utils.Set[models.PositiveTrait]{},
		NegativeTraits: &utils.Set[models.NegativeTrait]{},
		// Everything else is blank for an egg
	}, err
}

// Buy a Tama egg. Creates one from scratch, and deducts the cost of purchase from the user.
func BuyTamaEgg(serverId string, channelId string, userId string) (models.Tama, models.User, error) {
	var tama models.Tama
	var user models.User

	user, err := db.GetUser(serverId, userId)
	if err != nil {
		log.Printf("Could not fetch user details for %s: %s\n", userId, err.Error())
	}

	// Make sure the user has room for another egg
	if user.Tamas.Size() >= constants.TamaLimitPerUser {
		err = errors.TamaLimitReachedError{}
		log.Printf("%s -> User %s\n", err.Error(), userId)
		return tama, user, err
	}

	cost := constants.TamaEggPurchaseCost

	// Attempt to buy an egg
	user, err = db.ModifyUserPoints(serverId, userId, -cost)
	if err != nil {
		log.Println("Could not modify user points:", err)
		return tama, user, err
	}

	// Since now we have to be careful about refunding their points if any errors happen, invert our error handling.
	tama, err = makeTamaEgg(serverId)
	if err == nil {
		// Mark the egg as owned by the requesting user.
		tama.Owner = userId
		// Store the Tama's details in the DB.
		err = db.StoreTama(serverId, channelId, &tama)
	}
	// Mark the user as the owner of this Tama and increment their purchase count.
	if err == nil {
		user, err = db.ModifyUserTamas(serverId, userId, db.Add, tama.Id)
	}

	if err != nil {
		log.Println("Failed to buy Tama egg:", err)
		// Refund the user's points.
		_, err2 := db.ModifyUserPoints(serverId, userId, cost)
		if err2 != nil {
			log.Printf("Failed to refund user %s's %.0f points: %s\n", userId, cost, err2.Error())
			return tama, user, err2
		}
	}

	return tama, user, err
}
