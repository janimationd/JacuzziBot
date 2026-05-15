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
	log.Println(">>> TamaHungerHandler: 1")
	if session.Handle == nil {
		log.Println("Discord session is nil, not checking Tama hunger.")
		return false
	}
	log.Println(">>> TamaHungerHandler: 2")

	// Iterate over all of the servers/guilds we're currently connected to.
	for _, guild := range session.Handle.State.Guilds {
		log.Printf(">>> TamaHungerHandler: 3 - %s\n", guild.ID)
		serverId := guild.ID
		allTamas, err := db.GetAllTamas(serverId, "", true, true)
		if err != nil {
			log.Printf("Unable to fetch Tamas for server %s (%s): %s\n", serverId, guild.Name, err.Error())
			continue
		}
		log.Println(">>> TamaHungerHandler: 4")

		channelId := db.GetTamaChannel(serverId)
		if channelId == "" {
			log.Printf("Couldn't react to Tama hunger for server %s: no registered Tama channel.\n", serverId)
			continue
		}
		log.Println(">>> TamaHungerHandler: 5")

		// Sort tamas by owner
		sortedTamas := make([]*models.Tama, len(allTamas))
		i := 0
		for _, tama := range allTamas {
			sortedTamas[i] = tama
			i++
		}
		log.Println(">>> TamaHungerHandler: 6")
		slices.SortFunc(sortedTamas, func(a, b *models.Tama) int {
			return cmp.Compare(a.Owner, b.Owner)
		})
		log.Println(">>> TamaHungerHandler: 7")

		summary := "# Tamas were hungry! :weary::thought_balloon::poultry_leg:\n " +
			"This message is being sent because midnight just passed in some players' local timezones, " +
			"and they forgot to feed their pets yesterday!"
		lastOwner := ""
		anyTamasWereHungry := false
		// Iterate over all of the Tamas in that server
		for _, tama := range sortedTamas {
			log.Println(">>> TamaHungerHandler: 8")
			user, err := db.GetUser(serverId, tama.Owner)
			if err != nil {
				log.Printf("Couldn't fetch user %s in server %s from DB: %s\n", tama.Owner, serverId, err.Error())
				continue
			}
			log.Println(">>> TamaHungerHandler: 9")
			timezone, err := time.LoadLocation(user.Timezone)
			if err != nil {
				log.Printf("Couldn't convert user %s's timezone %s: %s\n", tama.Owner, user.Timezone, err.Error())
				continue
			}
			log.Println(">>> TamaHungerHandler: 10")

			// Check if user's midnight just happened
			y, m, d := now.In(timezone).Date()
			lastMidnight := time.Date(y, m, d, 0, 0, 0, 0, timezone)
			log.Printf(">>> TamaHungerHandler: 11 - event.LastCheckTime=%s, lastMidnight=%s\n",
				event.LastCheckTime.String(), lastMidnight.String())
			midnightJustPassed := event.LastCheckTime.Before(lastMidnight) && !lastMidnight.After(now)
			if event.LastCheckTime.IsZero() || !midnightJustPassed {
				log.Println(">>> TamaHungerHandler: 11.5")
				// The owner's last local midnight didn't happen between this check and the last, or this is the first
				// time we're checking, so skip.
				continue
			}
			log.Println(">>> TamaHungerHandler: 12")

			result, err := db.TamaReactToHunger(serverId, tama.Id, timezone)
			if err != nil {
				log.Println(">>> TamaHungerHandler: 13")
				// db code already logged the problem
				continue
			}
			log.Println(">>> TamaHungerHandler: 14")
			if result.JustDied {
				cause := "Neglect (hunger)"
				tamas.HandleTamaDeathWorkflow(serverId, tama.Id, cause, tx)
				log.Println(">>> TamaHungerHandler: 15")
			}
			log.Println(">>> TamaHungerHandler: 16")
			if result.FinalMoodDelta < 0 {
				log.Println(">>> TamaHungerHandler: 17")
				anyTamasWereHungry = true

				// Per-owner headers
				if lastOwner != tama.Owner {
					summary += fmt.Sprintf("\n## <@%s>'s hungry Tamas", tama.Owner)
					lastOwner = tama.Owner
					log.Println(">>> TamaHungerHandler: 18")
				}
				summary += result.Summary
				log.Println(">>> TamaHungerHandler: 19")
			}
			log.Println(">>> TamaHungerHandler: 20")
		}
		log.Println(">>> TamaHungerHandler: 21")

		if anyTamasWereHungry {
			log.Println(">>> TamaHungerHandler: 22")
			// Send the summary message to the channel
			session.Handle.ChannelMessageSend(channelId, summary)
		}
		log.Println(">>> TamaHungerHandler: 23")
	}
	log.Println(">>> TamaHungerHandler: 24")
	return false
}
