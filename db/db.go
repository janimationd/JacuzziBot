package db

import (
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/janimationd/JacuzziBot/models"
)

const dbPath string = "db/"
const usersBucketName string = "Users"

// Open or create a BoltDB database for the given server ID.
// Calling code MUST handle closing the database.
func getDb(serverId string) (*bolt.DB, error) {
	// Create or open a server-specific database file
	db, err := bolt.Open(dbPath+serverId+".db", 0600, nil)
	if err != nil {
		log.Println("Error opening database: ", err)
		return nil, err
	}
	return db, nil
}

func getOrCreateUser(db *bolt.DB, userId string) (models.User, error) {
	var user models.User
	var userJson []byte

	err := db.Update(func(tx *bolt.Tx) error {
		// Create or fetch the Users bucket
		b, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
		if err != nil {
			return err
		}
		userJson = b.Get([]byte(userId))
		return nil
	})

	if err != nil {
		log.Println("Error reading from database:", err)
		return user, err
	}

	if userJson == nil {
		// User key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
			if err != nil {
				return err
			}
			user = models.User{
				UserId: userId,
				Points: 0,
			}
			userJson, err = user.ToJsonBytes()
			if err != nil {
				return err
			}
			return b.Put([]byte(userId), userJson)
		})
	}

	if userJson == nil || err != nil {
		log.Println("Unknown error, could not fetch or create user record: ", err)
		return user, err
	}

	// Load embedded JSON
	user, err = models.FromJsonBytes(userJson)
	if err != nil {
		log.Println("Error unmarshalling user: ", err)
	}

	return user, err
}

func modifyUserPoints(db *bolt.DB, userId string, pointsDelta float64) (models.User, error) {
	var user models.User
	var userJson []byte
	var err error

	if userJson == nil {
		// User key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
			if err != nil {
				return err
			}
			userJson = b.Get([]byte(userId))
			if userJson == nil {
				// User is not in database yet, so create a new record for them
				user = models.User{
					UserId: userId,
					Points: pointsDelta,
				}
			} else {
				// User is in database already, so update their record
				user, err = models.FromJsonBytes(userJson)
				if err != nil {
					log.Println("Error unmarshalling user: ", err)
					return err
				}
				user.Points += pointsDelta
			}

			userJson, err = user.ToJsonBytes()
			if err != nil {
				log.Println("Error marshalling user: ", err)
				return err
			}
			return b.Put([]byte(userId), userJson)
		})
	}

	if user == (models.User{}) || err != nil {
		log.Println("Unknown error, could not create or update user record: ", err)
		return user, err
	}

	return user, err
}

func GetUser(serverId string, userId string) (models.User, error) {
	var user models.User

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return user, err
	}
	defer db.Close()

	return getOrCreateUser(db, userId)
}

func ModifyUserPoints(serverId string, userId string, pointsDelta float64) (models.User, error) {
	var user models.User

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return user, err
	}
	defer db.Close()

	return modifyUserPoints(db, userId, pointsDelta)
}
