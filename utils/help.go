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

	result += "\nI'll get more ways for you to earn and spend points over time, so look forward to that!\n"
	result += "\nTry to keep point farming to a minumum :wink:"
	return result
}
