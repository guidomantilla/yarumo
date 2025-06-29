package uxid

import (
	"github.com/akshayvadher/cuid2"
	nanoid "github.com/devmiek/nanoid-go"
	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

const (
	UUIDv4 Type = iota
	NANOID
	CUID2
	UUIDv7
	ULID
	XId
)

type Type int

func NewUUIDv4() string {
	return BuildId(UUIDv4)
}

func NewNANOID() string {
	return BuildId(NANOID)
}

func NewCUID2() string {
	return BuildId(CUID2)
}

func NewUUIDv7() string {
	return BuildId(UUIDv7)
}

func NewULID() string {
	return BuildId(ULID)
}

func NewXId() string {
	return BuildId(XId)
}

// BuildId generates a new unique id based on the type provided.
// If the type is not provided, it will default to ULID.
// The following types are supported: UUIDv4, NANOID, CUID2, UUIDv7, ULID, XId.
// Random-based types are UUIDv4, NANOID, CUID2.
// Time-based types are UUIDv7, ULID, XId.
func BuildId(kind Type) string {

	switch kind {
	case UUIDv4:
		id := uuid.New()
		return id.String()
	case NANOID:
		id, _ := nanoid.New()
		return id
	case CUID2:
		id := cuid2.CreateId()
		return id

	case UUIDv7:
		id, _ := uuid.NewV7()
		return id.String()
	case ULID:
		id := ulid.Make()
		return id.String()
	case XId:
		id := xid.New()
		return id.String()
	default:
		id := ulid.Make()
		return id.String()
	}
}
