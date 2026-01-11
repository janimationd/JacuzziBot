package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/discord"
	"github.com/janimationd/JacuzziBot/scheduler"
)

// Entry point for the program
func main() {
	// Run all tests we have on startup.
	RunTestsIfRequested()

	// Setup Discord session
	err := discord.Open()
	if err != nil {
		log.Println("Failed to open Discord session:", err)
		return
	}
	defer discord.Close()

	// Execute goroutines using a WaitGroup so we can wait for them to finish later when SIGINT has been received.
	var waitGroup sync.WaitGroup
	// Scheduler goroutine
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	waitGroup.Go(func() {
		scheduler.Run(ctx)
	})

	// Wait here until CTRL-C or other termination signal
	log.Printf("%s is now running. Press CTRL-C to exit.", constants.BotName)
	<-ctx.Done()

	log.Printf("%s shutting down...", constants.BotName)
	waitGroup.Wait()
	log.Println("All goroutines have completed; exiting.")
}
