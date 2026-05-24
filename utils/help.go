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
	result += "  - `/gamble <chance> <wager>` - Wager some of your points, gambling to maybe win back more\n"

	result += "- Here are some other bot features with separate help pages:\n"
	result += "  - `/tama-help` - Info on the Tamas minigame where you hatch pets from eggs and care for them\n"

	return result
}

func TamaHelp() string {
	var result string

	result += "Tamas is a minigame where you buy eggs, care for them until they hatch, and then feed and " +
		"play with them to keep them happy and productive. Your Tama pets will interact with other people's pets " +
		"and develop relationships with them. If Tamas like each other enough, they may fall in love! " +
		"If you neglect your pets, **they will die at -10 mood!** " +
		"Other pets will react to a pet dying according to their attitude towards it. " +
		"On top of all that, very happy Tamas will earn you points (more than the cost of caring for them).\n\n"

	result += "Here are the commands you can use to play the minigame " +
		"**(You need to call `/set-timezone` once** before running any of the following commands):\n"
	result += fmt.Sprintf("- `/buy-tama-egg` - Purchase a Tama egg for **%s point%s**.\n",
		FormatUIFloat(constants.TamaEggPurchaseCost), Plural(constants.TamaEggPurchaseCost))
	result += "- `/name-tama` - Name a Tama. By default Tama pets are referred to by their ID numbers, " +
		"but you can assign them a name after they hatch instead.\n"
	result += fmt.Sprintf("- `/care-tama` - Care for a Tama. "+
		"If it hasn't hatched yet, it gets closer to hatching (%s cooldown). "+
		"If it has already hatched, you play with the Tama and improve its mood (%s cooldown).\n",
		FormatUIDuration(constants.EggCareCooldown), FormatUIDuration(constants.TamaCareCooldown))
	result += fmt.Sprintf("- `/feed-tama` - Buy 1 food **for %s point%s** and feed a Tama. "+
		"You need to feed each Tama you own every calendar day (your local time). "+
		"Every day that you don't, hunger increases by 1. "+
		"Each day that a Tama is hungry, it loses mood equal to its hunger. "+
		"Forgetting to feed your Tamas is the easiest way to kill them! :skull:\n",
		FormatUIFloat(constants.TamaFeedCost), Plural(constants.TamaFeedCost))
	result += "- `/check-tama` - See the current status of one or all of your Tamas.\n"
	// result += fmt.Sprintf("- `/claim-tama-egg <id>` - Claim an unclaimed Tama egg. "+
	// 	"Newly hatched eggs can only be claimed by the owners of their parent Tamas for %d day%s.\n",
	// 	constants.OnlyParentOwnersCanClaimDays, Plural(constants.OnlyParentOwnersCanClaimDays))
	result += "- `/transfer-tama` - Transfer a Tama to another user. " +
		"Useful if you're going on vacation and won't be able to care for your Tamas temporarily, " +
		"or otherwise want to stop playing.\n"

	return result
}
