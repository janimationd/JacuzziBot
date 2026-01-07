package db

import (
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
)

const voiceCallsBucketName string = "VoiceCalls"

func getOrCreateVoiceCall(db *bolt.DB, channelId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall
	var voiceCallJson []byte

	err := db.Update(func(tx *bolt.Tx) error {
		// Create or fetch the VoiceCalls bucket
		bucket, err := tx.CreateBucketIfNotExists([]byte(voiceCallsBucketName))
		if err != nil {
			return err
		}
		voiceCallJson = bucket.Get([]byte(channelId))
		return nil
	})

	if err != nil {
		log.Println("Error reading from database:", err)
		return voiceCall, err
	}

	if voiceCallJson == nil {
		// ChannelId key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(voiceCallsBucketName))
			if err != nil {
				return err
			}
			voiceCall = models.VoiceCall{
				ChannelId: channelId,
				Users:     &utils.Set[string]{},
			}
			voiceCallJson, err = models.ToJsonBytes(voiceCall)
			if err != nil {
				return err
			}
			return bucket.Put([]byte(channelId), voiceCallJson)
		})
	}

	if voiceCallJson == nil || err != nil {
		log.Println("Unknown error, could not fetch or create voice call record: ", err)
		return voiceCall, err
	}

	// Load embedded JSON
	voiceCall, err = models.FromJsonBytes[models.VoiceCall](voiceCallJson)
	if err != nil {
		log.Println("Error unmarshalling voice call: ", err)
	}

	return voiceCall, err
}

type operation bool

const (
	Add    operation = true
	Remove operation = false
)

func modifyUsers(db *bolt.DB, channelId string, op operation, userId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall

	// Encapsulate all logic in a transaction to avoid race conditions
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(voiceCallsBucketName))
		if err != nil {
			return err
		}
		voiceCallJson := bucket.Get([]byte(channelId))
		if voiceCallJson == nil {
			switch op {
			case Add:
				{
					users := utils.NewSet[string]()
					users.Add(userId)
					// Voice call is not in database yet, so create a new record for it
					voiceCall = models.VoiceCall{
						ChannelId: channelId,
						Users:     users,
					}
				}
			case Remove:
				{
					log.Printf("Can't remove user from voice call, no known call in channel %s.\n", channelId)
					// Swallow the error for now, not sure what to do if we surfaced this.
					return nil
				}
			}
		} else {
			// Voice call is in database already, so update its record
			voiceCall, err = models.FromJsonBytes[models.VoiceCall](voiceCallJson)
			if err != nil {
				log.Println("Error unmarshalling voice call: ", err)
				return err
			}
			switch op {
			case Add:
				added := voiceCall.Users.Add(userId)
				if !added {
					log.Printf("Voice call for channel %s already had user %s in it; didn't add.\n", channelId, userId)
				}
			case Remove:
				removed := voiceCall.Users.Remove(userId)
				if !removed {
					log.Printf("Voice call for channel %s didn't have user %s in it; didn't remove.\n", channelId, userId)
				}
			}
		}

		voiceCallJson, err = models.ToJsonBytes(voiceCall)
		if err != nil {
			log.Println("Error marshalling voice call: ", err)
			return err
		}
		return bucket.Put([]byte(channelId), voiceCallJson)
	})

	if voiceCall.ChannelId == "" || err != nil {
		log.Println("Unknown error, could not create or update voice call record: ", err)
		return voiceCall, err
	}

	return voiceCall, err
}

func getVoiceCallForUser(db *bolt.DB, userId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall

	// Encapsulate all logic in a transaction to avoid race conditions
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(voiceCallsBucketName))
		if bucket == nil {
			return nil
		}
		err := bucket.ForEach(func(k, jsonBytes []byte) error {
			vc, err := models.FromJsonBytes[models.VoiceCall](jsonBytes)
			if err != nil {
				log.Println("Error unmarshalling voice call, skipping:", err)
				return err
			}
			if vc.ContainsUser(userId) {
				voiceCall = vc
				return nil
			}
			return nil
		})

		return err
	})

	return voiceCall, err
}

func getAllVoiceCallsWithParticipants(db *bolt.DB) ([]models.VoiceCall, error) {
	var voiceCalls []models.VoiceCall

	// Encapsulate all logic in a transaction to avoid race conditions
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(voiceCallsBucketName))
		if bucket == nil {
			return nil
		}
		err := bucket.ForEach(func(k, jsonBytes []byte) error {
			vc, err := models.FromJsonBytes[models.VoiceCall](jsonBytes)
			if err != nil {
				log.Println("Error unmarshalling voice call, skipping:", err)
				return err
			}
			if vc.Users.Size() != 0 {
				voiceCalls = append(voiceCalls, vc)
			}
			return nil
		})

		return err
	})

	return voiceCalls, err
}

func GetVoiceCall(serverId string, channelId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return voiceCall, err
	}
	defer db.Close()

	return getOrCreateVoiceCall(db, channelId)
}

func ModifyUsers(serverId string, channelId string, op operation, userId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return voiceCall, err
	}
	defer db.Close()

	return modifyUsers(db, channelId, op, userId)
}

func GetVoiceCallForUser(serverId string, userId string) (models.VoiceCall, error) {
	var voiceCall models.VoiceCall

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return voiceCall, err
	}
	defer db.Close()

	return getVoiceCallForUser(db, userId)
}

// Get all active voice calls that have at least one user in them right now.
func GetAllVoiceCallsWithParticipants(serverId string) ([]models.VoiceCall, error) {
	var voiceCalls []models.VoiceCall

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return voiceCalls, err
	}
	defer db.Close()

	return getAllVoiceCallsWithParticipants(db)
}

func NukeVoiceCallBucket(serverId string) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		log.Println("Failed to nuke voice call database")
		return
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(voiceCallsBucketName))
	})
}
