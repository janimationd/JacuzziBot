package discord

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/db"
)

// Message handler
func MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.Message.Author.Username + ": " + m.Message.Content)

	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		fmt.Println(m.Author.ID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	user, err := db.AwardUserPoints(m.GuildID, m.Author.ID, 1)
	if err != nil {
		fmt.Println("Error awarding points to user:", err)
		return
	}

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "```\n"+string(jsonBytes)+"\n```")
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
