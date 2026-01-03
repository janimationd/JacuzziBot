package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/workflows"
)

func VoiceCallHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	log.Printf("Incoming voice state update for %s (%s) and channel %s\n",
		v.Member.User.ID, v.Member.DisplayName(), v.ChannelID)

	// Ignore all voice call interactions from the bot itself
	if v.UserID == s.State.User.ID {
		log.Println(v.UserID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	userPreviousVoiceCall, err := db.GetVoiceCallForUser(v.GuildID, v.UserID)
	if err != nil {
		log.Printf("Couldn't get previous voice call for user %s (%s), assuming none: %s\n",
			v.UserID, v.Member.DisplayName(), err.Error())
	}

	if v.ChannelID == userPreviousVoiceCall.ChannelId {
		log.Println("Previous and new channel are the same, skipping.")
		return
	}

	// Every other reaction gets the author and recipient some points!
	workflows.UpdateVoiceCallState(
		v.GuildID,
		userPreviousVoiceCall.ChannelId,
		v.ChannelID,
		v.UserID,
		v.Member.DisplayName(),
	)
}
