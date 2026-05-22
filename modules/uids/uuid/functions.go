package uuid

import (
	"github.com/google/uuid"

	"github.com/guidomantilla/yarumo/uids"
)

// Preconfigured UID generator singletons for the UUID v4 and v7
// algorithms. Consumers that want registry-based lookup must register
// them explicitly via uids.Register(UuidV4) and/or uids.Register(UuidV7)
// at startup.
var (
	UuidV4 = uids.NewUID(NameUUIDv4, UUIDv4)
	UuidV7 = uids.NewUID(NameUUIDv7, UUIDv7)
)

// UUIDv4 generates a random RFC 4122 version 4 UUID.
func UUIDv4() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", uids.ErrGeneration(err)
	}

	return id.String(), nil
}

// UUIDv7 generates a time-ordered RFC 4122 version 7 UUID.
func UUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", uids.ErrGeneration(err)
	}

	return id.String(), nil
}

// IsUUID reports whether s is a syntactically valid RFC 4122 UUID of any
// version (v1 through v7). It accepts the canonical 36-character hyphenated
// form as well as the variants documented by github.com/google/uuid.Parse.
func IsUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
