// Package uids ships the canonical UID generators (UUIDv4, UUIDv7, ULID,
// NanoID, CUID2, XID) and their format validators (IsUUID, IsULID, ...).
// The UID interface, function-type aliases, generic registry, and the
// trivial NewUID constructor live in modules/common/uids/. This package's
// init() registers the six preconfigured singletons against that
// registry, so importing it transparently enables Lookup by name.
package uids

import (
	cuids "github.com/guidomantilla/yarumo/common/uids"
)

var (
	_ cuids.UIDFn   = UUIDv4
	_ cuids.UIDFn   = NANOID
	_ cuids.UIDFn   = CUID2
	_ cuids.UIDFn   = UUIDv7
	_ cuids.UIDFn   = ULID
	_ cuids.UIDFn   = XID
	_ cuids.IsUIDFn = IsUUID
	_ cuids.IsUIDFn = IsULID
	_ cuids.IsUIDFn = IsNanoID
	_ cuids.IsUIDFn = IsCUID2
	_ cuids.IsUIDFn = IsXID
)
