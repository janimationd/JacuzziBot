package discord

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

type AuthConfig struct {
	AppId  string `json:"AppId"`
	PubKey string `json:"PubKey"`
	Token  string `json:"Token"`
}

//go:embed auth.json
var authJson []byte

// Load authentication config
func LoadAuth() (*AuthConfig, error) {
	var cfg AuthConfig

	// Load embedded JSON
	err := json.Unmarshal(authJson, &cfg)
	if err != nil {
		return nil, err
	}

	// Override token with environment variable if set
	if os.Getenv("DISCORD_TOKEN") != "" {
		cfg.Token = os.Getenv("DISCORD_TOKEN")
	}

	// Make sure token was loaded from somewhere
	if cfg.Token == "" {
		return nil, fmt.Errorf("Discord token is not set in auth.json or DISCORD_TOKEN environment variable")
	}

	return &cfg, nil
}
