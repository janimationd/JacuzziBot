package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/discord/handlers"
	"github.com/janimationd/JacuzziBot/discord/slashCommands"
	"github.com/janimationd/JacuzziBot/models"
)

func add(m *map[string]*models.SlashCommand, slashCommand *models.SlashCommand) {
	(*m)[slashCommand.Command.Name] = slashCommand
}

var commands = map[string]*models.SlashCommand{}

var registeredCommands []*models.SlashCommand

// This is going to be significantly reworked by code changes Araceli is working on, so I'm not worried about making
// this super scalable right now.
func registerInteractionCreateHandlers(session *discordgo.Session) {
	// There are many kinds of interactions, so we subscribe to specific subtypes here.
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Interaction.Type == discordgo.InteractionApplicationCommandAutocomplete {
			switch i.ApplicationCommandData().Name {
			case slashCommands.SetTimezone.Command.Name:
				slashCommands.TimezoneCityAutoComplete(s, i)
			}
		} else {
			if h, ok := commands[i.ApplicationCommandData().Name]; ok {
				log.Printf("Matched incoming interaction to handler: %s.\n", i.ApplicationCommandData().Name)
				h.Handler(s, i)
			} else {
				log.Printf("Couldn't match incoming interaction to a handler: %s.\n", i.ApplicationCommandData().Name)
			}
		}
	})
	log.Println("Registered command handler.")
}

func registerSlashCommands(session *discordgo.Session) {
	// List of slash commands to register
	add(&commands, &slashCommands.Help)
	add(&commands, &slashCommands.Points)
	add(&commands, &slashCommands.Give)
	add(&commands, &slashCommands.SetTimezone)
	add(&commands, &slashCommands.Gamble)

	// Register the list of commands
	for _, slashCommand := range commands {
		cmd, err := session.ApplicationCommandCreate(
			session.State.User.ID,
			// Passing this makes them guild-level commands which are client-side refreshed much faster than global ones.
			session.State.Application.GuildID,
			slashCommand.Command,
		)
		if err != nil {
			log.Println("Cannot register slash command:", err)
			continue
		}
		slashCommand.Command = cmd
		registeredCommands = append(registeredCommands, slashCommand)
		log.Printf("Registered command with name \"%s\".\n", slashCommand.Command.Name)
	}

	log.Printf("Registered %d commands.\n", len(registeredCommands))

	registerInteractionCreateHandlers(session)
}

func deregisterSlashCommands(session *discordgo.Session) {
	for _, slashCommand := range registeredCommands {
		if slashCommand != nil {
			err := session.ApplicationCommandDelete(session.State.User.ID, "", slashCommand.Command.ID)
			if err != nil {
				log.Printf("Couldn't deregister command with name \"%s\" (%s): %s\n",
					slashCommand.Command.Name, slashCommand.Command.ID, err.Error())
				continue
			}
			log.Printf("Deregistered command with name \"%s\" (%s).\n",
				slashCommand.Command.Name, slashCommand.Command.ID)
		}
	}

	// If you need to cleanup any leftover command registrations, do so here.
	// You can copy the command ID from a previous run in the server's settings under "Integrations",
	// then right click on the slash command and copy its ID.
	//session.ApplicationCommandDelete(session.State.User.ID, "", "1455810415937458304")
}

var Session *discordgo.Session

// Setup the Discord API listener and callbacks for handling various incoming events.
func Open() error {
	// Load bot config/auth details
	// SENSITIVE!!!
	auth, err := LoadAuth()
	if err != nil {
		log.Println("Error loading auth:", err)
		return err
	}

	// Create Discord session
	Session, err = discordgo.New("Bot " + auth.Token)
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return err
	}

	// Register handlers
	Session.AddHandler(handlers.MessageCreateHandler)
	Session.AddHandler(handlers.ReactionHandler)
	Session.AddHandler(handlers.VoiceCallHandler)

	// Open connection
	err = Session.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return err
	}

	registerSlashCommands(Session)

	log.Println("Discord session opened and waiting for events.")
	return nil
}

func Close() {
	deregisterSlashCommands(Session)

	// Cleanly close Discord session
	err := Session.Close()
	if err != nil {
		log.Println("Failed to close Discord session:", err)
		return
	}

	log.Println("Discord session closed.")
}
