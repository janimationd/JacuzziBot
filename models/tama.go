package models

import (
	"github.com/janimationd/JacuzziBot/utils"
)

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

const moodLimit Mood = 10
const relationshipScoreLimit RelationshipScore = 5

// A Tama pet
type Tama struct {
	// The server-unique JacuzziId of the Tama
	Id JacuzziId
	// The user ID of the owner. If "" this egg is unclaimed.
	Owner string
	// The user-configured name of the Tama after it has hatched.
	Name string
	// The mood of the Tama pet after it has hatched. Range [-10, 10], -10 = dead.
	Mood Mood
	// How much this Tama likes other Tamas it has interacted with. Range [-5, 5].
	Relationships map[JacuzziId]RelationshipScore
	// The positive traits of the Tama after it has hatched.
	PositiveTraits *utils.Set[PositiveTrait]
	// The negative traits of the Tama after it has hatched.
	NegativeTraits *utils.Set[NegativeTrait]
	// The parents of this Tama. There should only be 2 but we model this as a set for the utility it provides.
	Parents *utils.Set[JacuzziId]
	// The children of this Tama.
	Children *utils.Set[JacuzziId]
	// When the egg was laid in seconds since Unix epoch.
	EggLaidTime int64
	// When the egg was hatched in seconds since Unix epoch. We consider this the Tama's "birth" time too.
	HatchedTime int64
}

// Whether the Tama is alive or dead.
func (this *Tama) IsAlive() bool {
	return this.Mood > -moodLimit
}

// Whether this Tama is related to another Tama ID.
func (this *Tama) IsRelatedTo(otherId JacuzziId) bool {
	return this.Parents.Contains(otherId) || this.Children.Contains(otherId)
}

// Modify the Tama's mood by a delta.
func (this *Tama) ModifyMood(delta Mood) {
	this.Mood = utils.Clamp(this.Mood+delta, -moodLimit, moodLimit)
}

// Modify the Tama's relation score towards another Tama by a delta.
func (this *Tama) ModifyRelationshipScoreWith(otherId JacuzziId, delta RelationshipScore) {
	this.Relationships[otherId] =
		utils.Clamp(this.Relationships[otherId]+delta, -relationshipScoreLimit, relationshipScoreLimit)
}

// Whether the egg is claimed/owned or not.
func (this *Tama) IsOwned() bool {
	return this.Owner != ""
}

// Returns if this is an egg that hasn't hatched into a full pet yet.
func (this *Tama) IsEgg() bool {
	return this.HatchedTime == 0
}

// Returns if this is fully hatched pet already.
func (this *Tama) HasHatched() bool {
	return !this.IsEgg()
}
