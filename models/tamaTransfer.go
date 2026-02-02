package models

// A transfer of a Tama between owners. Uniquely identified by the Tama's ID, since only one transfer can be pending at
// a time for each Tama.
type TamaTransfer struct {
	TamaId     JacuzziId
	OldOwnerId string
	NewOwnerId string
}
