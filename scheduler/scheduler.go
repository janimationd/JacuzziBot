package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/scheduler/events"
)

// The registered event handlers.
var eventRegistry = map[string]models.EventHandler{}

// This ensures both that the event is in the database already (indexed by event.ID) and that the handler is
// registered (indexed by event.Handler). Caller chooses whether to overwrite if values for those keys already exist.
func RegisterEventAndHandler(
	event *models.ScheduledEvent,
	handler models.EventHandler,
	overwriteIfPresent bool,
) error {
	event.Init()
	// Add to the database.
	_, err := db.ScheduleEvent(event, overwriteIfPresent)
	if err != nil {
		return err
	}
	if overwriteIfPresent || eventRegistry[event.Handler] == nil {
		eventRegistry[event.Handler] = handler
	}
	return nil
}

// Puts all crucial events on the schedule at startup. "Crucial" here just means "we know we need it at startup".
func setupCrucialEvents() {
	// Create and append all scheduled events here. Examples (though they should each be defined in their own files):
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   name:     "ExampleRecurringEvent",
	//     interval: 1 * time.Minute,
	//     ...
	// })
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   name:     "ExampleOneOffEvent",
	// 	   nextTime: time.Now().Add(30 * time.Second),
	//     ...
	// })

	RegisterEventAndHandler(&events.VoiceCallPointAwarder, events.VoiceCallPointAwarderHandler, false)
}

const checkinterval = 1 * time.Second

// Setup and run the schedule.
func Run(ctx context.Context) {
	setupCrucialEvents()

	for {
		loopStartTime := time.Now()
		select {
		case <-ctx.Done():
			log.Println("Done signal received; shutting down scheduler.")
			return
		default:
			// Execute logic directly inside the database transactions to ensure thread safety.
			db.ForEachScheduledEvent(func(event *models.ScheduledEvent) (db.EventOperationResult, error) {
				handler := eventRegistry[event.Handler]
				if handler == nil {
					log.Printf("Event %s had non existent handler %s\n", event.ID, event.Handler)
					return db.DoNothing, nil
				}

				// Run the handler if it's time.
				modified := event.CheckAndHandle(handler)

				if event.IsDone() {
					log.Printf("Event %s is done; deleting it\n", event.ID)
					return db.DeleteEvent, nil
				} else if modified {
					log.Printf("Event %s was modified; updating it\n", event.ID)
					return db.UpdateEvent, nil
				}
				return db.DoNothing, nil
			})
		}
		// This won't be perfectly accurate, but that's fine.
		loopEndTime := time.Now()
		loopDuration := loopEndTime.Sub(loopStartTime)
		sleepDuration := max(checkinterval-loopDuration, 0)
		time.Sleep(sleepDuration)
	}
}
