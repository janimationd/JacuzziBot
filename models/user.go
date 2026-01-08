package models

import "github.com/janimationd/JacuzziBot/utils"

// All information we track that's associated with a user who can interact with the bot.
// There's other info we use that we explicitly *don't* track here, such as display names and pronouns,
// since they can change and it would be bad if we cached the wrong thing.
type User struct {
	// For Discord, this is a value like "123456789012345678".
	UserId string
	// Points are a float behind the scenes, but we should limit precision when showing to users.
	Points float64
	// The IDs of all the Tama pets/eggs owned by the user.
	Tamas *utils.Set[JacuzziId]
	// How many total Tama the user has purchase before. This influences the cost to buy another one.
	NumTamasPurchased uint64
}
