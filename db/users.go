package db

import (
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/janimationd/JacuzziBot/errs"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
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
		if userJson == nil {
			user = models.User{
				UserId: userId,
				Points: 0,
			}
			userJson, err = models.ToJsonBytes(user)
			if err != nil {
				return err
			}
			return bucket.Put([]byte(userId), userJson)
		} else {
			// Load embedded JSON
			user, err = models.FromJsonBytes[models.User](userJson)
			if err != nil {
				log.Println("Error unmarshalling user: ", err)
			}
			return nil
		}
	})

	if userJson == nil || err != nil {
		log.Println("Unknown error, could not fetch or create user record: ", err)
		return user, err
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

func modifyUserPoints(db *bolt.DB, userId string, pointsDelta float64, allowDebt bool) (models.User, error) {
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
				Points: 0,
			}
		} else {
			// User is in database already, so update their record
			user, err = models.FromJsonBytes[models.User](userJson)
			if err != nil {
				log.Println("Error unmarshalling user: ", err)
				return err
			}
		}
		originalPoints = user.Points
		user.Points += pointsDelta

		// If their new point value would be more negative, and debt isn't allowed, reject the action.
		if pointsDelta < 0 && user.Points < 0 && !allowDebt {
			return &errs.InsufficientPointsError{
				CurrentPoints:  originalPoints,
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
			log.Printf("User %s's points were modified by %s, now at %s.\n",
				user.UserId, utils.FormatUIFloat(pointsDelta), utils.FormatUIFloat(user.Points))
		}
		return err
	})

	if err != nil {
		log.Println("Unknown error, could not create or update user record:", err)
		return user, err
	}

	return user, err
}

func modifyUserFlairLevel(
	db *bolt.DB,
	userId string,
	expectedCurrentFlair models.FlairLevel,
	targetFlair models.FlairLevel,
) (models.User, float64, error) {
	var pointsDiff float64
	var user models.User

	if targetFlair >= models.FlairMax || targetFlair < models.FlairNone {
		return user, pointsDiff, fmt.Errorf("Target flair %d is invalid.", targetFlair)
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(usersBucketName))
		if bucket == nil {
			return fmt.Errorf("Users bucket didn't exist when trying to modify user flair")
		}
		userBytes := bucket.Get([]byte(userId))
		var err error
		user, err = models.FromJsonBytes[models.User](userBytes)
		if err != nil {
			return err
		}

		currentFlair := user.Flair
		if expectedCurrentFlair != currentFlair {
			return fmt.Errorf("Your flair level is different than it was when you initiated the request. " +
				"Please cancel the request and start another.")
		}
		if currentFlair == targetFlair {
			return fmt.Errorf("Your flair is already at the requested level.")
		}

		currentFlairProps := models.FlairProps[currentFlair]
		targetFlairProps := models.FlairProps[targetFlair]
		pointsDiff = currentFlairProps.TotalPointCost - targetFlairProps.TotalPointCost

		// They will be paying points to upgrade, so make sure they have enough
		if pointsDiff < 0 && user.Points+pointsDiff < 0 {
			return fmt.Errorf("You need %s point%s to upgrade your flair, but you only have %s!",
				utils.FormatUIFloat(-pointsDiff), utils.Plural(-pointsDiff), utils.FormatUIFloat(user.Points))
		}

		user.Points += pointsDiff
		user.Flair = targetFlair

		userBytes, err = models.ToJsonBytes(user)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(userId), userBytes)
	})

	if err != nil {
		log.Printf("Couldn't modify user flair: %s\n", err.Error())
	}

	return user, pointsDiff, err
}

func GetUser(serverId string, userId string) (models.User, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return models.User{}, err
	}
	defer db.Close()

	return getOrCreateUser(db, userId)
}

// Don't allow debt
func ModifyUserPoints(serverId string, userId string, pointsDelta float64) (models.User, error) {
	return ModifyUserPointsWithDebt(serverId, userId, pointsDelta, false)
}

// Maybe allow debt
func ModifyUserPointsWithDebt(serverId string, userId string, pointsDelta float64, allowDebt bool) (models.User, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return models.User{}, err
	}
	defer db.Close()

	return modifyUserPoints(db, userId, pointsDelta, allowDebt)
}

func SetUserTimezone(serverId string, userId string, timezone string) (models.User, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return models.User{}, err
	}
	defer db.Close()

	return setUserTimezone(db, userId, timezone)
}

func ModifyUserFlairLevel(
	serverId string,
	userId string,
	expectedCurrentFlair models.FlairLevel,
	targetFlair models.FlairLevel,
) (models.User, float64, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return models.User{}, 0, err
	}
	defer db.Close()

	return modifyUserFlairLevel(db, userId, expectedCurrentFlair, targetFlair)
}
