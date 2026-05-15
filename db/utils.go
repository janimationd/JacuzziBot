package db

import (
	"log"

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
