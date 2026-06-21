package workflows

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
)

func AddReactionGetsPoints(s *discordgo.Session, r *discordgo.MessageReactionAdd, message *discordgo.Message) {
	amount := constants.ReactionPoints

	// Give the reactor points
	_, err := db.ModifyUserPoints(r.GuildID, r.Member.User.ID, amount)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}

	// Give the author of the message being reacted to points
	_, err = db.ModifyUserPoints(r.GuildID, message.Author.ID, amount)
	if err != nil {
		log.Println("Error awarding points to reactor:", err)
		return
	}
}

func RemoveReactionLosesPoints(s *discordgo.Session, r *discordgo.MessageReactionRemove, message *discordgo.Message) {
	amount := -constants.ReactionPoints

	// Remove the reactor's points
	_, err := db.ModifyUserPointsWithDebt(r.GuildID, r.UserID, amount, true)
	if err != nil {
		log.Println("Error removing points from user:", err)
		return
	}

	// Remove the author of the message being reacted to's points
	_, err = db.ModifyUserPointsWithDebt(r.GuildID, message.Author.ID, amount, true)
	if err != nil {
		log.Println("Error removing points from message author:", err)
		return
	}
}
