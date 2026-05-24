package db

import (
	"encoding/json"

	"github.com/janimationd/JacuzziBot/models"
	"go.etcd.io/bbolt"
)

const predictionsBucketName = "predictions"

func storePrediction(db *bbolt.DB, prediction *models.Prediction) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(predictionsBucketName))
		if err != nil {
			return err
		}
		predictionBytes, err := json.Marshal(prediction)
		if err != nil {
			return err
		}
		return bucket.Put(models.BytesFromJacuzziId(prediction.Id), predictionBytes)
	})
}

func StorePrediction(serverId string, prediction *models.Prediction) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	return storePrediction(db, prediction)
}
