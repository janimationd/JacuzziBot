package db

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/janimationd/JacuzziBot/utils"
	bolt "go.etcd.io/bbolt"
)

const dbPath string = "db/"

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

type Operation bool

// The possible operations for a database modification
const (
	Add    Operation = true
	Remove Operation = false
)

func dumpBucketToFile(db *bolt.DB, bucketName string, filePath string) error {
	data := make(map[string][]byte)

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		// It's okay if the bucket doesn't exist
		if bucket == nil {
			return nil
		}

		bucket.ForEach(func(k, v []byte) error {
			data[string(k)] = slices.Clone(v)
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(data)
}

func wipeBucket(db *bolt.DB, bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(bucketName))
		return nil
	})
}

// Returns whether the bucket was restored. False is okay if the bucket didn't exist at the time of the backup.
func restoreBucketFromFile(db *bolt.DB, bucketName string, filePath string) (bool, error) {
	restored := false

	file, err := os.Open(filePath)
	// It's okay if the file doesn't exist, that just means that bucket didn't exist at the time of the backup.
	if err != nil {
		return restored, nil
	}

	data := make(map[string][]byte)
	err = gob.NewDecoder(file).Decode(&data)
	if err != nil {
		return restored, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// First wipe the bucket if it exists
		tx.DeleteBucket([]byte(bucketName))
		// Then recreate it
		bucket, err := tx.CreateBucket([]byte(bucketName))
		if err != nil {
			return err
		}
		// Then fill it with data from the backup
		for k, v := range data {
			err = bucket.Put([]byte(k), v)
			if err != nil {
				return err
			}
		}
		restored = true
		return nil
	})
	return restored, err
}

const backupDir = dbPath + "backups/"
const backupExt = ".bak"

func GetBackupsDirForServer(serverId string) string {
	return backupDir + serverId + "/"
}

func backupBuckets(serverId string, bucketNames []string) (string, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return "", err
	}
	defer db.Close()

	backupTime := time.Now().Format(utils.EventIdTimeFormat)
	path := GetBackupsDirForServer(serverId) + backupTime + "/"

	for _, bucketName := range bucketNames {
		fileName := bucketName + backupExt
		err = dumpBucketToFile(db, bucketName, path+fileName)
		if err != nil {
			return "", err
		}
	}
	return backupTime, nil
}

func wipeBuckets(serverId string, bucketNames []string) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, bucketName := range bucketNames {
		err = wipeBucket(db, bucketName)
		if err != nil {
			return err
		}
	}
	return nil
}

func restoreBucketsFromBackup(serverId string, backupTime string, bucketNames []string) error {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return err
	}
	defer db.Close()

	path := backupDir + serverId + "/" + backupTime + "/"

	restoredAnyBuckets := false
	for _, bucketName := range bucketNames {
		fileName := bucketName + backupExt
		restored, err := restoreBucketFromFile(db, bucketName, path+fileName)
		if err != nil {
			return err
		}
		restoredAnyBuckets = restoredAnyBuckets || restored
	}
	if restoredAnyBuckets {
		return nil
	} else {
		return fmt.Errorf(
			"Didn't restore any buckets! The backup timestamp might be invalid or the backup might be empty.")
	}
}
