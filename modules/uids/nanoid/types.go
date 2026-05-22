// Package nanoid provides a NanoID generator and format validator backed
// by github.com/devmiek/nanoid-go. Consumers use the preconfigured NanoID
// singleton or the free functions directly; for registry-based lookup,
// register it explicitly via uids.Register(NanoID).
package nanoid

import "github.com/guidomantilla/yarumo/uids"

var (
	_ uids.UIDFn   = NANOID
	_ uids.IsUIDFn = IsNanoID
)

// Name is the algorithm name registered by this provider.
const Name = "NanoID"
