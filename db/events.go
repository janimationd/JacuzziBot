package db

import (
	"encoding/json"
	"log"
	"regexp"
	"sort"

	"github.com/janimationd/JacuzziBot/models"
	"go.etcd.io/bbolt"
)

const scheduleDatabaseName = "ScheduledEvents"
const scheduleBucketName = scheduleDatabaseName

// What should we do with the event after your operation completes?
type EventOperationResult int8

const (
	// Don't do anything with the event afdter your operation completes.
	DoNothing EventOperationResult = 0
	// Update the event in the database after your operation completes (because you modified it).
	UpdateEvent EventOperationResult = 1
	// Delete the event in the database after your operation completes (because you no longer need it).
	DeleteEvent EventOperationResult = 2
)

// You MUST NOT read/modify the database inside these operations or we'll be at risk of deadlocking/undefined behavior.
// The tx parameter should be passed down to any re-entrant DB operations on the events DB.
type ScheduledEventOperation = func(*models.ScheduledEvent, *bbolt.Tx) (EventOperationResult, error)

func forEachScheduledEvent(db *bbolt.DB, op ScheduledEventOperation) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(scheduleBucketName))
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			event := &models.ScheduledEvent{}
			err := json.Unmarshal(v, event)
			if err != nil {
				// Don't fail the whole loop, just log and skip over the broken one.
				log.Printf("Couldn't unmarshall event %s: %s\n", k, err.Error())
				return nil
			}

			// Execute the caller-provided operation on the event
			opResult, err := op(event, tx)
			if err != nil {
				return err
			}
			switch opResult {
			case UpdateEvent:
				eventBytes, err := json.Marshal(event)
				if err != nil {
					// Don't fail the whole loop, just log and skip over the broken one.
					log.Printf("Couldn't marshall event %s: %s\n", k, err.Error())
					return nil
				}
				err = bucket.Put(k, eventBytes)
				if err != nil {
					// Don't fail the whole loop, just log and skip over the broken one.
					log.Printf("Couldn't update original event %s: %s\n", k, err.Error())
					return nil
				}
			case DeleteEvent:
				err := bucket.Delete(k)
				if err != nil {
					// Don't fail the whole loop, just log and skip over the broken one.
					log.Printf("Couldn't delete event %s: %s\n", k, err.Error())
					return nil
				}
			}
			return nil
		})
		return nil
	})
}

func scheduleEventWithTx(tx *bbolt.Tx, event *models.ScheduledEvent, overwriteIfPresent bool) (bool, error) {
	modified := false

	bucket, err := tx.CreateBucketIfNotExists([]byte(scheduleBucketName))
	if err != nil {
		return modified, err
	}

	if !overwriteIfPresent {
		existingEventBytes := bucket.Get([]byte(event.ID))
		if existingEventBytes != nil {
			return modified, nil
		}
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return modified, err
	}
	err = bucket.Put([]byte(event.ID), eventBytes)
	if err == nil {
		modified = true
	}

	return modified, err
}

func scheduleEvent(db *bbolt.DB, event *models.ScheduledEvent, overwriteIfPresent bool, tx *bbolt.Tx) (bool, error) {
	modified := false
	event.Init()
	var err error = nil

	if tx == nil {
		err = db.Update(func(tx *bbolt.Tx) error {
			modified, err = scheduleEventWithTx(tx, event, overwriteIfPresent)
			return err
		})
	} else {
		modified, err = scheduleEventWithTx(tx, event, overwriteIfPresent)
	}

	if err != nil {
		log.Printf("Couldn't schedule event %s: %s\n", event.ID, err.Error())
	}

	return modified, err
}

func getAllEvents(db *bbolt.DB, idFilterRegex string) ([]*models.ScheduledEvent, error) {
	result := make([]*models.ScheduledEvent, 0)

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(scheduleBucketName))
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			matched, err := regexp.Match(idFilterRegex, k)
			if err != nil {
				log.Printf("Couldn't match regex %s to event ID %s: %s\n", idFilterRegex, k, err.Error())
				return nil
			}
			if matched {
				event := new(models.ScheduledEvent)
				err := json.Unmarshal(v, event)
				if err != nil {
					log.Printf("Couldn't unmarshall event to JSON: %s\n", err.Error())
					return nil
				}
				result = append(result, event)
			}
			return nil
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by ID
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func cancelEvent(db *bbolt.DB, eventId string) bool {
	cancelled := false
	db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(scheduleBucketName))
		if bucket == nil {
			return nil
		}

		exists := bucket.Get([]byte(eventId)) != nil
		if exists {
			err := bucket.Delete([]byte(eventId))
			if err != nil {
				log.Printf("Couldn't delete scheduled event: %s\n", err.Error())
				return err
			}

			cancelled = true
		}
		return nil
	})

	return cancelled
}

// PUBLIC METHODS

// Perform an operation on each event in the schedule. Your op returns what we should do with the event after
// the function completes. If op returns an error, the entire loop exits early and the error is returned.
func ForEachScheduledEvent(op ScheduledEventOperation) error {
	// Create or open a server-specific database file
	db, err := getDb(scheduleDatabaseName)
	if err != nil {
		return err
	}
	defer db.Close()

	return forEachScheduledEvent(db, op)
}

// Adds a new event to the schedule. Returns whether the event was added into the schedule in or not.
// Can be used in a re-entrant context, so reuse tx if it's passed in.
func ScheduleEvent(event *models.ScheduledEvent, overwriteIfPresent bool, tx *bbolt.Tx) (bool, error) {
	if tx == nil {
		// Create or open a server-specific database file
		db, err := getDb(scheduleDatabaseName)
		if err != nil {
			return false, err
		}
		defer db.Close()

		return scheduleEvent(db, event, overwriteIfPresent, nil)
	} else {
		return scheduleEvent(nil, event, overwriteIfPresent, tx)
	}
}

// Get all events in the DB, with an optional ID filter.
// The slice will be sorted by event ID (so by event time).
func GetAllEvents(idFilterRegex string) ([]*models.ScheduledEvent, error) {
	// Create or open a server-specific database file
	db, err := getDb(scheduleDatabaseName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getAllEvents(db, idFilterRegex)
}

// Cancel an event based on its ID. Returns whether the event was cancelled. If false, assume the event didn't exist.
func CancelEvent(eventId string) bool {
	// Create or open a server-specific database file
	db, err := getDb(scheduleDatabaseName)
	if err != nil {
		log.Printf("Couldn't open schedule DB to cancel: %s\n", err.Error())
		return false
	}
	defer db.Close()

	return cancelEvent(db, eventId)
}
