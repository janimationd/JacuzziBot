package events

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord"
	"github.com/janimationd/JacuzziBot/models"
)

func chooseRandomOtherTama(firstTama *models.Tama, allTamas map[models.JacuzziId]*models.Tama) *models.Tama {
	if len(allTamas) < 2 {
		panic(fmt.Sprintf("Cannot choose another tama when there aren't at least 2 total: %d", len(allTamas)))
	}
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

	for _, tamaStale := range tamas {
		// Previous iterations of the loop might've updated the Tamas here in the DB, so we should fetch them again.
		otherTamaStale := chooseRandomOtherTama(tamaStale, tamas)
		tamaFresh, err := db.GetTama(serverId, tamaStale.Id)
		if err != nil {
			log.Printf("Couldn't get Tama #%d's details from DB: %s\n", tamaStale.Id, err.Error())
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

		summary, err := db.TamaInteract(serverId, tamaFresh.Id, otherTamaFresh.Id, "")
		if err != nil {
			log.Printf("%s couldn't interact with %s: %s. Skipping.\n",
				tamaFresh.GetNameAndId(), otherTamaFresh.GetNameAndId(), err.Error())
			continue
		}

		// Handle the effect of the Social butterfly trait (33% chance)
		if tamaFresh.PositiveTraits.Contains(models.SocialButterfly) && 0 == rand.IntN(3) {
			// TODO
		}

		// Record a string about what happen for later construction of a channel status message
	}

	// Construct the channel status message
	// Figure out the registered Tama channel (do this earlier actually)
	// Send the message to the registered Tama channel

	return false
}
