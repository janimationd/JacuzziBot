package utils

import (
	"fmt"

	"github.com/janimationd/JacuzziBot/constants"
)

// Construct a help string
func Help() string {
	var result string
	result += fmt.Sprintf("Hi! I'm **%s**, your humble Discord bot. Here are the things I can do:\n", constants.BotName)

	// List all methods of earning points here
	result += "- You earn points via the following actions:\n"
	result += fmt.Sprintf("  - Posting messages: **+%s points**\n",
		FormatUIFloat(constants.MessagePoints))
	result += fmt.Sprintf("  - Reacting to messages: **+%s points** for you *and the message author*\n",
		FormatUIFloat(constants.ReactionPoints))
	result += fmt.Sprintf("  - Being in a voice call: **+%s points** per minute *times how many people are in the call*\n",
		FormatUIFloat(constants.VoiceCallPointsPerParticipantPerMinute))

	// List all slash commands here
	result += "- Here are the slash commands you can perform:\n"
	result += "  - `/help` - Show this help text (also shown to the channel whenever you mention me)\n"
	result += "  - `/points [user]` - Check how many points someone has (omit the `user` to chickity-check yourself)\n"
	result += "  - `/give <recipient> <amount> [message]` - Give someone else some of your points, optionally with a message\n"
	result += "  - `/set-timezone <region> <city>` - Set your local timezone, which is used by other features\n"

	result += "- Here are some other bot features with separate help pages:\n"
	result += "  - `/tama-help` - Info on the Tama minigame where you hatch creatures from eggs and care for them\n"

	result += "\nI'll get more ways for you to earn and spend points over time, so look forward to that!\n"
	result += "\nTry to keep point farming to a minumum :wink:"

	return result
}

func TamaHelp() string {
	var result string

	result += "Tama is a minigame where you buy eggs, care for them until they hatch, and then feed and " +
		"play with them to keep them happy and productive. Your Tama pets will interact with other people's pets " +
		"and develop relationships with them. If Tamas like each other enough, they may fall in love and then mate! " +
		"Happy Tamas can earn you points, though you'll have to spend points to purchase eggs and keep them fed. " +
		"Basically, it's (legally distinct) Tamagotchi that can earn you points in the long run!\n"

	result += "- Here are the commands you can play the minigame with:\n"
	result += fmt.Sprintf("  - `/buy-tama-egg` - Purchase a Tama egg for **%s point%s**. "+
		"*You need to call `/set-timezone` before running this command.*\n",
		FormatUIFloat(constants.TamaEggPurchaseCost), Plural(constants.TamaEggPurchaseCost))

	return result
}
