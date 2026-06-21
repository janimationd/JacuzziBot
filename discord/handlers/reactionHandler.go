package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/workflows"
)

// Handler users reacting to each others' messages
func AddReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	log.Printf("Incoming reaction by %s (%s): :%s:\n",
		r.Member.User.ID, r.Member.User.DisplayName(), r.Emoji.Name)

	// Ignore all reactions from the bot itself
	if r.Member.User.ID == s.State.User.ID {
		log.Println(r.Member.User.ID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Printf("Couldn't lookup message being reacted to: Channel=%s, Message=%s\n", r.ChannelID, r.MessageID)
		return
	}

	// Ignore people reacting to their own messages
	if r.Member.User.ID == message.Author.ID {
		log.Printf("Ignoring user %s reacting to their own message\n", r.Member.User.ID)
		return
	}

	// Every other reaction gets the author and recipient some points!
	workflows.AddReactionGetsPoints(s, r, message)
}

// Handler users removing reactions from each others' messages
func RemoveReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	member, err := s.GuildMember(r.GuildID, r.UserID)
	if err != nil {
		log.Printf("Incoming reaction removal by %s: :%s:\n", r.UserID, r.Emoji.Name)
		log.Printf("Couldn't lookup member %s doing reaction removal\n", r.UserID)
	} else {
		log.Printf("Incoming reaction removal by %s (%s): :%s:\n", r.UserID, member.DisplayName(), r.Emoji.Name)
	}

	// Ignore all reaction removals from the bot itself
	if r.UserID == s.State.User.ID {
		log.Println(r.UserID + " == " + s.State.User.ID + " (ignoring)")
		return
	}

	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Printf("Couldn't lookup message being reacted to: Channel=%s, Message=%s\n", r.ChannelID, r.MessageID)
		return
	}

	// Ignore people removing reactions from their own messages
	if r.UserID == message.Author.ID {
		log.Printf("Ignoring user %s removing reaction from their own message\n", r.UserID)
		return
	}

	// Every other reaction removal subtracts points from the author and recipient!
	workflows.RemoveReactionLosesPoints(s, r, message)
}
