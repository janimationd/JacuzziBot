package models

// The payload to a TamaPlaytime scheduled event.
type TamaPlaytimePayload struct {
	// The server ID where this event was scheduled from
	ServerId string
}
