package events

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"slices"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
	"go.etcd.io/bbolt"
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
	index := rand.IntN(max)
	targetTama := allTamas[keys[index]]
	log.Printf("%s has chosen to interact with %s.\n", firstTama.GetNameAndId(), targetTama.GetNameAndId())
	return targetTama
}

func interact(
	serverId string,
	tamaStale *models.Tama,
	otherTamas map[models.JacuzziId]*models.Tama,
	indent string,
	tx *bbolt.Tx,
) (string, error) {
	otherTamaStale := chooseRandomOtherTama(tamaStale, otherTamas)

	// Previous interactions might've updated the Tamas here in the DB, so we should fetch them again.
	tama, err := db.GetTama(serverId, tamaStale.Id)
	if err != nil {
		return "", fmt.Errorf("Couldn't get Tama #%d's details from DB: %w\n", tamaStale.Id, err)
	}
	otherTama, err := db.GetTama(serverId, otherTamaStale.Id)
	if err != nil {
		return "", fmt.Errorf("Couldn't get Tama #%d's details from DB: %w\n", otherTamaStale.Id, err)
	}

	// They also might've died... so just skip this interaction if yes
	if tama.IsDead() || otherTama.IsDead() {
		return "", fmt.Errorf("Skipping newly dead Tama.")
	}

	result, err := db.TamaInteract(serverId, tama.Id, otherTama.Id, indent)
	if err != nil {
		log.Printf("%s couldn't interact with %s: %s. Skipping.\n",
			tama.GetNameAndId(), otherTama.GetNameAndId(), err.Error())
		return "", err
	}

	for _, newlyDeadTama := range result.NewlyDeadTamas {
		var interactingTama *models.Tama
		if newlyDeadTama.Id == tama.Id {
			interactingTama = otherTama
		} else {
			interactingTama = tama
		}
		cause := fmt.Sprintf("An interaction with %s", interactingTama.GetNameAndId())
		// Kick off the delayed workflow for a Tama dying (returns quickly).
		tamas.HandleTamaDeathWorkflow(serverId, newlyDeadTama.Id, cause, tx)
	}

	return result.Summary, nil
}

func TamaPlaytimeHandler(event *models.ScheduledEvent, tx *bbolt.Tx) bool {
	log.Printf("Event %s executed.\n", event.ID)

	if session.Handle == nil {
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

	channelId := db.GetTamaChannel(serverId)
	if channelId == "" {
		log.Printf("Couldn't execute playtime: no Tama channel is registered.\n")
		return false
	}

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

	// Record a string about what happened for later construction of a channel status message
	summary := "# Play time!\nEveryone's Tama's played together. Here's what happened:"

	// Sort a list of all the Tamas by the Owner, so we can group the summary strings into headings per owner.
	sortedTamas := make([]*models.Tama, numTamas)
	i := 0
	for _, tama := range tamas {
		sortedTamas[i] = tama
		i++
	}
	slices.SortFunc(sortedTamas, func(a, b *models.Tama) int {
		return cmp.Compare(a.Owner, b.Owner)
	})

	lastOwner := ""
	for _, tama := range sortedTamas {
		log.Printf("Processing interactions for Tama %s.\n", tama.GetNameAndId())
		summary += "\n"

		// Per-owner headers
		if lastOwner != tama.Owner {
			summary += fmt.Sprintf("## <@%s>'s Tamas\n", tama.Owner)
			lastOwner = tama.Owner
		}

		summaryFirst, err := interact(serverId, tama, tamas, "", tx)
		if err != nil {
			log.Printf("Tama couldn't interact, skipping: %s\n", err.Error())
			continue
		}
		summary += summaryFirst

		// Handle the effect of the Social Butterfly trait (33% chance)
		if tama.PositiveTraits.Contains(models.SocialButterfly) && 0 == rand.IntN(3) {
			summarySecond, err := interact(serverId, tama, tamas, "  ", tx)
			if err != nil {
				log.Printf("Tama couldn't interact again, skipping: %s\n", err.Error())
				continue
			}
			summary += fmt.Sprintf(
				"\n- %s interacted a 2nd time because of its Social Butterfly trait (33%% chance):\n%s",
				tama.GetNameAndId(), summarySecond)
		}
	}

	// Send the summary message to the channel
	session.Handle.ChannelMessageSend(channelId, summary)

	return false
}
