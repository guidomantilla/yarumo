// Package ristretto provides a ristretto-backed implementation of the
// common/cache.Cache[K, V] contract.
//
// The constructor (NewRistrettoCache) is infallible and performs no I/O.
// The ristretto client is instantiated in Start per the
// common/lifecycle.Component contract; consumers wire the cache into the
// application lifecycle via lifecycle.Build, which dispatches Start and
// surfaces start-time failures through errChan as lifecycle.ErrStart.
//
// Keys: string keys, prefixed with "<name>:" by default. The prefix is
// decorative in ristretto (storage is per-instance) but applied uniformly
// across cache backends for consistency. Override with WithKeyPrefix.
// Prefix resolution is delegated to common/cache.ResolveKeyPrefix.
//
// Lifecycle: each cache owns a ristretto client that must be released by
// calling Stop. Stop is idempotent; Done closes after the first Stop
// completes.
package ristretto

import (
	ccache "github.com/guidomantilla/yarumo/core/common/cache"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// Type compliance: ristrettoCache implements the canonical
// Cache[string, any] shape from common/cache and the lifecycle.Component
// contract; Error implements the standard error contract.
var (
	_ ccache.Cache[string, any] = (*ristrettoCache[any])(nil)
	_ lifecycle.Component       = (*ristrettoCache[any])(nil)
	_ error                     = (*Error)(nil)
)
