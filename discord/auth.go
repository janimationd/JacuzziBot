package discord

import (
	_ "embed"
	"fmt"
	"os"
)

type AuthConfig struct {
	Token string `json:"Token"`
}

// Load authentication config
func LoadAuth() (*AuthConfig, error) {
	var cfg AuthConfig

	// Override token with environment variable if set
	if os.Getenv("DISCORD_TOKEN") != "" {
		cfg.Token = os.Getenv("DISCORD_TOKEN")
	}

	// Make sure token was loaded
	if cfg.Token == "" {
		return nil, fmt.Errorf("Discord token is not set in auth.json or DISCORD_TOKEN environment variable")
	}

	return &cfg, nil
}
