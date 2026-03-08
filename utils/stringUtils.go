package utils

import (
	"fmt"
	"math"
	"strings"
	"time"

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

// Roughly describes a duration of time in a human-readable format, e.g. "15 minutes" or "3 months".
func FormatUIDuration(duration time.Duration) string {
	const oneDay = 24 * time.Hour
	// We approximate 1 month as 30 days; sue me.
	const oneMonth = 30 * oneDay
	// Same thing, 365 days are a year.
	const oneYear = 365 * oneDay

	switch {
	case 0 <= duration && duration < time.Minute:
		amount := int64(duration.Seconds())
		return fmt.Sprintf("%d second%s", amount, Plural(float64(amount)))
	case time.Minute <= duration && duration < time.Hour:
		amount := int64(duration.Minutes())
		return fmt.Sprintf("%d minute%s", amount, Plural(float64(amount)))
	case time.Hour <= duration && duration < oneDay:
		amount := int64(duration.Hours())
		return fmt.Sprintf("%d hour%s", amount, Plural(float64(amount)))
	case oneDay <= duration && duration < oneMonth:
		amount := int64(duration.Hours() / 24)
		return fmt.Sprintf("%d day%s", amount, Plural(float64(amount)))
	case oneMonth <= duration && duration < oneYear:
		amount := int64(duration.Hours() / 24 / 30)
		return fmt.Sprintf("%d month%s", amount, Plural(float64(amount)))
	default:
		amount := int64(duration.Hours() / 24 / 365)
		return fmt.Sprintf("%d years%s", amount, Plural(float64(amount)))
	}
}

// Get the sign string of the numeric values ("-" for negatives, "" for zero, and "+" for positives).
func SignString[T Number](val T) string {
	switch {
	case val < 0:
		return "-"
	case val > 0:
		return "+"
	default:
		return ""
	}
}
