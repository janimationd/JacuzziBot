package models

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Parse a CustomId like "Action|10" into JacuzziId 10.
func ParseCustomIdToJacuzziId(customId string) JacuzziId {
	parts := strings.Split(customId, "|")
	if len(parts) != 2 {
		panic(fmt.Errorf("Couldn't parse CustomID %s into a single JacuzziId", customId))
	} else {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(fmt.Errorf("Couldn't parse CustomID second part %s: %w", parts[1], err))
		} else {
			return JacuzziId(id)
		}
	}
}

// Parse a CustomId like "Action|10|Text" into JacuzziId 10 and "Text".
func ParseCustomIdToJacuzziIdAndString(customId string) (JacuzziId, string) {
	log.Printf("Parsing CustomId \"%s\"\n", customId)
	parts := strings.Split(customId, "|")
	if len(parts) != 3 {
		panic(fmt.Errorf("Couldn't parse CustomID %s into a single JacuzziId and a string", customId))
	} else {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(fmt.Errorf("Couldn't parse CustomID second part %s: %w", parts[1], err))
		}

		return JacuzziId(id), parts[2]
	}
}
