package models

import "github.com/janimationd/JacuzziBot/utils"

// Data we track about ongoing voice calls
type VoiceCall struct {
	// The voice channel ID
	ChannelId string
	// The user IDs currently participating in the voice call
	Users utils.Set[string]
}

func (v VoiceCall) ContainsUser(userId string) bool {
	return v.Users.Contains(userId)
}
