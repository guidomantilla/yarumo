// Package uuid provides RFC 4122 UUID v4 and v7 generators and a UUID
// format validator. Consumers use the preconfigured UuidV4 and UuidV7
// singletons or the free functions directly; for registry-based lookup,
// register them explicitly via uids.Register(uuid.UuidV4) and/or
// uids.Register(uuid.UuidV7).
package uuid

import "github.com/guidomantilla/yarumo/uids"

var (
	_ uids.UIDFn   = UUIDv4
	_ uids.UIDFn   = UUIDv7
	_ uids.IsUIDFn = IsUUID
)

// Algorithm names registered by this provider.
const (
	NameUUIDv4 = "UUIDv4"
	NameUUIDv7 = "UUIDv7"
)
