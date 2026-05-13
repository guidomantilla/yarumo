package uids

import (
	"github.com/akshayvadher/cuid2"
	nanoid "github.com/devmiek/nanoid-go"
	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/rs/xid"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// UUIDv4 generates a random RFC 4122 version 4 UUID.
func UUIDv4() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", cerrs.Wrap(ErrGenerationFailed, err)
	}

	return id.String(), nil
}

// NANOID generates a tiny, secure, URL-friendly unique string ID.
func NANOID() (string, error) {
	id, err := nanoid.New()
	if err != nil {
		return "", cerrs.Wrap(ErrGenerationFailed, err)
	}

	return id, nil
}

// CUID2 generates a collision-resistant unique identifier.
func CUID2() (string, error) {
	return cuid2.CreateId(), nil
}

// UUIDv7 generates a time-ordered RFC 4122 version 7 UUID.
func UUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", cerrs.Wrap(ErrGenerationFailed, err)
	}

	return id.String(), nil
}

// ULID generates a universally unique lexicographically sortable identifier.
func ULID() (string, error) {
	return ulid.Make().String(), nil
}

// XID generates a globally unique ID inspired by MongoDB ObjectID.
func XID() (string, error) {
	return xid.New().String(), nil
}
