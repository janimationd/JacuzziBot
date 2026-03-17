package events

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord"
	"github.com/janimationd/JacuzziBot/models"
)

func chooseRandomOtherTama(firstTama *models.Tama, allTamas map[models.JacuzziId]*models.Tama) *models.Tama {
	// We use one less than the count of Tamas since we're excluding ourself
	keys := make([]models.JacuzziId, 0, len(allTamas)-1)
	for k := range allTamas {
		// Don't include the initiating tama as a possible target
		if k != firstTama.Id {
			keys = append(keys, k)
		}
	}
	max := len(keys)
	index := rand.Intn(max)
	return allTamas[keys[index]]
}

func TamaPlaytimeHandler(event *models.ScheduledEvent) bool {
	log.Printf("Event %s executed.\n", event.ID)

	if discord.Session == nil {
		log.Println("Discord session is nil, not coordinating Tama playtime.")
		return false
	}

	// Parse out the event payload
	payload := models.TamaPlaytimePayload{}
	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Printf("Couldn't unmarshall JSON to TamaPlaytimePayload (%s): %s\n", event.Payload, err.Error())
		return false
	}
	serverId := payload.ServerId

	// Get all alive, hatched Tamas on the server
	tamas, err := db.GetAllTamas(serverId, "", true, true)
	if err != nil {
		log.Printf("Couldn't get all Tamas in server %s: %s\n", serverId, err.Error())
		return false
	}

	numTamas := len(tamas)
	if numTamas < 2 {
		log.Printf("Not enough Tamas in server %s for playtime (only %d).\n", serverId, numTamas)
		return false
	}
	log.Printf("%d Tamas found in server %s.\n", numTamas, serverId)

	for _, tama := range tamas {
		// Previous iterations of the loop might've updated the Tamas here in the DB, so we should fetch them again.
		otherTamaStale := chooseRandomOtherTama(tama, tamas)
		tamaFresh, err := db.GetTama(serverId, tama.Id)
		if err != nil {
			log.Printf("Couldn't get Tama #%d's details from DB: %s\n", tama.Id, err.Error())
			continue
		}
		otherTamaFresh, err := db.GetTama(serverId, otherTamaStale.Id)
		if err != nil {
			log.Printf("Couldn't get Tama #%d's details from DB: %s\n", otherTamaStale.Id, err.Error())
			continue
		}

		if tamaFresh.IsDead() || otherTamaFresh.IsDead() {
			log.Println("Skipping newly dead Tama.")
			continue
		}

		// Choose a random interaction
		// Execute the interaction
		// Account for relevant traits
		// Write any changes back to DB
		// Record a string about what happen for later construction of a channel status message
	}

	// Construct the channel status message
	// Figure out the registered Tama channel (do this earlier actually)
	// Send the message to the registered Tama channel

	return false
}
