package utils

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/discord/session"
)

func GetCommandOption(
	interaction *discordgo.InteractionCreate,
	optionName string,
) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range interaction.ApplicationCommandData().Options {
		if opt.Name == optionName {
			return opt
		}
	}
	return nil
}

const DiscordMessageRuneLimit = 2000

func SplitMessage(message string, maxRunes int) []string {
	var parts []string

	runes := []rune(message)

	for len(runes) > 0 {
		if len(runes) <= maxRunes {
			parts = append(parts, string(runes))
			break
		}

		splitAt := maxRunes
		for i := maxRunes - 1; i >= 0; i-- {
			if runes[i] == '\n' {
				splitAt = i
				break
			}
		}

		parts = append(parts, string(runes[:splitAt]))
		runes = runes[splitAt+1:] // skip the newline itself
	}

	return parts
}

func SendLongMessage(channelId string, message string) error {
	messages := SplitMessage(message, DiscordMessageRuneLimit)
	for _, msg := range messages {
		_, err := session.Handle.ChannelMessageSend(channelId, msg)
		if err != nil {
			log.Printf("Couldn't send (partial?) channel message: %s\n", err.Error())
			return err
		}
	}
	return nil
}
