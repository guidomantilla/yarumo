// Package cache provides the generic Cache[K, V] contract and a
// reference in-memory backend.
//
// Cache[K, V] embeds common/lifecycle.Component so every cache obeys the
// same Name / Start / Stop / Done contract used by the other long-lived
// components in the workspace (http.Server, grpc.Server, cron.Scheduler,
// diagnostics.*). Backends that hold external connections (e.g. redis,
// memcached) perform their handshake in Start; backends that hold only
// in-process state (memory, ristretto) implement Start as a no-op.
//
// Consumers construct cache instances explicitly via the backend's
// constructor (e.g. NewMemoryCache, modules/cache.BuildRedisCache) and
// wire them into the application's lifecycle via lifecycle.Build. There
// is no package-level registry and no default cache: every cache is
// passed by reference through application code.
//
// The package also exposes two shared primitives that backend
// implementations reuse: Codec for value serialization, and
// ResolveKeyPrefix for cache-name namespace resolution.
package cache

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// Type compliance: memoryCache satisfies the canonical Cache[string, any]
// shape and the lifecycle.Component contract; JSONCodec satisfies the
// Codec contract; Error implements the standard error contract; the
// public free function ResolveKeyPrefix matches its declared Fn alias;
// and the free factories in errors.go match their declared Fn aliases.
var (
	_ Cache[string, any]  = (*memoryCache[string, any])(nil)
	_ lifecycle.Component = (*memoryCache[string, any])(nil)
	_ error               = (*Error)(nil)
	_ Codec               = JSONCodec{}

	_ ResolveKeyPrefixFn = ResolveKeyPrefix

	_ ErrCacheFn         = ErrCache
	_ ErrMissFn          = ErrMiss
	_ ErrTypeAssertionFn = ErrTypeAssertion
)

// Cache is the generic cache contract. Implementations must be safe for
// concurrent use by multiple goroutines. The Name / Start / Stop / Done
// lifecycle contract is inherited from common/lifecycle.Component; see
// that package for the full set of invariants (idempotent Stop, single
// Done close, no re-Start). Calls to the data methods after Stop yield
// implementation-defined behavior.
type Cache[K comparable, V any] interface {
	lifecycle.Component
	// Get returns the value stored at key. It returns an error wrapping
	// ErrCacheMiss when the key is absent. Implementations may also return
	// errors specific to their backend (e.g. ErrCacheTypeAssertion in
	// memoryCache when the stored value's dynamic type does not match V).
	Get(ctx context.Context, key K) (V, error)
	// Set stores value under key with the given TTL. Implementations may
	// ignore ttl when their backend does not support per-entry expiration.
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	// Delete removes the value at key. It is a no-op when the key is absent.
	Delete(ctx context.Context, key K) error
	// Has reports whether key is present. It returns an error if the
	// backend cannot determine presence.
	Has(ctx context.Context, key K) (bool, error)
	// Clear removes every entry.
	Clear(ctx context.Context) error
}

// Codec encodes and decodes cache values for backends that store raw
// bytes (e.g. redis, memcached, distributed key/value stores). Backends
// that operate on Go values directly (e.g. an in-memory map cache) do
// not use a Codec. The default implementation provided by this package
// is JSONCodec; consumers needing other formats (msgpack, gob, protobuf)
// supply their own Codec.
type Codec interface {
	// Encode marshals v into bytes for backend storage.
	Encode(v any) ([]byte, error)
	// Decode unmarshals data into the target pointed to by v. v must be a
	// non-nil pointer to a value of the cache's V type.
	Decode(data []byte, v any) error
}

// ResolveKeyPrefixFn is the function type for ResolveKeyPrefix.
type ResolveKeyPrefixFn func(name, configured string) string

// ErrCacheFn is the function type for ErrCache.
type ErrCacheFn func(causes ...error) error

// ErrMissFn is the function type for ErrMiss.
type ErrMissFn func(causes ...error) error

// ErrTypeAssertionFn is the function type for ErrTypeAssertion.
type ErrTypeAssertionFn func(causes ...error) error
