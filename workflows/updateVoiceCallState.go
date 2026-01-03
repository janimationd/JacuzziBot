package workflows

import (
	"log"

	"github.com/janimationd/JacuzziBot/db"
)

func UpdateVoiceCallState(
	serverId string,
	previousChannelId string,
	newChannelId string,
	userId string,
	userDisplayName string,
) {
	// Remove from the old voice call if there was one.
	if previousChannelId != "" {
		_, err := db.ModifyUsers(serverId, previousChannelId, db.Remove, userId)
		if err != nil {
			log.Println("Couldn't remove user from previous voice call:", err)
		} else {
			log.Printf("Removed user %s (%s) from old voice call in channel %s\n", userId, userDisplayName, previousChannelId)
		}
	}

	// Add into the new voice call if there is one
	if newChannelId != "" {
		_, err := db.ModifyUsers(serverId, newChannelId, db.Add, userId)
		if err != nil {
			log.Println("Couldn't add user to new voice call:", err)
		} else {
			log.Printf("Added user %s (%s) to new voice call in channel %s\n", userId, userDisplayName, newChannelId)
		}
	}
}
