package db

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type User struct {
	UserId string
	Points int
}

const UsersBucketName string = "Users"

// Open or create a BoltDB database for the given server ID.
// Calling code MUST handle closing the database.
func getDb(serverId string) (*bolt.DB, error) {
	// Create or open a server-specific database file
	db, err := bolt.Open(serverId+".db", 0600, nil)
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return nil, err
	}
	return db, nil
}

func getOrCreateUser(db *bolt.DB, userId string) (User, error) {
	var user User
	var userJson []byte

	err := db.Update(func(tx *bolt.Tx) error {
		// Create or fetch the Users bucket
		b, err := tx.CreateBucketIfNotExists([]byte(UsersBucketName))
		if err != nil {
			return err
		}
		userJson = b.Get([]byte(userId))
		return nil
	})

	if err != nil {
		fmt.Println("Error reading from database:", err)
		return user, err
	}

	if userJson == nil {
		// User key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(UsersBucketName))
			if err != nil {
				return err
			}
			user = User{
				UserId: userId,
				Points: 0,
			}
			userJson, err = json.Marshal(user)
			if err != nil {
				return err
			}
			return b.Put([]byte(userId), userJson)
		})
	}

	if userJson == nil || err != nil {
		fmt.Println("Unknown error, could not fetch or create user record: ", err)
		return user, err
	}

	// Load embedded JSON
	err = json.Unmarshal(userJson, &user)
	if err != nil {
		fmt.Println("Error unmarshalling user: ", err)
	}

	return user, err
}

func awardUserPoints(db *bolt.DB, userId string, points int) (User, error) {
	var user User
	var userJson []byte
	var err error

	if userJson == nil {
		// User key not found in bucket, create a record for them
		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(UsersBucketName))
			if err != nil {
				return err
			}
			userJson = b.Get([]byte(userId))
			if userJson == nil {
				// User is not in database yet, so create a new record for them
				user = User{
					UserId: userId,
					Points: points,
				}
			} else {
				// User is in database already, so update their record
				err = json.Unmarshal(userJson, &user)
				if err != nil {
					fmt.Println("Error unmarshalling user: ", err)
					return err
				}
				user.Points += points
			}

			userJson, err = json.Marshal(user)
			if err != nil {
				fmt.Println("Error marshalling user: ", err)
				return err
			}
			return b.Put([]byte(userId), userJson)
		})
	}

	if user == (User{}) || err != nil {
		fmt.Println("Unknown error, could not create or update user record: ", err)
		return user, err
	}

	return user, err
}

func GetUser(serverId string, userId string) (User, error) {
	var user User

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return user, err
	}
	defer db.Close()

	return getOrCreateUser(db, userId)
}

func AwardUserPoints(serverId string, userId string, points int) (User, error) {
	var user User

	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return user, err
	}
	defer db.Close()

	return awardUserPoints(db, userId, points)
}
