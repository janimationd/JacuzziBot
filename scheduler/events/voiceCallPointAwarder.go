package events

import (
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"go.etcd.io/bbolt"
)

// When X people are in a voice call, award them each Y points every minute.
var VoiceCallPointAwarder = models.ScheduledEvent{
	ID:                  "VoiceCallPointAwarder",
	Interval:            1 * time.Minute,
	Handler:             "VoiceCallPointAwarderHandler",
	RestartGapTolerance: 5 * time.Minute,
}

func VoiceCallPointAwarderHandler(event *models.ScheduledEvent, _ time.Time, _ *bbolt.Tx) bool {
	if session.Handle == nil {
		log.Println("Discord session is nil, not checking voice calls.")
		return false
	}

	// Iterate over all of the servers/guilds we're currently connected to.
	for _, guild := range session.Handle.State.Guilds {
		voiceCalls, err := db.GetAllVoiceCallsWithParticipants(guild.ID)
		if err != nil {
			log.Printf("Unable to fetch voice calls for server %s (%s): %s\n", guild.ID, guild.Name, err.Error())
			continue
		}
		// Iterate over all of the voice calls in that server
		for _, voiceCall := range voiceCalls {
			participantCount := voiceCall.Users.Size()
			pointAwardValue := float64(participantCount) * constants.VoiceCallPointsPerParticipantPerMinute
			if participantCount > 0 {
				log.Printf("Awarding %.2f points to all %d users in voice channel %s.\n",
					pointAwardValue, participantCount, voiceCall.ChannelId)
			}
			// Iterate over all of the users in the voice call
			for userId := range voiceCall.Users.All() {
				db.ModifyUserPoints(guild.ID, userId, pointAwardValue)
			}
		}
	}
	return false
}
