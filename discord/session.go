package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

// Setup the Discord API listener and callbacks for handling various incoming events.
func Open() error {
	// Load bot config/auth details
	// SENSITIVE!!!
	cfg, err := LoadAuth()
	if err != nil {
		log.Println("Error loading auth:", err)
		return err
	}

	token := cfg.Token

	// Create Discord session
	session, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return err
	}

	// Register message handler
	session.AddHandler(MessageCreateHandler)

	// Open connection
	err = session.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return err
	}

	log.Println("Discord session opened and waiting for events.")
	return nil
}

func Close() {
	// Cleanly close Discord session
	err := session.Close()
	if err != nil {
		log.Println("Failed to close Discord session:", err)
		return
	}

	log.Println("Discord session closed.")
}
