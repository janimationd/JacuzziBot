package workflows

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"github.com/janimationd/JacuzziBot/utils"
)

func BotMentionPrintsHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSendReply(m.ChannelID, utils.Help(), m.Reference())
	if err != nil {
		log.Println("Failed to send help reply:", err)
		return
	}
}

func BotMentionPrintsTamaHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSendReply(m.ChannelID, utils.TamaHelp(), m.Reference())
	if err != nil {
		log.Println("Failed to send tama help reply:", err)
		return
	}
}
