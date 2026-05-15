package models

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

// An event in the schedule.
type ScheduledEvent struct {
	// A server-unique ID for this event. This should also be descriptive, e.g. also function as an event name. If
	// you're generating many instances of an event from one part of code, you'll need to make sure this is unique
	// somehow, for example, by appending a millisecond (or better) timestamp.
	ID string
	// The next scheduled time to trigger the event at. If zero, it means the event will not happen again.
	NextTime time.Time
	// How long between scheduled event triggers. If zero, it means NextTime will not be advanced after the next event.
	Interval time.Duration
	// If this is a recurring event, this defines how long the bot has to be shut down for before we give up on trying
	// to catch back up to where we were. For example: if Interval is X, this is Y, and the bot is shut down for longer
	// than Y, then we don't race through all the Intervals we missed. Instead we just skip NextTime ahead to the next
	// Interval boundary. In practice we don't expect long bot outages for the actual bot, but this happens often with
	// our personal bot instances used for testing and local development.
	RestartGapTolerance time.Duration
	// If the implementation of the event needs to use LastCheckTime for its logic, then set this to true. Otherwise
	// should be false. The reason we don't assume true for every event is this causes a write back to the DB after
	// every per-second check, so it's a performance optimization.
	UsesLastCheckTime bool
	// If UsesLastCheckTime is true, this is the last time the event was checked to see if it should execute. Only some
	// event types will need to use this, such as events that check poll for specific, asynchronous, non-consumable
	// events transpiring (e.g. wall clock passing a certain time). This isn't great and could probably be improved.
	// This is managed for you, so don't touch this.
	LastCheckTime time.Time
	// The handler name for the event.
	Handler string
	// A data payload to be stored with the event, containing the parameters to the Handler.
	Payload json.RawMessage
}

// A function to handle an event when it is time for the event to "happen".
//
// CONTRACT RULES:
//  1. You are allowed to edit the event as you see fit, but your handler MUST return true if you edit it so we know to
//     update the event in the database.
//  2. You don't need to manually modify timing fields on the event unless you're doing something non-standard.
//  3. event.NextTime is advanced by event.Interval by calling code after your handler is called.
//  4. The "now" param is a time you should use as "the time this event is executing at" in place of any additional
//     time.Now() calls. This is the same time we use to determine if your event needs to be executed, so this
//     maintains zero-gap timing within your logic between executions of the event.
//  5. DO NOT initiate any new transactions in the events DB WITHOUT passing tx down and reusing it, or else the
//     program WILL deadlock.
type EventHandler func(event *ScheduledEvent, now time.Time, tx *bbolt.Tx) bool

// Set NextTime to be the next wall clock-aligned Interval boundary.
func (event *ScheduledEvent) Init() {
	// Pre-scheduled events don't need to be initialized.
	if !event.NextTime.IsZero() {
		return
	}

	now := time.Now()
	result := now.Round(event.Interval)
	// If it rounded down, add one Interval to get our next time.
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

// Advances the NextTime of the event to the next scheduled time after its current value. Returns whether the event
// was modified.
func (event *ScheduledEvent) updateNextTime() bool {
	if event.NextTime.IsZero() {
		log.Printf("ScheduledEvent \"%s\" has no NextTime to advance past.\n", event.ID)
		return false
	}
	if event.Interval == 0 {
		log.Printf("ScheduledEvent \"%s\" will not execute again.\n", event.ID)
		// Reset NextTime to 0.
		event.NextTime = time.Time{}
		return true
	}
	event.NextTime = event.NextTime.Add(event.Interval)
	log.Printf("ScheduledEvent \"%s\" advanced to NextTime = %s\n", event.ID, event.NextTime.String())
	return true
}

// Check if it's time for the event to happen, and then handle it and advance NextTime if yes. Returns whether or not
// the event was modified.
func (event *ScheduledEvent) CheckAndHandle(handler EventHandler, now time.Time, tx *bbolt.Tx) bool {
	modified := false

	// If NextTime is more than RestartGapTolerance in the past, then the bot has been shut down for a while and we
	// shouldn't iterate through every Interval since it last ran to catch back up. Instead skip ahead to what NextTime
	// should next be.
	shouldCheckRestartGapTolerance := event.Interval != 0 && event.RestartGapTolerance != 0
	if shouldCheckRestartGapTolerance && time.Until(event.NextTime) <= -event.RestartGapTolerance {
		originalNextTime := event.NextTime
		event.NextTime = time.Time{}
		// Reinitialize the event to properly calculate NextTime from the Interval.
		event.Init()
		log.Printf("Event %s's NextTime was [%s], which is too far in the past. "+
			"Advancing NextTime to [%s] and skipping the Intervals we missed.\n",
			event.ID, originalNextTime, event.NextTime)
		modified = true
	} else if !now.Before(event.NextTime) {
		// If it's time to execute the event...
		// Order of operands here is important: always execute the commands even if modified is already true.
		modified = handler(event, now, tx) || modified
		modified = event.updateNextTime() || modified
	}
	// See if LastCheckTime needs to be updated
	if event.UsesLastCheckTime {
		event.LastCheckTime = now
		modified = true
	}
	return modified
}
