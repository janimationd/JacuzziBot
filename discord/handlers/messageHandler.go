package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/workflows"
)

// Message handler
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Printf("Incoming message by %s (%s): %s\n",
		m.Message.Author.ID, m.Message.Author.DisplayName(), m.Message.Content)

	// Ignore all messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		log.Println(m.Author.ID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	// Every other message gets the author some points!
	workflows.MessageGetsPoints(s, m)

	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			// The bot was mentioned
			workflows.BotMentionPrintsHelp(s, m)
			// If mentioned multiple times, only print once
			break
		}
	}
}
