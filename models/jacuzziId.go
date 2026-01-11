package models

import "encoding/binary"

// A JaccuziId is a server-unique unsigned integral ID. All things identified by such an ID have a unique value across
// all types of objects on the same server. IDs are granted sequentially starting from 1. A value of 0 means "no ID".
type JacuzziId = uint64

// Uninitialized, or error value.
const NoId JacuzziId = 0

// Convert a byte-serialized JacuzziId to its integral type value. We use little-endian encoding to match the linux
// system the bot will be running on.
func JacuzziIdFromBytes(bytes []byte) JacuzziId {
	return JacuzziId(binary.BigEndian.Uint64(bytes))
}

// Convert a JacuzziId to a byte-serialized format. We use little-endian encoding to match the linux system the bot
// will be running on.
func BytesFromJacuzziId(id JacuzziId) []byte {
	buf := make([]byte, 8)
	return binary.BigEndian.AppendUint64(buf, id)
}
