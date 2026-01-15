package main

import (
	"os"

	"github.com/janimationd/JacuzziBot/discord/slashCommands"
)

// Define unit tests here. We don't require unit tests for every change, but some might be useful still.
// Tests only run if the RUN_UNIT_TESTS environment variable is set.
func RunTestsIfRequested() {
	if os.Getenv("RUN_UNIT_TESTS") == "true" {
		// List all tests here
		slashCommands.TestTimezones()
	}
}
