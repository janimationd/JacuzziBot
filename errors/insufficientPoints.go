package errors

import (
	"fmt"

	"github.com/janimationd/JacuzziBot/utils"
)

// The user doesn't have enough points to complete the action
type InsufficientPointsError struct {
	CurrentPoints  float64
	RequiredPoints float64
}

func (e *InsufficientPointsError) Error() string {
	return fmt.Sprintf(
		"Insufficient points to complete action: %s points vs %s required.",
		utils.FormatUIFloat(e.CurrentPoints),
		utils.FormatUIFloat(e.RequiredPoints),
	)
}
