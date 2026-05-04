package tamas

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"go.etcd.io/bbolt"
)

func HandleTamaDeathWorkflow(serverId string, tamaId models.JacuzziId, cause string, tx *bbolt.Tx) {
	log.Printf("Scheduling death workflow for Tama %d.\n", tamaId)

	payload := models.TamaDeathPayload{
		ServerId: serverId,
		TamaId:   tamaId,
		Cause:    cause,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Could not marshall TamaDeathPayload: %s\n", err.Error())
		return
	}

	event := models.ScheduledEvent{
		ID:       fmt.Sprintf("TamaDeath-%s-%d", serverId, tamaId),
		NextTime: time.Now().Add(5 * time.Second),
		Handler:  "TamaDeathHandler",
		Payload:  payloadBytes,
	}

	db.ScheduleEvent(&event, true, tx)
}
