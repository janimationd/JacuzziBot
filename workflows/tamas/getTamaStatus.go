package tamas

import (
	"fmt"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

// Get a string describing the Tama's hunger state. String will end with a period.
func getHungerString(this *models.Tama, ownerTimezone *time.Location) string {
	hunger := this.Hunger(ownerTimezone)
	switch {
	case hunger == 0:
		return "well fed and doesn't need any more food until tomorrow."
	case hunger == 1:
		return fmt.Sprintf("hungry. `/feed-tama %d` it today to prevent a mood penalty at midnight.", this.Id)
	case hunger > 1:
		return fmt.Sprintf("%dx hungry. `/feed-tama %d` it %d times to fill its belly.",
			hunger, this.Id, hunger)
	}
	return "<unknown>."
}

// Return a string describing the Tama's parents.
func getParentsString(this *models.Tama) string {
	result := ""
	firstParent := true
	for parentId := range this.Parents.All() {
		parent, err := db.GetTama(this.ServerId, parentId)
		if err == nil {
			if !firstParent {
				result += ", "
			}
			result += parent.GetNameAndId()
			firstParent = false
		} else {
			// Swallow and log errors
			log.Printf("Couldn't fetch parent Tama details: %s\n", err.Error())
		}
	}
	return result
}

// Return a string describing the Tama's children.
func getChildrenString(this *models.Tama) string {
	result := ""
	firstChild := true
	for childId := range this.Children.All() {
		child, err := db.GetTama(this.ServerId, childId)
		if err == nil {
			if !firstChild {
				result += ", "
			}
			result += child.GetNameAndId()
			firstChild = false
		} else {
			// Swallow and log errors
			log.Printf("Couldn't fetch child Tama details: %s\n", err.Error())
		}
	}
	return result
}

// Return a string describing all of the Tama's positive traits and their descriptions.
func getPositiveTraitsString(this *models.Tama) string {
	result := ""
	for trait := range this.PositiveTraits.All() {
		result += "  - "
		switch trait {
		case models.Friendly:
			result += "Friendly: whenever another Tama's attitude toward this Tama improves, " +
				"there's a 33% chance it improves by one extra point.\n"
		case models.SocialButterfly:
			result += "Social Butterfly: 33% chance to execute two social interactions whenever one would normally occur.\n"
		case models.Fertile:
			result += "Fertile: increased chance to mate with the target of love (future feature).\n"
		}
	}
	return result
}

// Return a string describing all of the Tama's negative traits and their descriptions.
func getNegativeTraitsString(this *models.Tama) string {
	result := ""
	for trait := range this.NegativeTraits.All() {
		result += "  - "
		switch trait {
		case models.Bully:
			result += "Bully: increased chance to pick on other Tamas.\n"
		case models.Annoying:
			result += "Annoying: 33% chance that playing with another Tama won't actually improve its attitude " +
				"towards us.\n"
		}
	}
	return result
}

// Return a string describing one of the Tama's relationships with another Tama.
func getRelationshipString(
	target models.JacuzziId,
	score models.RelationshipScore,
	loveTarget models.JacuzziId,
) string {
	result := ""
	switch {
	case score == -5:
		result += "despises"
	case score >= -4 && score <= -3:
		result += "hates"
	case score >= -2 && score <= -1:
		result += "dislikes"
	case score == 0:
		result += "doesn't mind"
	case score >= 1 && score <= 2:
		result += "likes"
	case score >= 3 && score <= 4:
		result += "adores"
	case score == 5:
		if target == loveTarget {
			result += "loves"
		} else {
			result += "is courting"
		}
	}

	if result == "" {
		result = "(unknown)"
	} else {
		result += fmt.Sprintf(" (%s%d)", utils.SignString(score), score)
	}
	return result
}

// Return a string describing all of the Tama's relationships with other Tamas.
func getRelationshipsString(this *models.Tama) string {
	result := ""
	for k, v := range this.Relationships {
		tama, err := db.GetTama(this.ServerId, k)
		if err == nil {
			result += fmt.Sprintf("  - %s %s\n", getRelationshipString(k, v, this.LoveTarget), tama.GetNameAndId())
		} else {
			// Swallow and log
			log.Printf("Couldn't fetch relationship target Tama details: %s\n", err.Error())
		}
	}
	return result
}

// Return a string describing the care state of the Tama.
func getNextCareTimeString(this *models.Tama, timezone *time.Location) string {
	nextCareTime := this.GetNextCareTime()
	if time.Now().Before(nextCareTime) {
		until := time.Until(nextCareTime)
		return fmt.Sprintf("doesn't need any more care until `%s` (in %s)",
			nextCareTime.In(timezone).Format(utils.TimeFormat), utils.FormatUIDuration(until))
	} else {
		result := fmt.Sprintf("can be cared for now by its owner (`/care-tama %d`)", this.Id)
		if this.Mood == models.TamaMoodLimit {
			result += ", but its mood is already at maximum"
		}
		return result
	}
}

// Return a string describing the ownership state of the Tama.
func getOwnerString(this *models.Tama) string {
	ownerString := this.Owner
	if ownerString == "" {
		ownerString = "nobody"
	}
	return fmt.Sprintf("owned by <@%s>", ownerString)
}

// Get a string status message for this Tama. Should describe all the fields a user might care/should know about.
func GetTamaStatus(this *models.Tama, timezone *time.Location, headerLevel string) string {
	result := ""
	if this.IsEgg() {
		// Header: ID
		if headerLevel != "" {
			result += fmt.Sprintf("%s Egg #%d\n", headerLevel, this.Id)
		}

		// Owner
		result += fmt.Sprintf("- It is %s.\n", getOwnerString(this))

		// Care count
		careCountBeforeHatching := constants.EggCareHatchThreshold - this.EggCareCount
		result += fmt.Sprintf("- It has been cared for %d times (needs %d more to hatch).\n",
			this.EggCareCount, careCountBeforeHatching)
	} else {
		// Header: Name and ID
		if headerLevel != "" {
			if this.Name != "" {
				result += fmt.Sprintf("%s %s\n", headerLevel, this.GetNameAndId())
			} else {
				result += fmt.Sprintf("%s Tama #%d\n", headerLevel, this.Id)
			}
		}

		// Owner
		result += fmt.Sprintf("- It is %s.\n", getOwnerString(this))

		// Mood (includes whether it's alive or dead)
		result += fmt.Sprintf("- It is %s", this.GetMoodString())
		hourlyAward := this.GetHourlyPointAward()
		if hourlyAward > 0 {
			result += fmt.Sprintf(", and is earning you %s point%s per hour because of it!\n",
				utils.FormatUIFloat(hourlyAward), utils.Plural(hourlyAward))
		} else {
			result += ".\n"
		}

		// Hunger
		result += fmt.Sprintf("- It is %s\n", getHungerString(this, timezone))

		// Age
		age := time.Since(time.Unix(this.HatchedTime, 0))
		if this.IsAlive() && age < time.Minute {
			result += "- It is freshly hatched!\n"
		} else if this.IsAlive() {
			result += fmt.Sprintf("- It is %s old.\n", utils.FormatUIDuration(age))
		} else {
			result += fmt.Sprintf("- It would have been %s old.\n", utils.FormatUIDuration(age))
		}

		// Parents
		if this.Parents.Size() != 0 {
			result += fmt.Sprintf("- Its parents are: %s.\n", getParentsString(this))
		}

		// Children
		if this.Children.Size() != 0 {
			result += fmt.Sprintf("- Its children are: %s.\n", getChildrenString(this))
		}

		// Positive traits
		result += "- Positive traits:\n"
		result += getPositiveTraitsString(this)

		// Negative traits
		result += "- Negative traits:\n"
		result += getNegativeTraitsString(this)

		// Love
		if this.LoveTarget != models.NoId {
			tama, err := db.GetTama(this.ServerId, this.LoveTarget)
			if err == nil {
				result += fmt.Sprintf("- It is in love with %s!\n", tama.GetNameAndId())
			} else {
				// Swallow and log errors
				log.Printf("Couldn't fetch LoveTarget Tama details: %s\n", err.Error())
			}
		}

		// Relationships
		if len(this.Relationships) > 0 {
			result += "- Relationships:\n"
			result += getRelationshipsString(this)
		}
	}

	// When can it next be cared for?
	result += fmt.Sprintf("- It %s.\n", getNextCareTimeString(this, timezone))

	return result
}
