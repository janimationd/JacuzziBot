package constants

// The maximum amount of Tama/eggs that one user can own at a time.
const TamaLimitPerUser = 10

// The duration in seconds after which two Tamas mate and lay an egg during which only the owners of the parent Tamas
// can claim the egg. After this duration expires, anyone can claim the egg.
const OnlyParentOwnersCanClaimDays = 3
const OnlyParentOwnersCanClaimSeconds int64 = OnlyParentOwnersCanClaimDays * 24 * 60 * 60
