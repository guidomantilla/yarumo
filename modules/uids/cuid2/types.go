// Package cuid2 provides a CUID2 generator and format validator backed by
// github.com/akshayvadher/cuid2. Consumers use the preconfigured Cuid2
// singleton or the free functions directly; for registry-based lookup,
// register it explicitly via uids.Register(Cuid2).
package cuid2

import "github.com/guidomantilla/yarumo/uids"

var (
	_ uids.UIDFn   = CUID2
	_ uids.IsUIDFn = IsCUID2
)

// Name is the algorithm name registered by this provider.
const Name = "CUID2"
