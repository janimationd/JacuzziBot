package handlers

import (
	"log"
	"regexp"
	"strings"

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

	// See if the bot needs to respond to a mention
	if strings.Contains(m.Content, "<@"+s.State.User.ID+">") {
		// If the message is asking abaout the Tamas minigame
		tamasMentioned, err := regexp.MatchString("[^a-zA-Z][tT]amas{0,1}[^a-zA-Z]", m.Content)
		if err != nil {
			log.Printf("Couldn't check for Tamas mention in message: %s\n", err.Error())
		}
		if tamasMentioned {
			workflows.BotMentionPrintsTamaHelp(s, m)
		} else {
			workflows.BotMentionPrintsHelp(s, m)
		}
	}
}
