package db

import (
	"log"

	"github.com/janimationd/JacuzziBot/models"
	bolt "go.etcd.io/bbolt"
)

// This database keeps track of the next incrementally increasing JacuzziId that's available.
// We purposefully don't track a database of all IDs mapped to what they were claimed by.

const nextIdBucketName string = "NextJacuzziId"
const nextIdKey string = "next"

func claimNextJacuzziId(db *bolt.DB) (models.JacuzziId, error) {
	var nextId models.JacuzziId

	// Do all the logic in the same transaction
	err := db.Update(func(tx *bolt.Tx) error {
		// Create or fetch the IDs bucket
		bucket, err := tx.CreateBucketIfNotExists([]byte(nextIdBucketName))
		if err != nil {
			return err
		}
		nextIdBytes := bucket.Get([]byte(nextIdKey))
		// Key doesn't exist yet, create it.
		if nextIdBytes == nil {
			nextId = 1
		} else {
			nextId = models.JacuzziIdFromBytes(nextIdBytes)
		}
		return bucket.Put([]byte(nextIdKey), models.BytesFromJacuzziId(nextId+1))
	})

	if err != nil {
		log.Println("Error interacting with database:", err)
	} else {
		log.Printf("JacuzziId %d claimed!\n", nextId)
	}
	return nextId, err
}

// Claim the next available JacuzziId. Internally increments the next available value.
func ClaimNextJacuzziId(serverId string) (models.JacuzziId, error) {
	var nextId models.JacuzziId

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nextId, err
	}
	defer db.Close()

	return claimNextJacuzziId(db)
}
