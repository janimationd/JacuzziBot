package events

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"go.etcd.io/bbolt"
)

// Tama pets that are very happy earn points for their owners.
var TamaMoodEarnsPointsAwarder = models.ScheduledEvent{
	ID:                  "TamaMoodEarnsPointsAwarder",
	Interval:            1 * time.Hour,
	Handler:             "TamaMoodEarnsPointsHandler",
	RestartGapTolerance: 6 * time.Hour,
}

func TamaMoodEarnsPointsHandler(event *models.ScheduledEvent, _ *bbolt.Tx) bool {
	log.Printf("Event %s executed.\n", event.ID)

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

		numTamas := len(tamas)
		log.Printf("%d Tamas found in server %s.\n", numTamas, serverId)

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

		// Record a string about what happened for later construction of a channel status message
		summary := "# Tama productivity\nHappy Tamas find trinkets, earning their owners points every hour. Here's what happened:"

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
			// Skip tamas with too low of mood
			if tama.Mood < models.TamaMinorPointActionMoodThreshold {
				continue
			}

			summary += "\n"

			// Per-owner headers
			if lastOwner != tama.Owner {
				summary += fmt.Sprintf("## <@%s>'s Tamas\n", tama.Owner)
				lastOwner = tama.Owner
			}

			var reward float64
			if tama.Mood >= models.TamaMajorPointActionMoodThreshold {
				reward = constants.TamaMajorPointActionReward
			} else {
				reward = constants.TamaMinorPointActionReward
			}

			db.ModifyUserPoints(serverId, tama.Owner, reward)

			summary += fmt.Sprintf("- %s earned you %s point%s because it was %s.", tama.GetNameAndId(),
				utils.FormatUIFloat(reward), utils.Plural(reward), tama.GetMoodString())
		}

		// Send the summary message to the channel
		session.Handle.ChannelMessageSend(channelId, summary)
	}

	return false
}
