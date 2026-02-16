package models

import (
	"encoding/json"
	"log"
	"time"
)

// An event in the schedule.
type ScheduledEvent struct {
	// A server-unique ID for this event. This should also be descriptive, e.g. also function as an event name.
	ID string
	// The next scheduled time to trigger the event at. If zero, it means the event will not happen again.
	NextTime time.Time
	// How long between scheduled event triggers. If zero, it means nextTime will not be advanced after the next event.
	Interval time.Duration
	// The handler name for the event.
	Handler string
	// A data payload to be stored with the event, containing the parameters to the Handler.
	Payload json.RawMessage
}

// A function to handle an event when it is time for the event to "happen". You are allowed to edit the event as you
// see fit. You DO NOT need to do any timing logic in your function unless you're doing something non-standard.
// NextTime is advanced by Interval for you after your handler is called.
type EventHandler func(event *ScheduledEvent)

// Set nextTime to be the next wall clock-aligned interval boundary.
func (event *ScheduledEvent) Init() {
	// Pre-scheduled events don't need to be initialized.
	if !event.NextTime.IsZero() {
		return
	}

	now := time.Now()
	result := now.Round(event.Interval)
	// If it rounded down, add one interval to get our next time.
	if result.Before(now) {
		result = result.Add(event.Interval)
	}
	event.NextTime = result
	log.Printf("ScheduledEvent \"%s\" initialized with nextTime = %s\n", event.ID, event.NextTime.String())
}

// Returns true if the event is not going to happen again, false otherwise.
func (event *ScheduledEvent) IsDone() bool {
	return event.NextTime.IsZero()
}

// Advances the nextTime of the event to the next scheduled time after its current value.
func (event *ScheduledEvent) updateNextTime() {
	if event.NextTime.IsZero() {
		return
	}
	if event.Interval == 0 {
		log.Printf("ScheduledEvent \"%s\" will not execute again.\n", event.ID)
		// Reset nextTime to 0.
		event.NextTime = time.Time{}
		return
	}
	event.NextTime = event.NextTime.Add(event.Interval)
	log.Printf("ScheduledEvent \"%s\" advanced to nextTime = %s\n", event.ID, event.NextTime.String())
}

// Check if it's time for the event to happen, and then handle it and advance NextTime if yes. Returns whether or not
// the event happened.
func (event *ScheduledEvent) CheckAndHandle(handler EventHandler) bool {
	now := time.Now()

	// If it's time to execute the event
	if now.Equal(event.NextTime) || now.After(event.NextTime) {
		handler(event)
		event.updateNextTime()
		return true
	}
	return false
}
