package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/janimationd/JacuzziBot/constants"
)

// Formats a floating point value with up to a certain amount of decimal precision, as needed.
func FormatUIFloat(f float64) string {
	var result string

	// If the amount would normally just show up as "0" when it actually isn't.
	limit := math.Pow(10, -float64(constants.UIFloatMaxDisplayedDecimalPrecision))
	if f > -limit && f < limit && f != 0 {
		result = fmt.Sprintf("%f", f)
	} else {
		formatStr := fmt.Sprintf("%%.%df", constants.UIFloatMaxDisplayedDecimalPrecision)
		result = fmt.Sprintf(formatStr, f)
	}

	// Trim any trailing zeros or decimal points
	result = strings.TrimRight(result, "0")
	result = strings.TrimRight(result, ".")

	return result
}

// Returns "s" if the amount would warrant it, otherwise returns "".
// For formatting plural nouns, e.g. "0 points", "0.5 points", "1 point", "1.5 points", "2 points"
func Plural(amount float64) string {
	if math.Abs(amount) != 1 {
		return "s"
	}
	return ""
}
