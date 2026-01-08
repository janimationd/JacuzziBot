package models

import "github.com/janimationd/JacuzziBot/utils"

type PositiveTrait uint64
type NegativeTrait uint64
type Mood = int8
type RelationshipScore = int8

const (
	Friendly        PositiveTrait = 0
	SocialButterfly PositiveTrait = 1
	Sympathetic     PositiveTrait = 2
	Fertile         PositiveTrait = 3
)

const (
	Bully    NegativeTrait = 0
	Jealous  NegativeTrait = 1
	Annoying NegativeTrait = 2
)

// A Tama pet
type Tama struct {
	// The server-unique JacuzziId of the Tama
	Id JacuzziId
	// True if the Tama hasn't hatched yet, and is still an egg. If true, all later fields can/will be default values.
	IsEgg bool
	// The user-configured name of the Tama after it has hatched.
	Name string
	// The mood of the Tama pet after it has hatched.
	Mood Mood
	// How much this Tama likes other Tamas it has interacted with.
	Relationships map[JacuzziId]RelationshipScore
	// The positive traits of the Tama after it has hatched.
	PositiveTraits *utils.Set[PositiveTrait]
	// The negative traits of the Tama after it has hatched.
	NegativeTraits *utils.Set[NegativeTrait]
}
