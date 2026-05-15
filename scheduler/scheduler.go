package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/models"
	"go.etcd.io/bbolt"
)

// The registered event handlers.
var eventRegistry = map[string]models.EventHandler{}

// Register an event handler.
func RegisterHandler(handlerName string, handler models.EventHandler, overwriteIfPresent bool) {
	if overwriteIfPresent || eventRegistry[handlerName] == nil {
		eventRegistry[handlerName] = handler
	}
}

// This ensures both that the event is in the database already (indexed by event.ID) and that the handler is
// registered (indexed by event.Handler). Caller chooses whether to overwrite if values for those keys already exist.
// Can be called from a re-entrant context, so reuse the tx if provided.
func RegisterEventAndHandler(
	event *models.ScheduledEvent,
	handler models.EventHandler,
	overwriteIfPresent bool,
	tx *bbolt.Tx,
) error {
	//debug.PrintStack()
	// Add to the database.
	_, err := db.ScheduleEvent(event, overwriteIfPresent, tx)
	if err != nil {
		return err
	}
	RegisterHandler(event.Handler, handler, overwriteIfPresent)
	return nil
}

const checkInterval = 1 * time.Second

// Setup and run the schedule.
func Run(ctx context.Context) {
	for {
		// Use a single now calculation and pass it down into the events to maintain zero-gap timing.
		now := time.Now()
		select {
		case <-ctx.Done():
			log.Println("Done signal received; shutting down scheduler.")
			return
		default:
			// Execute all logic directly inside a database transaction to ensure thread safety.
			db.ForEachScheduledEvent(func(event *models.ScheduledEvent, tx *bbolt.Tx) (db.EventOperationResult, error) {
				handler := eventRegistry[event.Handler]
				if handler == nil {
					log.Printf("Event %s had non existent handler %s\n", event.ID, event.Handler)
					return db.DoNothing, nil
				}

				// Run the handler if it's time.
				modified := event.CheckAndHandle(handler, now, tx)

				if event.IsDone() {
					log.Printf("Event %s is done; deleting it\n", event.ID)
					return db.DeleteEvent, nil
				} else if modified {
					//log.Printf("Event %s was modified; updating it\n", event.ID)
					return db.UpdateEvent, nil
				}
				return db.DoNothing, nil
			})
		}
		// This won't be perfectly accurate, but that's fine.
		loopEndTime := time.Now()
		loopDuration := loopEndTime.Sub(now)
		sleepDuration := max(checkInterval-loopDuration, 0)
		time.Sleep(sleepDuration)
	}
}
