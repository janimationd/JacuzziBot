package events

import (
	"cmp"
	"log"
	"slices"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"go.etcd.io/bbolt"
)

// Tama pets that are very happy earn points for their owners.
var TamaMoodEarnsPointsAwarder = models.ScheduledEvent{
	ID:                  "TamaMoodEarnsPointsAwarder",
	Interval:            time.Hour,
	Handler:             "TamaMoodEarnsPointsHandler",
	RestartGapTolerance: 2 * time.Hour,
}

func TamaMoodEarnsPointsHandler(event *models.ScheduledEvent, _ time.Time, _ *bbolt.Tx) bool {
	if session.Handle == nil {
		log.Println("Discord session is nil, not earning any points from Tama mood.")
		return false
	}

	for _, guild := range session.Handle.State.Guilds {
		serverId := guild.ID

		channelId := db.GetTamaChannel(serverId)
		if channelId == "" {
			// Server doesn't have the minigame set up yet, ignore.
			return false
		}

		// Get all alive, hatched Tamas on the server
		tamas, err := db.GetAllTamas(serverId, "", true, true)
		if err != nil {
			log.Printf("Couldn't get all Tamas in server %s: %s\n", serverId, err.Error())
			return false
		}

		anyPointEarners := false
		for _, tama := range tamas {
			if tama.Mood >= models.TamaMinorPointActionMoodThreshold {
				anyPointEarners = true
				break
			}
		}

		if !anyPointEarners {
			log.Println("No Tamas are happy enough to earn points right now.")
			return false
		}

		// Sort a list of all the Tamas by the Owner, so we can group the summary strings into headings per owner.
		sortedTamas := make([]*models.Tama, len(tamas))
		i := 0
		for _, tama := range tamas {
			sortedTamas[i] = tama
			i++
		}
		slices.SortFunc(sortedTamas, func(a, b *models.Tama) int {
			return cmp.Compare(a.Owner, b.Owner)
		})

		for _, tama := range sortedTamas {
			// Skip tamas with too low of mood
			if tama.Mood < models.TamaMinorPointActionMoodThreshold {
				continue
			}

			var reward float64 = tama.GetHourlyPointAward()
			db.ModifyUserPoints(serverId, tama.Owner, reward)

			log.Printf("%s earned %s %s point%s because it was %s.", tama.GetNameAndId(),
				tama.Owner, utils.FormatUIFloat(reward), utils.Plural(reward), tama.GetMoodString())
		}
	}

	return false
}
