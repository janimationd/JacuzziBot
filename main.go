package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/discord"
)

func main() {
	// Load bot config/auth details
	// SENSITIVE!!!
	cfg, err := discord.LoadAuth()
	if err != nil {
		log.Println("Error loading auth:", err)
		return
	}

	token := cfg.Token

	// Create Discord session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return
	}

	// Register message handler
	dg.AddHandler(discord.MessageCreateHandler)

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")

	// Wait here until CTRL-C or other termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close Discord session
	dg.Close()
}
