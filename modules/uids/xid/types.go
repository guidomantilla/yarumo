// Package xid provides an XID generator and format validator backed by
// github.com/rs/xid. Consumers use the preconfigured XId singleton or the
// free functions directly; for registry-based lookup, register it
// explicitly via uids.Register(XId).
package xid

import "github.com/guidomantilla/yarumo/uids"

var (
	_ uids.UIDFn   = XID
	_ uids.IsUIDFn = IsXID
)

// Name is the algorithm name registered by this provider.
const Name = "XID"
