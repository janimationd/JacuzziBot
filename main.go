package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/discord"
)

func waitForSignal() {
	log.Printf("%s is now running. Press CTRL-C to exit.", constants.BotName)

	// Wait here until CTRL-C or other termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Entry point for the program
func main() {
	err := discord.Open()
	if err != nil {
		log.Println("Failed to open Discord connection:", err)
		return
	}
	defer discord.Close()

	waitForSignal()
	log.Printf("%s shutting down...", constants.BotName)
}
