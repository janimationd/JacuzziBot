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
	Fertile
	// New ones should go above this
	PositiveTraitMax
)

const (
	Bully NegativeTrait = iota
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
const TamaRelationshipScoreLimit RelationshipScore = 5

// These are the mood thresholds at or above which Tamas will earn their owner points.
const TamaMinorPointActionMoodThreshold = TamaMoodLimit / 2
const TamaMajorPointActionMoodThreshold = TamaMoodLimit

// A Tama pet
type Tama struct {
	// The server-unique JacuzziId of the Tama
	Id JacuzziId
	// The user ID of the owner. If "" this egg is unclaimed.
	Owner string
	// The user-configured name of the Tama after it has hatched.
	Name string
	// The last time the pet was fed by its owner. When first hatched, this will be its hatched time, so the owner
	// doesn't have to feed it that day.
	LastFeedTime int64
	// The hunger value at the LastFeedTime. Updates every time the pet is fed.
	LastHunger Hunger
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

// Count the amount of midnights between from and to in the given timezone.
func midnightsBetween(from, to time.Time, loc *time.Location) int {
	count := -1
	current := from
	for !current.After(to) {
		// Step exactly one calendar day forward in local time,
		// letting time.Date handle any DST normalization.
		y, m, d := current.Date()
		current = time.Date(y, m, d+1, 0, 0, 0, 0, loc)
		count++
	}
	return count
}

// How hungry the Tama is. Starts at 0, and increases by 1 at midnight each day. Feeding reduces it by 1.
// Before its hunger increases each day, its hunger at the end of the last day is used to decrease its mood.
// Must pass in the owner timezone so we can calculate the hunger correctly.
func (this *Tama) Hunger(timezone *time.Location) Hunger {
	nowLocal := time.Now().In(timezone)
	lastFeedTimeLocal := time.Unix(this.LastFeedTime, 0).In(timezone)

	// Count midnights that have elapsed since LastFeedTime by comparing calendar dates in the owner's local timezone.
	midnightsSinceLastFeed := midnightsBetween(lastFeedTimeLocal, nowLocal, timezone)

	return this.LastHunger + Hunger(midnightsSinceLastFeed)
}

// Feed the Tama one food.
func (this *Tama) Feed(timezone *time.Location) {
	currentHunger := this.Hunger(timezone)
	this.LastFeedTime = time.Now().Unix()
	this.LastHunger = Hunger(max(0, int(currentHunger)-1))
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

// The result of modifying a Tama's mood.
type ModifyMoodResult struct {
	// The actual amount the mood changed by. Could less than expected if capped by a limit.
	FinalDelta Mood
	// Is already dead
	WasDead bool
	// The Tama lost enough mood to die.
	JustDied bool
}

// Modify the Tama's mood by a delta.
func (this *Tama) ModifyMood(delta Mood) ModifyMoodResult {
	beforeMood := this.Mood
	// Dead Tamas can't be resurrected
	if this.IsDead() {
		return ModifyMoodResult{
			FinalDelta: 0,
			WasDead:    true,
			JustDied:   false,
		}
	}
	this.Mood = utils.Clamp(this.Mood+delta, -TamaMoodLimit, TamaMoodLimit)
	finalDelta := this.Mood - beforeMood
	return ModifyMoodResult{
		FinalDelta: finalDelta,
		WasDead:    false,
		JustDied:   this.IsDead(),
	}
}

// The result of a relationship score change w.r.t. love states.
type LoveResult int

const (
	// Nothing happened w.r.t. the pets' love state.
	NoChange LoveResult = iota
	// The pets fell in love!
	FellInLove
	// The pets are in love, and this prevented them from losing relationship score.
	LovePreventedDecrease
	// The pets fell out of love!
	FellOutOfLove
)

// The result of calling Tama.ModifyRelationshipScoreWith().
type RelationshipScoreModificationResult struct {
	// The amount the relationship score actually changed by (might be different than expected if it got clamped).
	FinalDelta RelationshipScore
	// The amount the Tama's mood might have been affected by in the same direction.
	MoodDelta Mood
	// Any relationship score bonus that was added from the Friendly trait triggerring.
	FriendlyBonus RelationshipScore
	// Any special results related to love state.
	LoveEvent LoveResult
	// True if the other Tama's Annoying trait blocked a Play-related increase.
	AnnoyingBlockedIncrease bool
	// Whether the Tama was already dead.
	WasDead bool
	// Whether the Tama just died as a result of a mood change.
	JustDied bool
}

// Modify the Tama's relationship score towards another Tama by a delta. If this is due to a TamaInteraction, pass it
// in for additional effects. If not, pass in TamaInteractionMax.
func (this *Tama) ModifyRelationshipScoreWith(
	other *Tama,
	delta RelationshipScore,
	interaction TamaInteraction,
) RelationshipScoreModificationResult {
	var loveEvent LoveResult = NoChange

	// Cannot modify relationships for a dead Tama
	if this.IsDead() {
		log.Printf("Cannot modify dead Tama %s's relationship score.\n", this.GetNameAndId())
		return RelationshipScoreModificationResult{
			WasDead: true,
		}
	}

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

	// If the tamas are in love, there's a 66% chance that a decrease in relationship score doesn't happen.
	if this.Loves(other) && delta < 0 && 0 != rand.IntN(3) {
		delta = 0
		loveEvent = LovePreventedDecrease
	}

	// Handle the effect of the Annoying trait (33% chance).
	annoyingBlockedIncrease := delta > 0 && interaction == Play && other.NegativeTraits.Contains(Annoying) &&
		0 == rand.IntN(3)
	if annoyingBlockedIncrease {
		delta = 0
	}

	// The maximum possible relationship score depends on whether the pets are direct children/parents of each other.
	maxScore := TamaRelationshipScoreLimit
	if this.IsRelatedTo(other.Id) {
		maxScore -= 1
	}

	// Update the relationship score
	this.Relationships[other.Id] =
		utils.Clamp(this.Relationships[other.Id]+delta, -TamaRelationshipScoreLimit, maxScore)

	finalDelta := this.Relationships[other.Id] - oldRelationshipScore

	// Anytime a relationship score changes, there's a 33% chance the Tama's mood changes by 1 in the same direction.
	moodDelta := Mood(0)
	if 0 == rand.IntN(3) {
		moodDelta = Mood(utils.Sign(delta))
		modifyMoodResult := this.ModifyMood(moodDelta)
		moodDelta = modifyMoodResult.FinalDelta
		if modifyMoodResult.JustDied {
			log.Printf("Tama %s just died as a result of losing mood from interacting with %s.\n",
				this.GetNameAndId(), other.GetNameAndId())
			return RelationshipScoreModificationResult{
				FinalDelta: finalDelta,
				MoodDelta:  moodDelta,
				JustDied:   true,
			}
		}
	}

	// Check for falling in love
	if this.Relationships[other.Id] == TamaRelationshipScoreLimit &&
		other.Relationships[this.Id] == TamaRelationshipScoreLimit &&
		this.LoveTarget == NoId && other.LoveTarget == NoId {
		this.LoveTarget = other.Id
		other.LoveTarget = this.Id
		loveEvent = FellInLove
	}

	// Check for falling out of love
	if (this.Relationships[other.Id] <= 0 ||
		other.Relationships[this.Id] <= 0) &&
		this.LoveTarget == other.Id && other.LoveTarget == this.Id {
		this.LoveTarget = NoId
		other.LoveTarget = NoId
		loveEvent = FellOutOfLove
	}

	// Build result
	return RelationshipScoreModificationResult{
		FinalDelta:              finalDelta,
		MoodDelta:               Mood(moodDelta),
		FriendlyBonus:           RelationshipScore(friendlyBonus),
		LoveEvent:               loveEvent,
		AnnoyingBlockedIncrease: annoyingBlockedIncrease,
	}
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
		return fmt.Sprintf("%s (#%d)", this.Name, this.Id)
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

	return time.Unix(this.LastCareTime, 0).Add(cooldown)
}

// The egg hatches!
func (this *Tama) Hatch() {
	this.HatchedTime = time.Now().Unix()
	this.LastFeedTime = this.HatchedTime

	// If this egg was a result of two other pets mating, it will already have its starting traits set to the union
	// of its parents' sets of traits. We must now choose from those pools.
	numPositiveTraits := 1
	numNegativeTraits := 1
	if this.Parents.Size() > 0 {
		this.PositiveTraits = this.PositiveTraits.ChooseRandomNFromSet(numPositiveTraits)
		this.NegativeTraits = this.NegativeTraits.ChooseRandomNFromSet(numNegativeTraits)
	} else {
		// Otherwise, generate some random fresh traits.
		this.PositiveTraits = utils.ChooseRandomNIntegers(numPositiveTraits, PositiveTraitMax)
		this.NegativeTraits = utils.ChooseRandomNIntegers(numNegativeTraits, NegativeTraitMax)
	}

	// Set initial relationship scores with parents to one less than "courting".
	for parent := range this.Parents.All() {
		this.Relationships[parent] = TamaRelationshipScoreLimit - 1
	}
}

func (this *Tama) GetHourlyPointAward() float64 {
	if this.Mood >= TamaMajorPointActionMoodThreshold {
		return constants.TamaHourlyMajorPointActionReward
	} else if this.Mood >= TamaMinorPointActionMoodThreshold {
		return constants.TamaHourlyMinorPointActionReward
	} else {
		return 0
	}
}

// Get a string describing the Tama's mood.
func (this *Tama) GetMoodString() string {
	return PreviewMoodString(this.Mood, this.Id)
}

// Preview the string description for a mood value.
func PreviewMoodString(mood Mood, id JacuzziId) string {
	moodDesc := ""
	switch mood {
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
		log.Printf("Tama #%d had an invalid mood value %d.\n", id, mood)
		if mood < -TamaMoodLimit {
			moodDesc = "dead"
		} else if mood > TamaMoodLimit {
			moodDesc = "**glowing**"
		} else {
			moodDesc = "confused"
		}
	}

	return fmt.Sprintf("%s (mood %s%d)", moodDesc, utils.SignString(mood), mood)
}

// Whether this Tama is courting another.
func (this *Tama) IsCourting(other *Tama) bool {
	return this.Relationships[other.Id] == TamaRelationshipScoreLimit && this.LoveTarget == NoId
}

// Whether this tama is in love with another.
func (this *Tama) Loves(other *Tama) bool {
	return this.LoveTarget == other.Id
}

// Calculate the Tama's age
func (this *Tama) Age() time.Duration {
	return time.Since(time.Unix(this.HatchedTime, 0))
}

// Calculate the Tama's sell value
func (this *Tama) SellValueAndEquation() (float64, string) {
	daysOld := this.Age().Hours() / 24
	sellValue := daysOld * (1 + (float64(this.Mood) / float64(TamaMoodLimit)))
	equation := fmt.Sprintf("%s = daysOld:%s * (1 + (mood:%d / maxMood:%d))",
		utils.FormatUIFloat(sellValue), utils.FormatUIFloat(daysOld), this.Mood, TamaMoodLimit)
	return sellValue, equation
}
