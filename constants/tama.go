package constants

import "time"

// The duration in seconds after which two Tamas mate and lay an egg during which only the owners of the parent Tamas
// can claim the egg. After this duration expires, anyone can claim the egg.
const OnlyParentOwnersCanClaimDays = 3
const OnlyParentOwnersCanClaimSeconds int64 = OnlyParentOwnersCanClaimDays * 24 * 60 * 60

// Cannot name a Tama with more characters than this.
const MaxTamaNameLength = 64

// Characters that cannot be used in a Tama name.
const TamaNameDisallowedCharacters = "*_~#:\\[\\](){}`<>@\\/\\\\\\n\\r"
const TamaNameDisallowedCharactersForPrinting = "*_~#:[](){}`<>@/\\"

// After the egg has been cared for this many times, it hatches.
const EggCareHatchThreshold uint8 = 3

// The cooldown period for caring for an egg
const EggCareCooldown = 8 * time.Hour

// The cooldown period for caring for a hatched Tama
const TamaCareCooldown = 8 * time.Hour
