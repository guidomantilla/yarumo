// Package ulid provides a ULID generator and format validator backed by
// github.com/oklog/ulid/v2. Consumers use the preconfigured Ulid singleton
// or the free functions directly; for registry-based lookup, register it
// explicitly via uids.Register(Ulid).
package ulid

import "github.com/guidomantilla/yarumo/uids"

var (
	_ uids.UIDFn   = ULID
	_ uids.IsUIDFn = IsULID
)

// Name is the algorithm name registered by this provider.
const Name = "ULID"
