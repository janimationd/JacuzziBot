package models

type TamaDeathPayload struct {
	// The ID of the server that the Tama belongs to.
	ServerId string
	// The ID of the Tama that died.
	TamaId JacuzziId
	// The cause of this Tama's death.
	Cause string
}
