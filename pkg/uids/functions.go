package uids

import (
	"github.com/akshayvadher/cuid2"
	nanoid "github.com/devmiek/nanoid-go"
	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

func UUIDv4() string {
	id := uuid.New()
	return id.String()
}

func NANOID() string {
	id, _ := nanoid.New()
	return id
}

func CUID2() string {
	id := cuid2.CreateId()
	return id
}

func UUIDv7() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func ULID() string {
	id := ulid.Make()
	return id.String()
}

func XID() string {
	id := xid.New()
	return id.String()
}
