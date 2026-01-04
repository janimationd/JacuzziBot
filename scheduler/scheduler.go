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
	// How long between scheduled event triggers. If zero, it means nextTime will not be advanced after the next event.
	interval time.Duration
	// The callback function to call when the event triggers. Passes in the event name.
	callback func(string)
}

// Set nextTime to be the next wall clock-aligned interval boundary.
func (event *scheduledEvent) init() {
	// Pre-scheduled events don't need to be initialized.
	if !event.nextTime.IsZero() {
		return
	}

	now := time.Now()
	result := now.Round(event.interval)
	// If it rounded down, add one interval to get our next time.
	if result.Before(now) {
		result = result.Add(event.interval)
	}
	event.nextTime = result
	log.Printf("scheduledEvent \"%s\" initialized with nextTime = %s\n", event.name, event.nextTime.String())
}

// Returns true if the event is not going to happen again, false otherwise.
func (event *scheduledEvent) isDone() bool {
	return event.nextTime.IsZero()
}

// Advances the nextTime of the event to the next scheduled time after its current value.
func (event *scheduledEvent) updateNextTime() {
	if event.nextTime.IsZero() {
		return
	}
	if event.interval == 0 {
		log.Printf("scheduledEvent \"%s\" will not execute again.\n", event.name)
		// Reset nextTime to 0.
		event.nextTime = time.Time{}
		return
	}
	event.nextTime = event.nextTime.Add(event.interval)
	log.Printf("scheduledEvent \"%s\" advanced to nextTime = %s\n", event.name, event.nextTime.String())
}

func (event *scheduledEvent) check() {
	now := time.Now()

	// If it's time to execute the event
	if now.Equal(event.nextTime) || now.After(event.nextTime) {
		event.callback(event.name)
		event.updateNextTime()
	}
}

var schedule []*scheduledEvent

// Puts all known events on the schedule.
func setup() {
	// Create and append all scheduled events here
	schedule = append(schedule, &scheduledEvent{
		name:     "ExampleRecurringEvent",
		interval: 1 * time.Minute,
		callback: func(eventName string) { log.Println(eventName + ": " + time.Now().String()) },
	})
	schedule = append(schedule, &scheduledEvent{
		name:     "ExampleOneOffEvent",
		nextTime: time.Now().Add(30 * time.Second),
		callback: func(eventName string) { log.Println(eventName + ": " + time.Now().String()) },
	})

	// Initialize all the events and set their next times.
	for _, event := range schedule {
		event.init()
	}
}

const checkinterval = 1 * time.Second

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
				if !event.isDone() {
					event.check()
				}
			}
		}
		// This won't be perfectly accurate, but that's fine.
		loopEndTime := time.Now()
		loopDuration := loopEndTime.Sub(loopStartTime)
		sleepDuration := max(checkinterval-loopDuration, 0)
		time.Sleep(sleepDuration)
	}
}
