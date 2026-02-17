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
	// If this is a recurring event, this defines how long the bot has to be shut down for before we give up on trying
	// to catch back up to where we were. For example: if Interval is X, this is Y, and the bot is shut down for longer
	// than Y, then we don't race through all the Intervals we missed. Instead we just skip NextTime ahead to the next
	// Interval boundary. In practice we don't expect long bot outages for the actual bot, but this happens often with
	// our personal bot instances used for testing and local development.
	RestartGapTolerance time.Duration
	// The handler name for the event.
	Handler string
	// A data payload to be stored with the event, containing the parameters to the Handler.
	Payload json.RawMessage
}

// A function to handle an event when it is time for the event to "happen". You are allowed to edit the event as you
// see fit, but your handler MUST return true if you edit it so we know to update the event in the database. You DO
// NOT need to do any timing logic in your function unless you're doing something non-standard. NextTime is advanced
// by Interval for you after your handler is called.
type EventHandler func(event *ScheduledEvent) bool

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
	log.Printf("ScheduledEvent \"%s\" initialized with NextTime = [%s]\n", event.ID, event.NextTime.String())
}

// Returns true if the event is not going to happen again, false otherwise.
func (event *ScheduledEvent) IsDone() bool {
	return event.NextTime.IsZero()
}

// Advances the nextTime of the event to the next scheduled time after its current value. Returns whether the event
// was modified.
func (event *ScheduledEvent) updateNextTime() bool {
	if event.NextTime.IsZero() {
		log.Printf("ScheduledEvent \"%s\" has no NextTime to advance past.\n", event.ID)
		return false
	}
	if event.Interval == 0 {
		log.Printf("ScheduledEvent \"%s\" will not execute again.\n", event.ID)
		// Reset nextTime to 0.
		event.NextTime = time.Time{}
		return true
	}
	event.NextTime = event.NextTime.Add(event.Interval)
	log.Printf("ScheduledEvent \"%s\" advanced to NextTime = %s\n", event.ID, event.NextTime.String())
	return true
}

// Check if it's time for the event to happen, and then handle it and advance NextTime if yes. Returns whether or not
// the event was modified.
func (event *ScheduledEvent) CheckAndHandle(handler EventHandler) bool {
	// If NextTime is more than RestartGapTolerance in the past, then the bot has been shut down for a while and we
	// shouldn't iterate through every Interval since it last ran to catch back up. Instead skip ahead to what NextTime
	// should next be.
	shouldCheckRestartGapTolerance := event.Interval != 0 && event.RestartGapTolerance != 0
	// log.Printf("> Interval = %s, RestartGapTolerance = %s, NextTime = [%s], TimeUntil = %s\n",
	// 	event.Interval.String(), event.RestartGapTolerance.String(), event.NextTime, time.Until(event.NextTime).String())
	if shouldCheckRestartGapTolerance && time.Until(event.NextTime) <= -event.RestartGapTolerance {
		originalNextTime := event.NextTime
		event.NextTime = time.Time{}
		// Reinitialize the event to properly calculate NextTime from the Interval.
		event.Init()
		log.Printf("Event %s's NextTime was [%s], which is too far in the past. "+
			"Advancing NextTime to [%s] and skipping the Intervals we missed.\n",
			event.ID, originalNextTime, event.NextTime)
		return true
	}

	now := time.Now()
	// If it's time to execute the event
	if now.Equal(event.NextTime) || now.After(event.NextTime) {
		modified := handler(event)
		// Order of operands is important: always execute updateNextTime().
		modified = event.updateNextTime() || modified
		return modified
	}
	return false
}
