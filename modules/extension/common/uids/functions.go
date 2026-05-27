package uids

import (
	"regexp"

	"github.com/akshayvadher/cuid2"
	nanoid "github.com/devmiek/nanoid-go"
	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/rs/xid"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

// nanoIDRegex matches the default NanoID format: 21 characters from the
// URL-safe alphabet (A-Z, a-z, 0-9, _, -). The upstream
// github.com/devmiek/nanoid-go library does not expose a parser, so the
// canonical default alphabet and length are encoded here.
var nanoIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)

// cuid2Regex matches the canonical CUID2 format: 24 characters, lowercase
// alphanumeric, starting with a letter. Although
// github.com/akshayvadher/cuid2 exposes IsCuid, that helper accepts any
// length between 2 and 32 — too permissive for the default-length CUID2
// emitted by Generate. Length 24 is anchored here to match the canonical
// output of cuid2.CreateId.
var cuid2Regex = regexp.MustCompile(`^[a-z][a-z0-9]{23}$`)

// UUIDv4 generates a random RFC 4122 version 4 UUID.
func UUIDv4() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", cerrs.Wrap(cuids.ErrGenerationFailed, err)
	}

	return id.String(), nil
}

// NANOID generates a tiny, secure, URL-friendly unique string ID.
func NANOID() (string, error) {
	id, err := nanoid.New()
	if err != nil {
		return "", cerrs.Wrap(cuids.ErrGenerationFailed, err)
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
		return "", cerrs.Wrap(cuids.ErrGenerationFailed, err)
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

// IsUUID reports whether s is a syntactically valid RFC 4122 UUID of any
// version (v1 through v7). It accepts the canonical 36-character hyphenated
// form as well as the variants documented by github.com/google/uuid.Parse.
func IsUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// IsULID reports whether s is a syntactically valid ULID: 26 characters in
// Crockford Base32 (case-insensitive, excluding I, L, O, U). Uses
// ulid.ParseStrict so that invalid characters within the canonical length
// are rejected.
func IsULID(s string) bool {
	_, err := ulid.ParseStrict(s)
	return err == nil
}

// IsNanoID reports whether s matches the default NanoID format: 21
// characters from the URL-safe alphabet (A-Z, a-z, 0-9, underscore, hyphen).
// Custom alphabets or sizes are intentionally rejected.
func IsNanoID(s string) bool {
	return nanoIDRegex.MatchString(s)
}

// IsCUID2 reports whether s is a syntactically valid CUID2: exactly 24
// characters, lowercase alphanumeric, starting with a letter. Non-default
// lengths produced by CreateIdOf are intentionally rejected.
func IsCUID2(s string) bool {
	return cuid2Regex.MatchString(s)
}

// IsXID reports whether s is a syntactically valid XID: 20 characters in
// base32hex encoding as produced by github.com/rs/xid.
func IsXID(s string) bool {
	_, err := xid.FromString(s)
	return err == nil
}
