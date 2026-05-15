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
	"go.etcd.io/bbolt"
)

func TamaDeathHandler(event *models.ScheduledEvent, _ time.Time, tx *bbolt.Tx) bool {
	var payload models.TamaDeathPayload

	if session.Handle == nil {
		log.Printf("Cannot handle Tama death: Discord session is nil.\n")
		return false
	}

	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Printf("Couldn't unmarshall TamaDeathPayload (%s) from JSON: %s\n", event.Payload, err.Error())
		return false
	}

	serverId := payload.ServerId
	deadTamaId := payload.TamaId
	cause := payload.Cause

	channelId := db.GetTamaChannel(serverId)
	if channelId == "" {
		log.Println("Couldn't handle Tama death: no registered Tama channel.")
		return false
	}

	deadTama, err := db.GetTama(serverId, deadTamaId)
	if err != nil {
		log.Printf("Couldn't fetch details of dead tama to react: %s\n", err.Error())
		return false
	}

	summary := fmt.Sprintf("# :skull: %s has died! :headstone:\nOwned by: <@%s>\nCause: \"%s\"\n",
		deadTama.GetNameAndId(), deadTama.Owner, cause)

	allTamas, err := db.GetAllTamas(serverId, "", true, true)

	for tamaId, tama := range allTamas {
		// This shouldn't happen, but for safety
		if tamaId == deadTamaId {
			continue
		}

		result, err := db.TamaReactToDeath(serverId, tamaId, deadTamaId)
		if err != nil {
			log.Printf("Couldn't handle Tama %s's reaction to %s's death: %s\n",
				tama.GetNameAndId(), deadTama.GetNameAndId(), err.Error())
			continue
		}

		relationshipScore := tama.Relationships[deadTamaId]
		// Skip over Tamas who didn't care
		if result.FinalMoodDelta == 0 && relationshipScore == 0 {
			continue
		}

		summary += fmt.Sprintf("\n- %s's mood changed by %s%d in reaction (attitude towards it was %s%d)",
			tama.GetNameAndId(), utils.SignString(result.FinalMoodDelta), result.FinalMoodDelta,
			utils.SignString(relationshipScore), relationshipScore)
	}

	// Send the summary message to the channel
	session.Handle.ChannelMessageSend(channelId, summary)

	return false
}
