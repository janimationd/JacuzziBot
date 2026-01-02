package slashCommands

import "github.com/bwmarrin/discordgo"

func getCommandOption(
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
