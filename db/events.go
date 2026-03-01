package db

import (
	"encoding/json"
	"log"

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
type ScheduledEventOperation = func(*models.ScheduledEvent) (EventOperationResult, error)

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
			opResult, err := op(event)
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

func scheduleEvent(db *bbolt.DB, event *models.ScheduledEvent, overwriteIfPresent bool) (bool, error) {
	modified := false

	err := db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(scheduleBucketName))
		if err != nil {
			return err
		}

		if !overwriteIfPresent {
			existingEventBytes := bucket.Get([]byte(event.ID))
			if existingEventBytes != nil {
				return nil
			}
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(event.ID), eventBytes)
		if err != nil {
			modified = true
		}

		return err
	})

	if err != nil {
		log.Printf("Couldn't schedule event %s: %s\n", event.ID, err.Error())
	}

	return modified, err
}

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
func ScheduleEvent(event *models.ScheduledEvent, overwriteIfPresent bool) (bool, error) {
	// Create or open a server-specific database file
	db, err := getDb(scheduleDatabaseName)
	if err != nil {
		return false, err
	}
	defer db.Close()

	return scheduleEvent(db, event, overwriteIfPresent)
}
