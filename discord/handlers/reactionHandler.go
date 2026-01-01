package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/workflows"
)

// Message handler
func ReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	log.Printf("Incoming reaction by %s (%s): :%s:\n",
		r.Member.User.ID, r.Member.User.DisplayName(), r.Emoji.Name)

	// Ignore all messages from the bot itself
	if r.Member.User.ID == s.State.User.ID {
		log.Println(r.Member.User.ID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	// Every other reaction gets the author and recipient some points!
	workflows.ReactionGetsPoints(s, r)
}
