package events

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"github.com/janimationd/JacuzziBot/workflows"
	"go.etcd.io/bbolt"
)

func SchedulePredictionStateUpdateEvent(
	serverId string,
	channelId string,
	predictionId models.JacuzziId,
	stateUpdateTime time.Time,
) (*models.ScheduledEvent, error) {
	payload := models.PredictionStateUpdateEventPayload{
		ServerId:     serverId,
		ChannelId:    channelId,
		PredictionId: predictionId,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	event := &models.ScheduledEvent{
		ID: fmt.Sprintf("PredictionStateUpdate-%s-%s-%s",
			serverId, models.StringFromJacuzziId(predictionId), stateUpdateTime.Format(utils.EventIdTimeFormat)),
		NextTime: stateUpdateTime,
		Handler:  "PredictionStateUpdateHandler",
		Payload:  payloadBytes,
	}
	scheduled, err := db.ScheduleEvent(event, true, nil)
	if err != nil {
		return nil, err
	}
	if !scheduled {
		return nil, fmt.Errorf("Failed to schedule prediction state update event")
	}
	return event, nil
}

func PredictionStateUpdateHandler(event *models.ScheduledEvent, now time.Time, _ *bbolt.Tx) bool {
	payload := &models.PredictionStateUpdateEventPayload{}
	err := json.Unmarshal(event.Payload, payload)
	if err != nil {
		log.Printf("Couldn't unmarshall prediction state update event payload: %s\n", err.Error())
		return false
	}
	log.Printf("Handling state update event for prediction %d\n", payload.PredictionId)

	prediction, err := db.GetPrediction(payload.ServerId, payload.PredictionId)
	if err != nil {
		log.Printf("Couldn't fetch prediction from DB for state update: %s\n", err.Error())
		return false
	}

	workflows.CreateOrUpdatePredictionMessage(prediction, payload.ChannelId, now)
	_, err = session.Handle.ChannelMessageEdit(payload.ChannelId, prediction.MessageId, prediction.DisplayString(now))
	if err != nil {
		log.Printf("Couldn't edit original prediction channel message `%s` when updating state: %s",
			prediction.MessageId, err.Error())
	}
	return false
}
