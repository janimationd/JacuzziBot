package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/discord/handlers"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/discord/slashCommands"
	"github.com/janimationd/JacuzziBot/discord/slashCommands/tama"
	"github.com/janimationd/JacuzziBot/models"
)

func add(m *map[string]*models.SlashCommand, slashCommand *models.SlashCommand) {
	(*m)[slashCommand.Command.Name] = slashCommand
}

var commands = map[string]*models.SlashCommand{}

var registeredCommands []*models.SlashCommand

// This is going to be significantly reworked by code changes Araceli is working on, so I'm not worried about making
// this super scalable right now.
func registerInteractionCreateHandlers() {
	// There are many kinds of interactions, so we subscribe to specific subtypes here.
	session.Handle.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Interaction.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commands[i.ApplicationCommandData().Name]; ok {
				log.Printf("Matched incoming interaction to handler: %s.\n", i.ApplicationCommandData().Name)
				h.Handler(s, i)
			} else {
				log.Printf("Couldn't match incoming interaction to a handler: %s.\n", i.ApplicationCommandData().Name)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			switch i.ApplicationCommandData().Name {
			case slashCommands.SetTimezone.Command.Name:
				slashCommands.TimezoneCityAutoComplete(s, i)
			}
		case discordgo.InteractionMessageComponent:
			customId := i.MessageComponentData().CustomID
			switch customId {
			case tama.BuyTamaEggConfirmPurchaseId:
				tama.HandleBuyTamaEggConfirmPurchase(s, i)
			case tama.BuyTamaEggCancelPurchaseId:
				tama.HandleBuyTamaEggCancelPurchase(s, i)
			case tama.SellTamaCancelSaleId:
				tama.HandleSellTamaCancelSale(s, i)
			// Handle custom IDs that have information appended to them
			default:
				if strings.HasPrefix(customId, tama.TransferTamaAcceptTransferId) {
					tama.HandleTransferTamaAcceptTransfer(s, i)
				} else if strings.HasPrefix(customId, tama.TransferTamaCancelTransferId) {
					tama.HandleTransferTamaCancelTransfer(s, i)
				} else if strings.HasPrefix(customId, "Prediction") {
					handlers.PredictionInteractionHandler(s, i)
				} else if strings.HasPrefix(customId, tama.SellTamaConfirmSaleId) {
					tama.HandleSellTamaConfirmSale(s, i)
				}
			}
		case discordgo.InteractionModalSubmit:
			customId := i.ModalSubmitData().CustomID
			switch customId {
			case "PredictionCreateModal":
				handlers.PredicationCreateSubmitHandler(s, i)
			default:
				handlers.PredictionInteractionHandler(s, i)
			}
		}
	})
	log.Println("Registered command handler.")
}

func registerSlashCommands() {
	// List of slash commands to register
	add(&commands, &slashCommands.Help)
	add(&commands, &slashCommands.Points)
	add(&commands, &slashCommands.Give)
	add(&commands, &slashCommands.SetTimezone)
	add(&commands, &slashCommands.Gamble)
	add(&commands, &slashCommands.Award)
	add(&commands, &slashCommands.CreatePrediction)
	add(&commands, &slashCommands.Roll)

	// Tama minigame commands
	add(&commands, &tama.TamaHelp)
	add(&commands, &tama.RegisterTamaChannel)
	add(&commands, &tama.BuyTamaEgg)
	add(&commands, &tama.SellTama)
	//add(&commands, &tama.ClaimTamaEgg)
	add(&commands, &tama.TransferTama)
	add(&commands, &tama.NameTama)
	add(&commands, &tama.CheckTama)
	add(&commands, &tama.CareTama)
	add(&commands, &tama.FeedTama)
	add(&commands, &tama.ScheduleTamaPlaytime)
	add(&commands, &tama.ListTamaPlaytimes)
	add(&commands, &tama.CancelTamaPlaytime)
	add(&commands, &tama.BackupTamas)
	add(&commands, &tama.RestoreTamas)
	add(&commands, &tama.WipeTamas)

	// Build the list of command definitions to sync
	var commandDefs []*discordgo.ApplicationCommand
	for _, slashCommand := range commands {
		commandDefs = append(commandDefs, slashCommand.Command)
	}

	// Single API call to sync all commands at once
	updatedCmds, err := session.Handle.ApplicationCommandBulkOverwrite(
		session.Handle.State.User.ID,
		session.Handle.State.Application.GuildID,
		commandDefs,
	)

	if err != nil {
		log.Println("Failed to bulk overwrite slash commands:", err)
	} else {
		// Update local commands map with the IDs Discord assigned
		for _, cmd := range updatedCmds {
			if sc, ok := commands[cmd.Name]; ok {
				sc.Command = cmd
				registeredCommands = append(registeredCommands, sc)
			}
		}
		log.Printf("Synced %d commands.\n", len(updatedCmds))
	}

	registerInteractionCreateHandlers()
}

func deregisterSlashCommands() {
	for _, slashCommand := range registeredCommands {
		if slashCommand != nil {
			err := session.Handle.ApplicationCommandDelete(session.Handle.State.User.ID, "", slashCommand.Command.ID)
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
	session.Handle, err = discordgo.New("Bot " + auth.Token)
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return err
	}

	// Configure session to allow caching of message details.
	// We only do this to allow us to deduct points from users who delete their messages, otherwise
	// we can't know the author of a deleted message.
	session.Handle.State.MaxMessageCount = 200
	session.Handle.Identify.Intents =
		session.Handle.Identify.Intents | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// Register handlers
	session.Handle.AddHandler(handlers.MessageCreateHandler)
	session.Handle.AddHandler(handlers.MessageDeleteHandler)
	session.Handle.AddHandler(handlers.AddReactionHandler)
	session.Handle.AddHandler(handlers.RemoveReactionHandler)
	session.Handle.AddHandler(handlers.VoiceCallHandler)

	// Open connection
	err = session.Handle.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return err
	}
	err = session.Handle.UpdateCustomStatus("♨️ The water's fine")
	if err != nil {
		log.Printf("Error updating bot's status: %s\n", err.Error())
	}

	registerSlashCommands()

	log.Println("Discord session opened and waiting for events.")
	return nil
}

func Close() {
	// We don't deregister on shutdown anymore, but keeping this code here in case we ever need to.
	//deregisterSlashCommands()

	// Cleanly close Discord session
	err := session.Handle.Close()
	if err != nil {
		log.Println("Failed to close Discord session:", err)
		return
	}

	log.Println("Discord session closed.")
}
