package utils

import (
	"fmt"
	"strings"

	"github.com/janimationd/JacuzziBot/constants"
)

// Formats a floating point value with up to a certain amount of decimal precision, as needed.
func FormatUIFloat(f float64) string {
	formatStr := fmt.Sprintf("%%.%df", constants.UIFloatMaxDisplayedDecimalPrecision)
	result := fmt.Sprintf(formatStr, f)

	// Trim any trailing zeros or decimal points
	result = strings.TrimRight(result, "0")
	result = strings.TrimRight(result, ".")

	return result
}
