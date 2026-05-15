package events

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/workflows/tamas"
	"go.etcd.io/bbolt"
)

// Every minute check to see if midnight has happened in any Tama owners' local timezones, and see if any Tamas are
// hungry enough to lose mood.
var TamaHungerChecker = models.ScheduledEvent{
	ID:                  "TamaHungerChecker",
	Interval:            10 * time.Second,
	Handler:             "TamaHungerHandler",
	RestartGapTolerance: 5 * time.Minute,
	UsesLastCheckTime:   true,
}

func TamaHungerHandler(event *models.ScheduledEvent, now time.Time, tx *bbolt.Tx) bool {
	if session.Handle == nil {
		log.Println("Discord session is nil, not checking Tama hunger.")
		return false
	}

	// Iterate over all of the servers/guilds we're currently connected to.
	for _, guild := range session.Handle.State.Guilds {
		serverId := guild.ID
		allTamas, err := db.GetAllTamas(serverId, "", true, true)
		if err != nil {
			log.Printf("Unable to fetch Tamas for server %s (%s): %s\n", serverId, guild.Name, err.Error())
			continue
		}

		channelId := db.GetTamaChannel(serverId)
		if channelId == "" {
			log.Printf("Couldn't react to Tama hunger for server %s: no registered Tama channel.\n", serverId)
			continue
		}

		// Sort tamas by owner
		sortedTamas := make([]*models.Tama, len(allTamas))
		i := 0
		for _, tama := range allTamas {
			sortedTamas[i] = tama
			i++
		}
		slices.SortFunc(sortedTamas, func(a, b *models.Tama) int {
			return cmp.Compare(a.Owner, b.Owner)
		})

		summary := "# Tamas were hungry! :weary::thought_balloon::poultry_leg:\n " +
			"This message is being sent because midnight just passed in some players' local timezones, " +
			"and they forgot to feed their pets yesterday!"
		lastOwner := ""
		anyTamasWereHungry := false
		// Iterate over all of the Tamas in that server
		for _, tama := range sortedTamas {
			user, err := db.GetUser(serverId, tama.Owner)
			if err != nil {
				log.Printf("Couldn't fetch user %s in server %s from DB: %s\n", tama.Owner, serverId, err.Error())
				continue
			}
			timezone, err := time.LoadLocation(user.Timezone)
			if err != nil {
				log.Printf("Couldn't convert user %s's timezone %s: %s\n", tama.Owner, user.Timezone, err.Error())
				continue
			}

			// Check if user's midnight just happened
			y, m, d := now.In(timezone).Date()
			lastMidnight := time.Date(y, m, d, 0, 0, 0, 0, timezone)
			midnightJustPassed := event.LastCheckTime.Before(lastMidnight) && !lastMidnight.After(now)
			if event.LastCheckTime.IsZero() || !midnightJustPassed {
				// The owner's last local midnight didn't happen between this check and the last, or this is the first
				// time we're checking, so skip.
				continue
			}

			result, err := db.TamaReactToHunger(serverId, tama.Id, timezone)
			if err != nil {
				// db code already logged the problem
				continue
			}
			if result.JustDied {
				cause := "Neglect (hunger)"
				tamas.HandleTamaDeathWorkflow(serverId, tama.Id, cause, tx)
			}
			if result.FinalMoodDelta < 0 {
				anyTamasWereHungry = true

				// Per-owner headers
				if lastOwner != tama.Owner {
					summary += fmt.Sprintf("\n## <@%s>'s hungry Tamas", tama.Owner)
					lastOwner = tama.Owner
				}
				summary += result.Summary
			}
		}

		if anyTamasWereHungry {
			// Send the summary message to the channel
			session.Handle.ChannelMessageSend(channelId, summary)
		}
	}
	return false
}
