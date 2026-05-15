package session

import "github.com/bwmarrin/discordgo"

// The singleton Discord API session, used from many places (hence its own package).
var Handle *discordgo.Session
