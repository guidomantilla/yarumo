// Package cache provides concrete backend implementations of the
// common/cache.Cache[K, V] contract. The package ships a ristretto-backed
// in-memory implementation and a redis-backed distributed implementation.
//
// Constructors: each backend exposes a New<Backend>Cache constructor
// (infallible) returning a Cache[string, V]. The connection handshake
// (redis PING, ristretto client init) runs in Start per the
// common/lifecycle.Component contract; consumers wire the cache into the
// application lifecycle via lifecycle.Build, which dispatches Start and
// surfaces start-time failures through errChan as lifecycle.ErrStart.
//
// Keys: both backends use string keys. Logical keys are prefixed with
// "<name>:" by default so that multiple caches can coexist in the same
// storage (relevant for redis sharing a DB; ristretto preserves the same
// convention even though its storage is per-instance, for consistency).
// Override the prefix with WithKeyPrefix. Prefix resolution is delegated
// to common/cache.ResolveKeyPrefix; serialization for byte-storage
// backends is delegated to common/cache.Codec (default JSONCodec).
//
// Lifecycle: each cache owns external resources (a ristretto client, a
// redis client) that must be released by calling Stop. Stop is
// idempotent; Done closes after the first Stop completes.
package cache

import (
	ccache "github.com/guidomantilla/yarumo/common/cache"
	"github.com/guidomantilla/yarumo/common/lifecycle"
)

// Type compliance: ristrettoCache and redisCache both implement the
// canonical Cache[string, any] shape from common/cache and the
// lifecycle.Component contract; Error implements the standard error
// contract.
var (
	_ ccache.Cache[string, any] = (*ristrettoCache[any])(nil)
	_ ccache.Cache[string, any] = (*redisCache[any])(nil)
	_ lifecycle.Component       = (*ristrettoCache[any])(nil)
	_ lifecycle.Component       = (*redisCache[any])(nil)
	_ error                     = (*Error)(nil)
)
