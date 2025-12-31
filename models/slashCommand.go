package models

import (
	"github.com/bwmarrin/discordgo"
)

type SlashCommand struct {
	// The Command that the user will execute. Can have a bunch of settings, including "Options"
	// which are the Command parameters, e.g. "/test-Command option1 [option2] [option3]".
	Command *discordgo.ApplicationCommand

	// The Handler, which is our code that's executed when the command is executed by a user.
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
