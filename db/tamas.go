package db

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/janimationd/JacuzziBot/constants"
	"github.com/janimationd/JacuzziBot/models"
	"github.com/janimationd/JacuzziBot/utils"
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

		if tama.IsDead() {
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

		if tama.IsDead() {
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

func getAllTamas(db *bolt.DB, userId string, aliveOnly bool, hatchedOnly bool) map[models.JacuzziId]*models.Tama {
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

			// Check filters
			satisfiesOwningUserFilter := userId == "" || tama.Owner == userId
			satisfiesAliveFilter := !aliveOnly || tama.IsAlive()
			satisfiesHatchedFilter := !hatchedOnly || tama.HasHatched()
			if satisfiesOwningUserFilter && satisfiesAliveFilter && satisfiesHatchedFilter {
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

// Choose a random interaction with a very specific probabilty distribution. The second return value will be true if
// instigator is a Bully, and that fact caused it to bully the target.
func chooseRandomInteraction(instigator *models.Tama, target *models.Tama) (models.TamaInteraction, bool) {
	weights := make([]int, models.TamaInteractionMax)
	weights[models.Play] = 5
	weights[models.Gift] = 2
	weights[models.PickOn] = 1

	// Bullies have twice the chance to pick on other pets.
	isABully := instigator.NegativeTraits.Contains(models.Bully)
	if isABully {
		weights[models.PickOn] *= 2
	}

	// When a pet is courting another, it has twice the chance to give gifts.
	if instigator.IsCourting(target) {
		weights[models.Gift] *= 2
	}

	totalWeight := 0
	for i := range models.TamaInteractionMax {
		totalWeight += weights[i]
	}

	roll := rand.IntN(totalWeight)
	for i := range models.TamaInteractionMax {
		if roll < weights[i] {
			// If this is a PickOn from a Bully which wouldn't have happened otherwise.
			if i == models.PickOn && isABully && roll == 1 {
				return i, true
			}
			return i, false
		}
		roll -= weights[i]
	}

	panic(fmt.Sprintf("Shouldn't have gotten here! Remaining roll: %d", roll))
}

func chooseRandomGiftOutcome() models.GiftOutcome {
	weights := make([]int, models.GiftOutcomeMax)
	weights[models.Likes] = 4
	weights[models.Indifferent] = 2
	weights[models.Hates] = 1

	totalWeight := 0
	for i := range models.GiftOutcomeMax {
		totalWeight += weights[i]
	}

	roll := rand.IntN(totalWeight)
	for i := range models.GiftOutcomeMax {
		if roll < weights[i] {
			return i
		}
		roll -= weights[i]
	}

	panic(fmt.Sprintf("Shouldn't have gotten here! Remaining roll: %d", roll))
}

func documentDirectionalInteractionResult(
	this *models.Tama,
	other *models.Tama,
	desiredChange models.RelationshipScore,
	result models.RelationshipScoreModificationResult,
	indent string,
) string {
	var summary string

	// Document primary and secondary effects
	if desiredChange != 0 {
		if desiredChange == result.FinalDelta {
			summary += fmt.Sprintf("\n%s  - %s's attitude towards %s changed by %s%d.",
				indent, this.GetNameAndId(), other.GetNameAndId(),
				utils.SignString(result.FinalDelta), result.FinalDelta)
		} else if result.FinalDelta == 0 {
			if result.LoveEvent == models.LovePreventedDecrease {
				summary += fmt.Sprintf(
					"\n%s  - Since they're in love, %s avoided losing %d attitude towards %s (66%% chance).",
					indent, this.GetNameAndId(), desiredChange, other.GetNameAndId())
			} else {
				var verb string
				var limit models.RelationshipScore
				if desiredChange > 0 {
					verb = "improve"
					limit = models.TamaRelationshipScoreLimit
				} else {
					verb = "worsen"
					limit = -models.TamaRelationshipScoreLimit
				}
				summary += fmt.Sprintf("\n%s  - %s's attitude towards %s cannot %s any more (already at %s%d).",
					indent, this.GetNameAndId(), other.GetNameAndId(), verb, utils.SignString(limit), limit)
			}
		} else {
			var limit models.RelationshipScore
			if desiredChange > 0 {
				limit = models.TamaRelationshipScoreLimit
			} else {
				limit = -models.TamaRelationshipScoreLimit
			}
			summary += fmt.Sprintf("\n%s  - %s's attitude towards %s changed by %s%d (capped at %s%d).",
				indent, this.GetNameAndId(), other.GetNameAndId(),
				utils.SignString(result.FinalDelta), result.FinalDelta,
				utils.SignString(limit), limit)
		}
	}
	if result.FriendlyBonus != 0 {
		summary += fmt.Sprintf(" This included a %s%d bonus because %s has the Friendly trait (33%% chance).",
			utils.SignString(result.FriendlyBonus), result.FriendlyBonus,
			other.GetNameAndId())
	}
	if result.MoodDelta != 0 {
		summary += fmt.Sprintf("\n%s    - Because of this %s's mood also changed by %s%d (33%% chance).",
			indent, this.GetNameAndId(),
			utils.SignString(result.MoodDelta), result.MoodDelta)
	}
	if result.LoveEvent == models.FellInLove {
		summary += fmt.Sprintf("\n%s    - Because of this, %s and %s fell in love! :heart:",
			indent, this.GetNameAndId(), other.GetNameAndId())
	}
	if result.LoveEvent == models.FellOutOfLove {
		summary += fmt.Sprintf("\n%s    - Because of this, %s and %s fell out of love! :broken_heart:",
			indent, this.GetNameAndId(), other.GetNameAndId())
	}

	return summary
}

func tamaInteract(
	db *bolt.DB,
	instigatorId models.JacuzziId,
	targetId models.JacuzziId,
	indent string,
) (string, error) {
	summary := fmt.Sprintf("%s- ", indent)

	err := db.Update(func(tx *bolt.Tx) error {
		// Fetch current state of both Tamas
		bucket := tx.Bucket([]byte(tamaBucketName))
		if bucket == nil {
			return fmt.Errorf("Tama bucket doesn't exist!")
		}

		instigatorBytes := bucket.Get(models.BytesFromJacuzziId(instigatorId))
		if instigatorBytes == nil {
			return fmt.Errorf("Instigator Tama %d doesn't exist!", instigatorId)
		}

		targetBytes := bucket.Get(models.BytesFromJacuzziId(targetId))
		if targetBytes == nil {
			return fmt.Errorf("Target Tama %d doesn't exist!", targetId)
		}

		instigator := models.Tama{}
		target := models.Tama{}

		err := json.Unmarshal(instigatorBytes, &instigator)
		if err != nil {
			return err
		}

		err = json.Unmarshal(targetBytes, &target)
		if err != nil {
			return err
		}

		// Choose a random interaction
		interaction, bullyTraitCausedPickOn := chooseRandomInteraction(&instigator, &target)

		desiredInstigatorRelationshipChange := models.RelationshipScore(0)
		var instigatorResult models.RelationshipScoreModificationResult
		desiredTargetRelationshipChange := models.RelationshipScore(0)
		var targetResult models.RelationshipScoreModificationResult

		// Execute the interaction
		switch interaction {
		case models.Play:
			// A -> B += 1, B -> A += 1
			desiredInstigatorRelationshipChange = 1
			instigatorResult = instigator.ModifyRelationshipScoreWith(&target, desiredInstigatorRelationshipChange)
			desiredTargetRelationshipChange = 1
			targetResult = target.ModifyRelationshipScoreWith(&instigator, desiredTargetRelationshipChange)
			summary += fmt.Sprintf("**%s played with %s**, and they had fun!",
				instigator.GetNameAndId(), target.GetNameAndId())
		case models.Gift:
			summary += fmt.Sprintf("**%s gave %s a gift**, and %s ",
				instigator.GetNameAndId(), target.GetNameAndId(), target.GetNameAndId())
			// Choose the outcome of the gift giving
			giftOutcome := chooseRandomGiftOutcome()
			switch giftOutcome {
			case models.Likes:
				// B -> A += 2
				desiredTargetRelationshipChange = 2
				targetResult = target.ModifyRelationshipScoreWith(&instigator, desiredTargetRelationshipChange)
				summary += "liked it!"
			case models.Indifferent:
				// A -> B -= 1
				desiredInstigatorRelationshipChange = -1
				instigatorResult = instigator.ModifyRelationshipScoreWith(&target, desiredInstigatorRelationshipChange)
				summary += "didn't care for it!"
			case models.Hates:
				// A -> B -= 2, B -> A -= 1
				desiredInstigatorRelationshipChange = -2
				instigatorResult = instigator.ModifyRelationshipScoreWith(&target, desiredInstigatorRelationshipChange)
				desiredTargetRelationshipChange = -1
				targetResult = target.ModifyRelationshipScoreWith(&instigator, desiredTargetRelationshipChange)
				summary += "hated it!"
			default:
				panic(fmt.Sprintf("Unknown GiftOutcome %d!", giftOutcome))
			}
		case models.PickOn:
			// B -> A -= 2
			desiredTargetRelationshipChange = -2
			targetResult = target.ModifyRelationshipScoreWith(&instigator, desiredTargetRelationshipChange)
			summary += fmt.Sprintf("**%s picked on %s**, what a dick!",
				instigator.GetNameAndId(), target.GetNameAndId())
		default:
			panic(fmt.Sprintf("Unknown TamaInteraction %d!", interaction))
		}

		// Document any special reasons for the interaction
		if bullyTraitCausedPickOn {
			summary += fmt.Sprintf("\n%s  - This only happened because %s has the Bully trait (doubled option weight).",
				indent, instigator.GetNameAndId())
		}

		// Write any changes back to DB
		instigatorBytes, err = json.Marshal(instigator)
		if err != nil {
			return err
		}
		targetBytes, err = json.Marshal(target)
		if err != nil {
			return err
		}

		err = bucket.Put(models.BytesFromJacuzziId(instigatorId), instigatorBytes)
		if err == nil {
			summary += documentDirectionalInteractionResult(
				&instigator, &target, desiredInstigatorRelationshipChange, instigatorResult, indent)
		} else {
			log.Printf("Couldn't write instigator Tama back to DB: %s\n", err.Error())
			summary += fmt.Sprintf("\n%s  - **Error updating %s in the database**"+constants.ErrorReportMessageSuffix,
				indent, instigator.GetNameAndId())
		}

		err = bucket.Put(models.BytesFromJacuzziId(targetId), targetBytes)
		if err == nil {
			summary += documentDirectionalInteractionResult(
				&target, &instigator, desiredTargetRelationshipChange, targetResult, indent)
		} else {
			log.Printf("Couldn't write target Tama back to DB: %s\n", err.Error())
			summary += fmt.Sprintf("\n%s  - **Error updating %s in the database**"+constants.ErrorReportMessageSuffix,
				indent, target.GetNameAndId())
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return summary, nil
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

// Get a set of all Tamas, optionally filtered by:
// - userId: if not "", filter to those owned by the given user
// - aliveOnly: if true, only return alive Tamas
// - hatchedOnly: if true, only return hatched Tamas (not eggs)
func GetAllTamas(serverId string, userId string, aliveOnly bool, hatchedOnly bool) (map[models.JacuzziId]*models.Tama, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return getAllTamas(db, userId, aliveOnly, hatchedOnly), nil
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

// One Tama interacts with another, randomly choosing an interaction type. Returns a string describing the interaction.
func TamaInteract(
	serverId string,
	instigatorId models.JacuzziId,
	targetId models.JacuzziId,
	indent string,
) (string, error) {
	// Create or open a server-specific database file
	db, err := getDb(serverId)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return tamaInteract(db, instigatorId, targetId, indent)
}
