package ulid

import (
	ulid "github.com/oklog/ulid/v2"

	"github.com/guidomantilla/yarumo/uids"
)

// Ulid is the preconfigured ULID generator singleton. Consumers that want
// registry-based lookup must register it explicitly via
// uids.Register(Ulid) at startup.
var Ulid = uids.NewUID(Name, ULID)

// ULID generates a universally unique lexicographically sortable identifier.
func ULID() (string, error) {
	return ulid.Make().String(), nil
}

// IsULID reports whether s is a syntactically valid ULID: 26 characters in
// Crockford Base32 (case-insensitive, excluding I, L, O, U). Uses
// ulid.ParseStrict so that invalid characters within the canonical length
// are rejected.
func IsULID(s string) bool {
	_, err := ulid.ParseStrict(s)
	return err == nil
}
