package db

import (
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/janimationd/JacuzziBot/errors"
	"github.com/janimationd/JacuzziBot/models"
)

const usersBucketName string = "Users"

func getOrCreateUser(db *bolt.DB, userId string) (models.User, error) {
	var user models.User
	var userJson []byte

	err := db.Update(func(tx *bolt.Tx) error {
		// Create or fetch the Users bucket
		bucket, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
		if err != nil {
			return err
		}
		userJson = bucket.Get([]byte(userId))
		return nil
	})

	if err != nil {
		log.Println("Error reading from database:", err)
		return user, err
	}

	if userJson == nil {
		// User key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
			if err != nil {
				return err
			}
			user = models.User{
				UserId: userId,
				Points: 0,
			}
			userJson, err = models.ToJsonBytes(user)
			if err != nil {
				return err
			}
			return bucket.Put([]byte(userId), userJson)
		})
	}

	if userJson == nil || err != nil {
		log.Println("Unknown error, could not fetch or create user record: ", err)
		return user, err
	}

	// Load embedded JSON
	user, err = models.FromJsonBytes[models.User](userJson)
	if err != nil {
		log.Println("Error unmarshalling user: ", err)
	}

	return user, err
}

func setUserTimezone(db *bolt.DB, userId string, timezone string) (models.User, error) {
	var user models.User

	// Encapsulate all logic in a transaction to avoid race conditions
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
		if err != nil {
			return err
		}
		userJson := bucket.Get([]byte(userId))

		if userJson == nil {
			// User is not in database yet, so create a new record for them
			user = models.User{
				UserId: userId,
			}
		} else {
			// User is in database already, so update their record
			user, err = models.FromJsonBytes[models.User](userJson)
			if err != nil {
				log.Println("Error unmarshalling user: ", err)
				return err
			}
		}
		user.Timezone = timezone

		userJson, err = models.ToJsonBytes(user)
		if err != nil {
			log.Println("Error marshalling user: ", err)
			return err
		}
		err = bucket.Put([]byte(userId), userJson)
		if err == nil {
			log.Printf("User %s's timezone was saves as %s.\n", user.UserId, timezone)
		}
		return err
	})

	if user == (models.User{}) || err != nil {
		log.Println("Could not create or update user record: ", err)
	}

	return user, err
}

func modifyUserPoints(db *bolt.DB, userId string, pointsDelta float64) (models.User, error) {
	var user models.User

	// Encapsulate all logic in a transaction to avoid race conditions
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(usersBucketName))
		if err != nil {
			return err
		}
		userJson := bucket.Get([]byte(userId))
		var originalPoints float64
		if userJson == nil {
			// User is not in database yet, so create a new record for them
			user = models.User{
				UserId: userId,
				Points: pointsDelta,
			}
			originalPoints = 0
		} else {
			// User is in database already, so update their record
			user, err = models.FromJsonBytes[models.User](userJson)
			if err != nil {
				log.Println("Error unmarshalling user: ", err)
				return err
			}
			originalPoints = user.Points
			user.Points += pointsDelta
		}

		// If their new point value would be negative, reject the action.
		if user.Points < 0 {
			return &errors.InsufficientPointsError{
				CurrentPoints: originalPoints,
				// Since (hopefully) the only way we would arrive at a negative value is a negative delta,
				// here we invert it to get a positive value.
				RequiredPoints: -pointsDelta,
			}
		}

		userJson, err = models.ToJsonBytes(user)
		if err != nil {
			log.Println("Error marshalling user: ", err)
			return err
		}
		err = bucket.Put([]byte(userId), userJson)
		if err == nil {
			log.Printf("User %s's points were modified by %.2f.\n", user.UserId, pointsDelta)
		}
		return err
	})

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

func SetUserTimezone(serverId string, userId string, timezone string) (models.User, error) {
	var user models.User

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return user, err
	}
	defer db.Close()

	return setUserTimezone(db, userId, timezone)
}
