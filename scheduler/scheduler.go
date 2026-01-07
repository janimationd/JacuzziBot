package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/scheduler/events"
)

var schedule []*models.ScheduledEvent

// Puts all known events on the schedule.
func setup() {
	// Create and append all scheduled events here. Examples (though they should each be defined in their own files):
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   name:     "ExampleRecurringEvent",
	//     interval: 1 * time.Minute,
	// 	   callback: func(eventName string) { log.Println(eventName + ": " + time.Now().String()) },
	// })
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   name:     "ExampleOneOffEvent",
	// 	   nextTime: time.Now().Add(30 * time.Second),
	// 	   callback: func(eventName string) { log.Println(eventName + ": " + time.Now().String()) },
	// })

	schedule = append(schedule, &events.VoiceCallPointAwarder)

	// Initialize all the events and set their next times.
	for _, event := range schedule {
		event.Init()
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
				if !event.IsDone() {
					event.Check()
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
