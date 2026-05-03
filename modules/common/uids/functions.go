package uids

import (
	"github.com/akshayvadher/cuid2"
	nanoid "github.com/devmiek/nanoid-go"
	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

// UUIDv4 generates a random RFC 4122 version 4 UUID.
func UUIDv4() string {
	return uuid.New().String()
}

// NANOID generates a tiny, secure, URL-friendly unique string ID.
func NANOID() string {
	id, _ := nanoid.New()
	return id
}

// CUID2 generates a collision-resistant unique identifier.
func CUID2() string {
	return cuid2.CreateId()
}

// UUIDv7 generates a time-ordered RFC 4122 version 7 UUID.
func UUIDv7() string {
	id, _ := uuid.NewV7()
	return id.String()
}

// ULID generates a universally unique lexicographically sortable identifier.
func ULID() string {
	return ulid.Make().String()
}

// XID generates a globally unique ID inspired by MongoDB ObjectID.
func XID() string {
	return xid.New().String()
}
