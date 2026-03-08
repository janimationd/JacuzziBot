package db

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/models"
	bolt "go.etcd.io/bbolt"
)

const tamaChannelBucketName string = "TamaChannel"
const tamaChannelKey string = tamaChannelBucketName
const tamaMinigameRoleBucketName string = "TamaMinigameRole"
const tamaMinigameRoleKey string = tamaMinigameRoleBucketName
const tamaBucketName string = "Tamas"
const tamaTransfersBucketName string = "TamaTransfers"

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
		return fmt.Errorf("Tama must be created in the registered channel <#%s>.", registeredChannelId)
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}
		tamaJson, err := json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON: %w", err)
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

func changeTamaOwner(
	db *bolt.DB,
	tamaId models.JacuzziId,
	newOwnerId string,
	replaceOwner bool,
) (*models.Tama, error) {
	tama := &models.Tama{}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}
		var tamaBytes []byte
		if bucket != nil {
			tamaBytes = bucket.Get(models.BytesFromJacuzziId(tamaId))
		}
		if tamaBytes == nil {
			return fmt.Errorf("Tama doesn't exist.")
		}

		err = json.Unmarshal(tamaBytes, tama)
		if err != nil {
			return fmt.Errorf("Could not unmarshall Tama JSON")
		}
		// Check whether we're not allowed to replace the owner
		if tama.IsOwned() && !replaceOwner {
			return fmt.Errorf("Tama is already owned by <@%s>, cannot overwrite", tama.Owner)
		}
		tama.Owner = newOwnerId
		tamaBytes, err = json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON: %w", err)
		}
		return bucket.Put(models.BytesFromJacuzziId(tama.Id), tamaBytes)
	})

	if err != nil {
		log.Println("Couldn't change Tama owner:", err)
		return nil, err
	}

	return tama, nil
}

func nameTama(
	db *bolt.DB,
	tamaId models.JacuzziId,
	newName string,
) (*models.Tama, error) {
	tama := &models.Tama{}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}
		var tamaBytes []byte
		if bucket != nil {
			tamaBytes = bucket.Get(models.BytesFromJacuzziId(tamaId))
		}
		if tamaBytes == nil {
			return fmt.Errorf("Tama doesn't exist.")
		}

		// Ensure the uniqueness of the name across all Tamas.
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			otherId := models.JacuzziIdFromBytes(k)
			// Skip this Tama
			if otherId == tamaId {
				continue
			}
			// Ignore errors
			err := json.Unmarshal(v, tama)
			if err == nil && tama.Name == newName {
				return fmt.Errorf("Name already in use on Tama %d.", otherId)
			}
		}

		err = json.Unmarshal(tamaBytes, tama)
		if err != nil {
			return fmt.Errorf("Could not unmarshall Tama JSON")
		}

		tama.Name = newName
		tamaBytes, err = json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON: %w", err)
		}
		return bucket.Put(models.BytesFromJacuzziId(tama.Id), tamaBytes)
	})

	if err != nil {
		log.Println("Couldn't name Tama:", err)
		return nil, err
	}

	return tama, nil
}

func careForTama(
	db *bolt.DB,
	tamaId models.JacuzziId,
	userTimezone *time.Location,
) (*models.Tama, bool, bool, error) {
	tama := &models.Tama{}
	hatched := false
	modified := false

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}

		idBytes := models.BytesFromJacuzziId(tamaId)
		tamaBytes := bucket.Get(idBytes)
		if tamaBytes == nil {
			return fmt.Errorf("Tama doesn't exist.")
		}

		err = json.Unmarshal(tamaBytes, tama)
		if err != nil {
			return fmt.Errorf("Could not unmarshall Tama JSON")
		}

		if !tama.IsAlive() {
			return fmt.Errorf("Tama %s is dead, and can no longer be cared for.", tama.GetNameAndId())
		}

		nextCareTime := tama.GetNextCareTime()
		if time.Now().Before(nextCareTime) {
			return fmt.Errorf("This action is still cooling down.\n\nYou will be able to perform it again at `%s`.",
				nextCareTime.In(userTimezone).String())
		}

		tama.LastCareTime = time.Now().Unix()
		if tama.IsEgg() {
			tama.EggCareCount++
			// Time to hatch!
			if tama.EggCareCount >= constants.EggCareHatchThreshold {
				tama.Hatch()
				hatched = true
			}
		} else {
			modified = tama.ModifyMood(1)
		}

		// Save it back to the DB
		tamaBytes, err = json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON: %w", err)
		}

		err = bucket.Put(idBytes, tamaBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, false, false, err
	}

	return tama, modified, hatched, nil
}

func feedTama(db *bolt.DB, tamaId models.JacuzziId) (*models.Tama, error) {
	tama := &models.Tama{}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaBucketName))
		if err != nil {
			return err
		}

		idBytes := models.BytesFromJacuzziId(tamaId)
		tamaBytes := bucket.Get(idBytes)
		if tamaBytes == nil {
			return fmt.Errorf("Tama doesn't exist.")
		}

		err = json.Unmarshal(tamaBytes, tama)
		if err != nil {
			return fmt.Errorf("Could not unmarshall Tama JSON")
		}

		if !tama.IsAlive() {
			return fmt.Errorf("Tama %s is dead, and can no longer be fed.", tama.GetNameAndId())
		}

		if tama.Hunger == 0 {
			return fmt.Errorf("This Tama is already full. You don't need to feed it again until tomorrow (your local time).")
		}

		// Reduce its hunger
		tama.Hunger -= 1

		// Save it back to the DB
		tamaBytes, err = json.Marshal(tama)
		if err != nil {
			return fmt.Errorf("Could not marshall Tama to JSON: %w", err)
		}

		err = bucket.Put(idBytes, tamaBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tama, nil
}

func getAllTamasOwnedByUser(db *bolt.DB, userId string) map[models.JacuzziId]*models.Tama {
	tamas := make(map[models.JacuzziId]*models.Tama)

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaBucketName))
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			tamaId := models.JacuzziIdFromBytes(k)
			if tamaId == models.NoId {
				log.Printf("Invalid Tama ID DB key: %s\n", string(k))
				return nil
			}

			tama := &models.Tama{}
			err := json.Unmarshal(v, tama)
			if err != nil {
				log.Printf("Invalid Tama JSON bytes: %s\n", string(v))
				return nil
			}

			if tama.Owner == userId {
				// Populate the map
				tamas[tamaId] = tama
			}
			return nil
		})

		return nil
	})

	return tamas
}

func getTamaMinigameRole(db *bolt.DB) string {
	var minigameRoleId string

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaMinigameRoleBucketName))
		if bucket == nil {
			return fmt.Errorf("No role for Tama minigame yet.")
		}
		minigameRoleBytes := bucket.Get([]byte(tamaMinigameRoleKey))
		if minigameRoleBytes == nil {
			return fmt.Errorf("No role for Tama minigame yet.")
		}
		minigameRoleId = string(minigameRoleBytes)
		return nil
	})
	if minigameRoleId == "" {
		log.Println("Could not fetch Tama minigame role:", err)
	}
	return minigameRoleId
}

func registerTamaMinigameRole(db *bolt.DB, roleId string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaMinigameRoleBucketName))
		if err != nil {
			log.Println("Couldn't fetch/create minigame role bucket:", err)
			return err
		}
		minigameRoleIdBytes := bucket.Get([]byte(tamaMinigameRoleKey))
		if minigameRoleIdBytes != nil {
			return fmt.Errorf("Minigame role <@%s> already exists.", string(minigameRoleIdBytes))
		}
		return bucket.Put([]byte(tamaMinigameRoleKey), []byte(roleId))
	})

	if err != nil {
		log.Println("Could not register Tama channel:", err)
	}
	return err
}

func createTamaTransfer(
	db *bolt.DB,
	tamaId models.JacuzziId,
	oldOwnerId string,
	newOwnerId string,
) (*models.TamaTransfer, error) {
	transfer := &models.TamaTransfer{
		TamaId:     tamaId,
		OldOwnerId: oldOwnerId,
		NewOwnerId: newOwnerId,
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(tamaTransfersBucketName))
		if err != nil {
			return err
		}

		transferBytes := bucket.Get(models.BytesFromJacuzziId(tamaId))
		if transferBytes != nil {
			err = json.Unmarshal(transferBytes, transfer)
			if err != nil {
				return err
			}

			return fmt.Errorf("A pending transfer for Tama #%d already exists: <@%s> -> <@%s>. "+
				"You'll have to cancel it before making a new one.",
				tamaId, transfer.OldOwnerId, transfer.NewOwnerId)
		}

		transferBytes, err = json.Marshal(transfer)
		if err != nil {
			return err
		}

		err = bucket.Put(models.BytesFromJacuzziId(tamaId), transferBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Couldn't create Tama transfer: %w", err)
	}

	return transfer, nil
}

func getTamaTransfer(db *bolt.DB, tamaId models.JacuzziId) (*models.TamaTransfer, error) {
	transfer := &models.TamaTransfer{}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaTransfersBucketName))
		if bucket == nil {
			return fmt.Errorf("%s bucket doesn't exist on get%s",
				tamaTransfersBucketName, constants.ErrorReportMessageSuffix)
		}

		transferBytes := bucket.Get(models.BytesFromJacuzziId(tamaId))
		if transferBytes == nil {
			return nil
		}

		err := json.Unmarshal(transferBytes, transfer)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Couldn't get Tama transfer: %w", err)
	}

	return transfer, nil
}

func deleteTamaTransfer(db *bolt.DB, tamaId models.JacuzziId) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tamaTransfersBucketName))
		if bucket == nil {
			return fmt.Errorf("%s bucket doesn't exist on delete%s",
				tamaTransfersBucketName, constants.ErrorReportMessageSuffix)
		}

		return bucket.Delete(models.BytesFromJacuzziId(tamaId))
	})
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

// Get the Tama from the database. Returns a descriptive error if the Tama doesn't exist.
func GetTama(serverId string, tamaId models.JacuzziId) (*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getTama(db, tamaId)
}

// If replaceOwner is false, will return an error when the Tama is already owned by someone.
func ChangeTamaOwner(
	serverId string,
	tamaId models.JacuzziId,
	newOwnerId string,
	replaceOwner bool,
) (*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return changeTamaOwner(db, tamaId, newOwnerId, replaceOwner)
}

func NameTama(
	serverId string,
	tamaId models.JacuzziId,
	newName string,
) (*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return nameTama(db, tamaId, newName)
}

// Care for the Tama.
// If its mood is updated, returns true for the second return value (false indicates it was already at max mood).
// If it is an egg, this might cause it to hatch, in which case we return true for the third return value.
func CareForTama(
	serverId string,
	tamaId models.JacuzziId,
	userTimezone *time.Location,
) (*models.Tama, bool, bool, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, false, false, err
	}
	defer db.Close()

	return careForTama(db, tamaId, userTimezone)
}

// Feed the Tama one food. Assumes you've already sold food to the Tama's owner successfully.
func FeedTama(serverId string, tamaId models.JacuzziId) (*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return feedTama(db, tamaId)
}

// Get a set of all Tamas owned by the user.
func GetAllTamasOwnedByUser(serverId string, userId string) (map[models.JacuzziId]*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getAllTamasOwnedByUser(db, userId), nil
}

func GetTamaMinigameRole(serverId string) string {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return ""
	}
	defer db.Close()

	return getTamaMinigameRole(db)
}

func RegisterTamaMinigameRole(serverId string, minigameRoleId string) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	return registerTamaMinigameRole(db, minigameRoleId)
}

func CreateTamaTransfer(
	serverId string,
	tamaId models.JacuzziId,
	oldOwnerId string,
	newOwnerId string,
) (*models.TamaTransfer, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return createTamaTransfer(db, tamaId, oldOwnerId, newOwnerId)
}

func GetTamaTransfer(serverId string, tamaId models.JacuzziId) (*models.TamaTransfer, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getTamaTransfer(db, tamaId)
}

func DeleteTamaTransfer(serverId string, tamaId models.JacuzziId) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	return deleteTamaTransfer(db, tamaId)
}
