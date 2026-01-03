package scheduler

import (
	"context"
	"log"
	"time"
)

// A recurring event in the schedule
type scheduledEvent struct {
	// The name of the event (should be unique, but not enforced)
	name string
	// The next scheduled time to trigger the event at. If zero, it means the event will not happen again.
	nextTime time.Time
	// How long between scheduled event triggers. If zero, it means NextTime will not be updated after the next event.
	Interval time.Duration
	// The callback function to call when the event triggers. Passes in the event name.
	callback func(string)
}

// Set NextTime to be the next wall clock-aligned Interval boundary.
func (event *scheduledEvent) init() {
	now := time.Now()
	result := now.Round(event.Interval)
	// If it rounded down, add one interval to get our next time.
	if result.Before(now) {
		result.Add(event.Interval)
	}
	event.nextTime = result
	log.Printf("scheduledEvent \"%s\" initialized with nextTime = %s\n", event.name, event.nextTime.String())
}

// Advances the NextTime of the event to the next scheduled time after its current value.
func (event *scheduledEvent) advanceNextTime() {
	if event.nextTime.Equal(time.UnixMilli(0)) || event.Interval == 0 {
		return
	}
	event.nextTime = event.nextTime.Add(event.Interval)
	log.Printf("scheduledEvent \"%s\" advanced to nextTime = %s\n", event.name, event.nextTime.String())
}

func (event *scheduledEvent) check() {
	now := time.Now()

	// If it's time to execute the event
	if now.Equal(event.nextTime) || now.After(event.nextTime) {
		event.callback(event.name)
		event.advanceNextTime()
	}
}

var schedule []*scheduledEvent

// Puts all known events on the schedule.
func setup() {
	// Create and append all scheduled events here
	schedule = append(schedule, &scheduledEvent{
		name:     "VoiceCallPointAwarder",
		Interval: 1 * time.Minute,
		callback: func(eventName string) { log.Println(eventName + ": " + time.Now().String()) },
	})

	// Initialize all the events and set their next times.
	for _, event := range schedule {
		event.init()
	}
}

const checkInterval = 1 * time.Second

// Setup and run the schedule.
func Run(ctx context.Context) {
	setup()

	for {
		loopStartTime := time.Now()
		select {
		case <-ctx.Done():
			log.Println("Done signal received; shutting down scheduler.")
			return
		default:
			for _, event := range schedule {
				event.check()
			}
		}
		// This won't be perfectly accurate, but that's fine.
		loopEndTime := time.Now()
		loopDuration := loopEndTime.Sub(loopStartTime)
		sleepDuration := max(checkInterval-loopDuration, 0)
		time.Sleep(sleepDuration)
	}
}
