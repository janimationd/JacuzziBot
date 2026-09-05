package models

type FlairLevel int

const (
	// No flair
	FlairNone FlairLevel = iota
	// Flair level 1
	Adherent
	// Flair level 2
	Devout
	// Flair level 3
	Fanatic
	// Flair level 4
	Zealot
	// Flair level 5
	Acolyte
	// Used for array limits, etc.
	FlairMax
)

// The properties structure of a flair level
type Flair struct {
	Name           string
	TotalPointCost float64
	ColorName      string
	ColorCode      int
	ColorEmoji     string
	RoleName       string
}

// The property sets of each flair level
var FlairProps = [FlairMax]Flair{
	{"No Flair", 0, "No Color", 0x0, "", ""},
	{"Adherent", 100, "Green", 0x12A312, " :green_square:", "Adherent (Flair Level 1)"},
	{"Devout", 500, "Blue", 0x4747EC, " :blue_square:", "Devout (Flair Level 2)"},
	{"Fanatic", 2000, "Red", 0xE10C0C, " :red_square:", "Fanatic (Flair Level 3)"},
	{"Zealot", 10000, "Purple", 0xAA22F8, " :purple_square:", "Zealot (Flair Level 4)"},
	{"Acolyte", 50000, "Gold", 0xF8D300, " :orange_square:", "Acolyte (Flair Level 5)"},
}

// All information we track that's associated with a user who can interact with the bot.
// There's other info we use that we explicitly *don't* track here, such as display names and pronouns,
// since they can change and it would be bad if we cached the wrong thing.
type User struct {
	// For Discord, this is a value like "123456789012345678".
	UserId string
	// Points are a float behind the scenes, but we should limit precision when showing to users.
	Points float64
	// The user's timezone. Stored in the format of "America/Los_Angeles", so something that can be fed into
	// time.LoadLocation().
	Timezone string
	// The user's purchased flair level (cosmetic rank)
	Flair FlairLevel
}
