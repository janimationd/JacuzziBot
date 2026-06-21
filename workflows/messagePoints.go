package workflows

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
)

func CreateMessageGetsPoints(s *discordgo.Session, m *discordgo.MessageCreate) {
	amount := constants.MessagePoints

	_, err := db.ModifyUserPoints(m.GuildID, m.Author.ID, amount)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}
}

func DeleteMessageLosesPoints(s *discordgo.Session, m *discordgo.MessageDelete, authorId string) {
	amount := -constants.MessagePoints

	_, err := db.ModifyUserPointsWithDebt(m.GuildID, authorId, amount, true)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}
}
