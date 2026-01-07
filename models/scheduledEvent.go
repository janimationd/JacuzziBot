package models

import (
	"log"
	"time"
)

// A recurring event in the schedule
type ScheduledEvent struct {
	// The Name of the event (should be unique, but not enforced)
	Name string
	// The next scheduled time to trigger the event at. If zero, it means the event will not happen again.
	NextTime time.Time
	// How long between scheduled event triggers. If zero, it means nextTime will not be advanced after the next event.
	Interval time.Duration
	// The Callback function to call when the event triggers. Passes in the event name.
	Callback func(string)
}

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
	log.Printf("ScheduledEvent \"%s\" initialized with nextTime = %s\n", event.Name, event.NextTime.String())
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
		log.Printf("ScheduledEvent \"%s\" will not execute again.\n", event.Name)
		// Reset nextTime to 0.
		event.NextTime = time.Time{}
		return
	}
	event.NextTime = event.NextTime.Add(event.Interval)
	log.Printf("ScheduledEvent \"%s\" advanced to nextTime = %s\n", event.Name, event.NextTime.String())
}

func (event *ScheduledEvent) Check() {
	now := time.Now()

	// If it's time to execute the event
	if now.Equal(event.NextTime) || now.After(event.NextTime) {
		event.Callback(event.Name)
		event.updateNextTime()
	}
}
