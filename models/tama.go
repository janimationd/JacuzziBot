package models

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/utils"
)

type PositiveTrait uint
type NegativeTrait uint
type TamaInteraction uint
type GiftOutcome uint
type Hunger uint
type Mood int
type RelationshipScore int

const (
	Friendly PositiveTrait = iota
	SocialButterfly
	Sympathetic
	Fertile
	// New ones should go above this
	PositiveTraitMax
)

const (
	Bully NegativeTrait = iota
	Jealous
	Annoying
	// New ones should go above this
	NegativeTraitMax
)

const (
	Play TamaInteraction = iota
	Gift
	PickOn
	// New ones should go above this
	TamaInteractionMax
)

const (
	Likes GiftOutcome = iota
	Indifferent
	Hates
	// New ones should go above this
	GiftOutcomeMax
)

const TamaMoodLimit Mood = 10
const relationshipScoreLimit RelationshipScore = 5

// A Tama pet
type Tama struct {
	// The server-unique JacuzziId of the Tama
	Id JacuzziId
	// The user ID of the owner. If "" this egg is unclaimed.
	Owner string
	// The user-configured name of the Tama after it has hatched.
	Name string
	// How hungry the Tama is. Starts at 0, and increases by 1 at midnight each day. Feeding reduces it by 1.
	// Before its hunger increases each day, its hunger at the end of the last day is used to decrease its mood.
	Hunger Hunger
	// The mood of the Tama pet after it has hatched. Range [-10, 10], -10 = dead.
	Mood Mood
	// How much this Tama likes other Tamas it has interacted with. Range [-5, 5].
	Relationships map[JacuzziId]RelationshipScore
	// The Tama this one is in love with. If it isn't in love with anyone right now, then this will be NoId.
	LoveTarget JacuzziId
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
	// How many times the egg has been cared for. Once this reaches eggCareHatchThreshold the egg will hatch.
	EggCareCount uint8
	// The last time the Tama/egg was cared for in seconds since Unix epoch.
	LastCareTime int64
	// The ID of the server where this Tama exists.
	ServerId string
}

// Whether the Tama is alive.
func (this *Tama) IsAlive() bool {
	return this.Mood > -TamaMoodLimit
}

// Whether the Tama is dead.
func (this *Tama) IsDead() bool {
	return !this.IsAlive()
}

// Whether this Tama is related to another Tama ID.
func (this *Tama) IsRelatedTo(otherId JacuzziId) bool {
	return this.Parents.Contains(otherId) || this.Children.Contains(otherId)
}

// Modify the Tama's mood by a delta. Returns true if the moof was modified, false if it was already at a limit.
func (this *Tama) ModifyMood(delta Mood) bool {
	beforeMood := this.Mood
	this.Mood = utils.Clamp(this.Mood+delta, -TamaMoodLimit, TamaMoodLimit)
	return beforeMood != this.Mood
}

// Modify the Tama's relation score towards another Tama by a delta. Returns:
// - first, the amount the relationship score actually changed by (might be different than expected if it got clamped)
// - second, the amount the Tama's mood might have been affected by in the same direction
// - third, any relationship score bonus that was added from the Friendly trait triggerring
func (this *Tama) ModifyRelationshipScoreWith(
	other *Tama,
	delta RelationshipScore,
) (RelationshipScore, Mood, RelationshipScore) {
	if this.Relationships == nil {
		this.Relationships = make(map[JacuzziId]RelationshipScore)
	}
	oldRelationshipScore := this.Relationships[other.Id]

	// Handle the effect of the Friendly trait (33% chance)
	friendly := delta > 0 && other.PositiveTraits.Contains(Friendly) && 0 == rand.IntN(3)
	friendlyBonus := 0
	if friendly {
		friendlyBonus = 1
		delta += RelationshipScore(friendlyBonus)
	}

	this.Relationships[other.Id] =
		utils.Clamp(this.Relationships[other.Id]+delta, -relationshipScoreLimit, relationshipScoreLimit)

	// Anytime a relationship score changes, there's a 33% chance the Tama's mood changes by 1 in the same direction.
	moodDelta := 0
	if 0 == rand.IntN(3) {
		moodDelta = utils.Sign(delta)
		this.ModifyMood(Mood(moodDelta))
	}

	relationshipScoreChange := this.Relationships[other.Id] - oldRelationshipScore
	return relationshipScoreChange, Mood(moodDelta), RelationshipScore(friendlyBonus)
}

// Whether the egg is claimed/owned or not.
func (this *Tama) IsOwned() bool {
	return this.Owner != ""
}

// Returns if this is an egg that hasn't hatched into a full pet yet.
func (this *Tama) IsEgg() bool {
	return this.HatchedTime == 0
}

// Returns if this is a fully hatched pet already.
func (this *Tama) HasHatched() bool {
	return !this.IsEgg()
}

// Get the Name and Id of the Tama (if it has a name) or just its Id otherwise.
func (this *Tama) GetNameAndId() string {
	if this.Name != "" {
		return fmt.Sprintf("\"%s\" (#%d)", this.Name, this.Id)
	} else {
		return fmt.Sprintf("#%d", this.Id)
	}
}

// Return the next time the Tama is allowed to have `/care-tama` used on it at.
func (this *Tama) GetNextCareTime() time.Time {
	// Figure out the cooldown we need to use
	var cooldown time.Duration
	if this.IsEgg() {
		cooldown = constants.EggCareCooldown
	} else {
		cooldown = constants.TamaCareCooldown
	}

	// If the cooldown hasn't yet expired
	return time.Unix(this.LastCareTime, 0).Add(cooldown)
}

// Hatch!
func (this *Tama) Hatch() {
	this.HatchedTime = time.Now().Unix()
	this.PositiveTraits = utils.ChooseRandomNIntegers(2, PositiveTraitMax)
	this.NegativeTraits = utils.ChooseRandomNIntegers(1, NegativeTraitMax)
}

// Get a string describing the Tama's mood.
func (this *Tama) GetMoodString() string {
	moodDesc := ""
	switch this.Mood {
	case -10:
		moodDesc = "dead"
	case -9, -8:
		moodDesc = "dying"
	case -7, -6, -5:
		moodDesc = "depressed"
	case -4, -3:
		moodDesc = "upset"
	case -2, -1:
		moodDesc = "sad"
	//////////////////
	case 0:
		moodDesc = "bored"
	//////////////////
	case 1, 2:
		moodDesc = "content"
	case 3, 4:
		moodDesc = "happy"
	case 5, 6, 7:
		moodDesc = "excited"
	case 8, 9:
		moodDesc = "ecstatic"
	case 10:
		moodDesc = "**glowing**"
	}

	if moodDesc == "" {
		// Shouldn't get here
		log.Printf("Tama #%d had an invalid mood value %d.\n", this.Id, this.Mood)
		if this.Mood < -TamaMoodLimit {
			moodDesc = "dead"
		} else if this.Mood > TamaMoodLimit {
			moodDesc = "**glowing**"
		} else {
			moodDesc = "confused"
		}
	}

	return fmt.Sprintf("%s (mood %s%d)", moodDesc, utils.SignString(this.Mood), this.Mood)
}
