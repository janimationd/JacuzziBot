package startup

import (
	"github.com/janimationd/JacuzziBot/scheduler"
	"github.com/janimationd/JacuzziBot/scheduler/events"
)

// Puts all crucial events on the schedule at startup. "Crucial" here just means "we know we need it at startup".
func SetupCrucialEventsAndHandlers() {
	// Schedule all crucial events here. Examples (though they should each be defined in their own files):
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   ID:                  "RecurringEvent",
	//     Interval:            1 * time.Minute,
	//     RestartGapTolerance: 5 * time.Minute,
	//     ...
	// })
	//
	// schedule = append(schedule, &ScheduledEvent{
	// 	   ID:       "OneOffEvent",
	// 	   NextTime: time.Now().Add(30 * time.Second),
	//     ...
	// })
	scheduler.RegisterEventAndHandler(&events.VoiceCallPointAwarder, events.VoiceCallPointAwarderHandler, true, nil)
	scheduler.RegisterEventAndHandler(&events.TamaMoodEarnsPointsAwarder, events.TamaMoodEarnsPointsHandler, true, nil)

	// Register all handlers for events that could be stored in the database already or could be scheduled later.
	scheduler.RegisterHandler("TamaPlaytimeHandler", events.TamaPlaytimeHandler, true)
	scheduler.RegisterHandler("TamaDeathHandler", events.TamaDeathHandler, true)
}
