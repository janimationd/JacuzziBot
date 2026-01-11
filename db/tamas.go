package db

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/janimationd/JacuzziBot/models"
	bolt "go.etcd.io/bbolt"
)

const tamaChannelBucketName string = "TamaChannel"
const tamaChannelKey string = tamaChannelBucketName
const tamaBucketName string = "Tamas"

func registerTamaChannel(db *bolt.DB, channelId string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaChannelBucketName))
		if err != nil {
			log.Println("Couldn't fetch/create channel bucket:", err)
			return err
		}
		channelIdBytes := bucket.Get([]byte(tamaChannelKey))
		if channelIdBytes != nil {
			return fmt.Errorf("Channel <#%s> is already registered as a Tama minigame.", string(channelIdBytes))
		}
		return bucket.Put([]byte(tamaChannelKey), []byte(channelId))
	})

	if err != nil {
		log.Println("Could not register Tama channel:", err)
	}
	return err
}

func getTamaChannel(db *bolt.DB) string {
	var channelId string

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaChannelBucketName))
		if bucket == nil {
			return fmt.Errorf("No channel registered for Tama minigame yet.")
		}
		channelBytes := bucket.Get([]byte(tamaChannelKey))
		if channelBytes == nil {
			return fmt.Errorf("No channel registered for Tama minigame yet.")
		}
		channelId = string(channelBytes)
		return nil
	})
	if channelId == "" {
		log.Println("Could not fetch Tama channel:", err)
	}
	return channelId
}

func storeTama(db *bolt.DB, channelId string, tama *models.Tama) error {
	registeredChannelId := getTamaChannel(db)
	if registeredChannelId == "" {
		return fmt.Errorf("No channel is registered as a Tama minigame yet.")
	}
	if registeredChannelId != channelId {
		return fmt.Errorf("Tama must be created in the registered channel #%s.", registeredChannelId)
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}
		tamaJson, err := json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON")
		}
		return bucket.Put(models.BytesFromJacuzziId(tama.Id), tamaJson)
	})

	if err != nil {
		log.Println("Couldn't store Tama:", err)
		return err
	}

	return nil
}

func getTama(db *bolt.DB, tamaId models.JacuzziId) (*models.Tama, error) {
	tama := &models.Tama{}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaBucketName))
		var tamaBytes []byte
		if bucket != nil {
			tamaBytes = bucket.Get(models.BytesFromJacuzziId(tamaId))
		}
		if tamaBytes == nil {
			return fmt.Errorf("Tama %d doesn't exist.", tamaId)
		}
		return json.Unmarshal(tamaBytes, tama)
	})

	if err != nil {
		log.Println("Couldn't get Tama:", err)
		return nil, err
	}

	return tama, err
}

func RegisterTamaChannel(serverId string, channelId string) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	return registerTamaChannel(db, channelId)
}

func GetTamaChannel(serverId string) string {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return ""
	}
	defer db.Close()

	return getTamaChannel(db)
}

// If the Tama already exists, overwrites it. channelId is the channel the request originated from; if it isn't the
// registered channel on this server then the request fails.
func StoreTama(serverId string, channelId string, tama *models.Tama) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	return storeTama(db, channelId, tama)
}

func GetTama(serverId string, tamaId models.JacuzziId) (*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getTama(db, tamaId)
}
