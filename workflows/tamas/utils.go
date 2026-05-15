package tamas

import (
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
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
		PositiveTraits: &utils.Set[models.PositiveTrait]{},
		NegativeTraits: &utils.Set[models.NegativeTrait]{},
		Parents:        &utils.Set[models.JacuzziId]{},
		Children:       &utils.Set[models.JacuzziId]{},
		EggLaidTime:    time.Now().Unix(),
		ServerId:       serverId,
		// Everything else is blank for an egg
	}, err
}
