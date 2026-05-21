package events

import (
	"log"
	"time"

	"github.com/janimationd/JacuzziBot/db"
	"github.com/janimationd/JacuzziBot/discord/session"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
	"go.etcd.io/bbolt"
)

// Once per day backup the entire state of all Tamas minigames.
var DailyTamaBackup = models.ScheduledEvent{
	ID:       "DailyTamaBackup",
	Interval: 24 * time.Hour,
	Handler:  "DailyTamaBackupHandler",
}

func DailyTamaBackupHandler(event *models.ScheduledEvent, _ time.Time, _ *bbolt.Tx) bool {
	if session.Handle == nil {
		log.Println("Discord session is nil, not backing up Tama minigame states.")
		return false
	}

	for _, guild := range session.Handle.State.Guilds {
		serverId := guild.ID

		// Keep the last 60 backups for this server
		err := utils.PruneOldestSubdirs(db.GetBackupsDirForServer(serverId), 60)
		if err != nil {
			log.Printf("Couldn't cleanup old backups for server %s: %s\n", serverId, err.Error())
		}

		backupTime, err := db.BackupTamaBuckets(serverId)
		if err != nil {
			log.Printf("Couldn't perform daily Tamas backup for server %s: %s\n", serverId, err.Error())
		} else {
			log.Printf("Automatically backed up Tamas minigame for server %s at time %s.\n", serverId, backupTime)
		}
	}

	return false
}
