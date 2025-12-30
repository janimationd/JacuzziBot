package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/janimationd/JacuzziBot/discord"
)

func waitForSignal() {
	// Wait here until CTRL-C or other termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Entry point for the program
func main() {
	log.Println("JacuzziBot is now running. Press CTRL-C to exit.")
	discord.Open()
	defer discord.Close()

	waitForSignal()
	log.Println("JacuzziBot shutting down...")
}
