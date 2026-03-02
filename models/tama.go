package models

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/utils"
)

type PositiveTrait uint16
type NegativeTrait uint16
type Hunger = uint8
type Mood = int8
type RelationshipScore = int8

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
	// How hungry the Tama is. Starts at 0, and increases by 1 at midnight each day. Feeding reduces it by 1.
	// Before its hunger increases each day, its hunger at the end of the last day is used to decrease its mood.
	Hunger Hunger
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
	// How many times the egg has been cared for. Once this reaches eggCareHatchThreshold the egg will hatch.
	EggCareCount uint8
	// The last time the Tama/egg was cared for in seconds since Unix epoch.
	LastCareTime int64
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

// Returns if this is a fully hatched pet already.
func (this *Tama) HasHatched() bool {
	return !this.IsEgg()
}

// Get the Name and Id of the Tama (if it haws a name) or just its Id.
func (this *Tama) GetNameAndId() string {
	if this.Name != "" {
		return fmt.Sprintf("%s (ID %d)", this.Name, this.Id)
	} else {
		return fmt.Sprint(this.Id)
	}
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
		moodDesc = "glowing"
	}

	if moodDesc == "" {
		// Shouldn't get here
		log.Printf("Tama %d had an invalid mood value %d.\n", this.Id, this.Mood)
		if this.Mood < -moodLimit {
			moodDesc = "dead"
		} else if this.Mood > moodLimit {
			moodDesc = "glowing"
		} else {
			moodDesc = "confused"
		}
	}

	return fmt.Sprintf("%s (mood %s%d)", moodDesc, utils.SignString(float64(this.Mood)), this.Mood)
}

// Get a string status message for this Tama.
func (this *Tama) StatusMessage() string {
	result := ""
	if this.IsEgg() {
		careCountBeforeHatching := constants.EggCareHatchThreshold - this.EggCareCount
		result = fmt.Sprintf("Egg %d has been cared for %d times (needs %d more to hatch)",
			this.Id, this.EggCareCount, careCountBeforeHatching)
	} else {
		result = fmt.Sprintf("Tama %s is %s", this.GetNameAndId(), this.GetMoodString())
		if this.IsAlive() {
			age := time.Since(time.Unix(this.HatchedTime, 0))
			result += fmt.Sprintf(" and %s old", utils.FormatUIDuration(age))
		}
	}
	return result
}

func randomlyChooseNOptions[T utils.Integer](n uint, max T) *utils.Set[T] {
	values := make([]T, max)
	for i := T(0); i < max; i++ {
		values[i] = i
	}

	rand.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})

	result := &utils.Set[T]{}
	for i := uint(0); i < n; i++ {
		result.Add(values[i])
	}

	return result
}

// Hatch!
func (this *Tama) Hatch() {
	this.HatchedTime = time.Now().Unix()
	this.PositiveTraits = randomlyChooseNOptions(2, PositiveTraitMax)
	this.NegativeTraits = randomlyChooseNOptions(1, NegativeTraitMax)
}
