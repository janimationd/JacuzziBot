package models

import (
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/constants"
)

// A JaccuziId is a server-unique unsigned integral ID. All things identified by such an ID have a unique value across
// all types of objects on the same server. IDs are granted sequentially starting from 1. A value of 0 means "no ID".
type JacuzziId = uint64

// Uninitialized, or error value.
const NoId JacuzziId = 0

// Convert a byte-serialized JacuzziId to its integral type value. We use little-endian encoding to match the linux
// system the bot will be running on.
func JacuzziIdFromBytes(bytes []byte) JacuzziId {
	return JacuzziId(binary.LittleEndian.Uint64(bytes))
}

// Convert a JacuzziId to a byte-serialized format. We use little-endian encoding to match the linux system the bot
// will be running on.
func BytesFromJacuzziId(id JacuzziId) []byte {
	return binary.LittleEndian.AppendUint64(nil, id)
}

// Convert a JacuzziId to a string.
func StringFromJacuzziId(id JacuzziId) string {
	return fmt.Sprint(id)
}

// Extract the numeric JacuzziId from the first parameter in the CustomId string.
func ExtractJacuzziIdFromCustomId(interaction *discordgo.InteractionCreate) (JacuzziId, error) {
	customId := interaction.MessageComponentData().CustomID
	parts := strings.Split(customId, "|")
	if len(parts) != 2 {
		return NoId, fmt.Errorf("Couldn't parse CustomID %s%s", customId, constants.ErrorReportMessageSuffix)
	} else {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return NoId, fmt.Errorf("Couldn't parse CustomID second part %s%s",
				parts[1], constants.ErrorReportMessageSuffix)
		} else {
			return JacuzziId(id), nil
		}
	}
}

// Unit tests to verify the above functions.
func JacuzziIdTests() {
	var id JacuzziId = 3
	bytes := BytesFromJacuzziId(id)
	afterId := JacuzziIdFromBytes(bytes)
	if id == afterId {
		log.Printf("JacuzziIdTests passed.\n")
	} else {
		log.Printf("JacuzziIdTests failed: %d != %d\n", id, afterId)
	}
}
