package errs

import (
	"fmt"

	"github.com/janimationd/JacuzziBot/constants"
)

// The user has already reached their limit of Tamas/eggs that they can own.
type TamaLimitReachedError struct{}

func (e TamaLimitReachedError) Error() string {
	return fmt.Sprintf("User limit for Tama/eggs of %d has already been reached.", constants.TamaLimitPerUser)
}
