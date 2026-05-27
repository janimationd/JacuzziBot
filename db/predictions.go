package db

import (
	"encoding/json"
	"fmt"

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

func getPrediction(db *bbolt.DB, predictionId models.JacuzziId) (*models.Prediction, error) {
	prediction := &models.Prediction{}

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(predictionsBucketName))
		if bucket == nil {
			return fmt.Errorf("Couldn't close prediction betting because predictions bucket didn't exist")
		}
		predictionBytes := bucket.Get(models.BytesFromJacuzziId(predictionId))
		if predictionBytes == nil {
			return fmt.Errorf("Couldn't close prediction betting because prediction %d didn't exist", predictionId)
		}

		return json.Unmarshal(predictionBytes, prediction)
	})

	if err != nil {
		return nil, err
	}

	return prediction, nil
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

func GetPrediction(serverId string, predictionId models.JacuzziId) (*models.Prediction, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getPrediction(db, predictionId)
}
