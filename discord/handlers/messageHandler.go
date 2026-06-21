package handlers

import (
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/workflows"
)

// Award points to the author of a new message
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Printf("Incoming message creation by %s (%s): %s\n",
		m.Message.Author.ID, m.Message.Author.DisplayName(), m.Message.Content)

	// Ignore all messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		log.Println(m.Author.ID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	// Every other message gets the author some points!
	workflows.CreateMessageGetsPoints(s, m)

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

// Subtract points from the deleter of a message
func MessageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	authorId := ""
	displayName := ""
	content := m.Content
	if m.Author != nil {
		log.Printf("Message deletion had m.Author info: %s\n", m.ID)
		authorId = m.Author.ID
		displayName = m.Author.DisplayName()
	}
	if m.BeforeDelete != nil {
		if m.BeforeDelete.Author != nil {
			log.Printf("Message deletion had m.BeforeDelete.Author details: %s\n", m.ID)
			if authorId == "" {
				authorId = m.BeforeDelete.Author.ID
			}
			if displayName == "" {
				displayName = m.BeforeDelete.Author.DisplayName()
			}
		}
		if content == "" {
			content = m.BeforeDelete.Content
		}
	}

	if authorId == "" {
		log.Printf("Message deletion had no author info: %s\n", m.ID)
		return
	}

	log.Printf("Incoming deletion of %s (%s)'s message: %s\n", authorId, displayName, content)

	// Ignore all deletions of the bot's messages
	if authorId == s.State.User.ID {
		log.Println(authorId + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	// Every other deletion deducts points from the author!
	workflows.DeleteMessageLosesPoints(s, m, authorId)
}
