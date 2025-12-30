package workflows

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
)

func NewMessageGetsPoints(s *discordgo.Session, m *discordgo.MessageCreate) {
	user, err := db.ModifyUserPoints(m.GuildID, m.Author.ID, constants.NewMessagePoints)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}

	log.Printf("User %s (%s) gained %.2f points, for a total of %.2f points.\n",
		m.Author.ID, m.Author.DisplayName(), constants.NewMessagePoints, user.Points)
}
