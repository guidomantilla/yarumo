// Package cache provides concrete backend implementations of the
// common/cache.Cache[K, V] contract. The package ships a ristretto-backed
// in-memory implementation and a redis-backed distributed implementation.
// Each implementation has its own BuildXxxCache constructor with an
// identical signature ([V any], same Options); consumers may use the
// returned cache directly or register it at common/cache.Register for the
// heterogeneous registry.
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
// redis client) that must be released by calling Stop. The package does
// not integrate with any lifecycle manager — the consumer decides how Stop
// gets invoked (defer, registration in modules/managed, etc.).
package cache

import (
	ccache "github.com/guidomantilla/yarumo/common/cache"
)

// Type compliance: ristrettoCache and redisCache both implement the
// canonical Cache[string, any] shape from common/cache, and Error
// implements the standard error contract.
var (
	_ ccache.Cache[string, any] = (*ristrettoCache[any])(nil)
	_ ccache.Cache[string, any] = (*redisCache[any])(nil)
	_ error                     = (*Error)(nil)
)
