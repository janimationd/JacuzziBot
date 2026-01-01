package workflows

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
)

func ReactionGetsPoints(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	user, err := db.ModifyUserPoints(r.GuildID, r.Member.User.ID, constants.ReactionPoints)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}

	// Give the react-er points
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Println("Error getting reaction message details:", err)
		return
	}
	log.Printf("User %s (%s) gained %.2f points, for a total of %.2f points.\n",
		r.Member.User.ID, r.Member.User.DisplayName(), constants.ReactionPoints, user.Points)

	// Give the author of the message being reacted to points
	author, err := db.ModifyUserPoints(r.GuildID, message.Author.ID, constants.ReactionPoints)
	if err != nil {
		log.Println("Error awarding points to user:", err)
		return
	}
	log.Printf("Author %s (%s) gained %.2f points, for a total of %.2f points.\n",
		message.Author.ID, message.Author.Username, constants.ReactionPoints, author.Points)
}
