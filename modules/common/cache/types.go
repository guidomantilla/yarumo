// Package cache provides a generic Cache[K, V] interface plus a package-level
// generic facade. Consumers register additional Cache implementations via
// Register and retrieve them via Lookup.
//
// The registry is heterogeneous: Register[K, V] accepts any Cache[K, V] and
// Lookup[K, V] retrieves it back by name with a type assertion. The
// package-level facade (Get, Set, Delete, Has, Clear, Stop) is generic too —
// each call delegates to the cache registered under "default" for the
// requested K and V; an ErrCacheTypeAssertion is returned when the registered
// "default" does not match the requested types.
//
// The package also exposes two shared primitives that backend implementations
// (e.g. modules/cache) reuse: Codec for value serialization, and
// ResolveKeyPrefix for cache-name namespace resolution.
package cache

import (
	"context"
	"time"
)

// Type compliance: the default memory cache satisfies the canonical
// Cache[string, any] shape, the package-level facade, registry, and shared
// primitive functions match their declared Fn aliases (instantiated with
// string/any for the generic ones), JSONCodec satisfies the Codec contract,
// and Error implements the standard error contract.
var (
	_ Cache[string, any] = (*memoryCache[string, any])(nil)
	_ error              = (*Error)(nil)
	_ Codec              = JSONCodec{}

	_ GetFn[string, any]    = Get[string, any]
	_ SetFn[string, any]    = Set[string, any]
	_ DeleteFn[string, any] = Delete[string, any]
	_ HasFn[string, any]    = Has[string, any]
	_ ClearFn[string, any]  = Clear[string, any]
	_ StopFn[string, any]   = Stop[string, any]

	_ RegisterFn[string, any] = Register[string, any]
	_ LookupFn[string, any]   = Lookup[string, any]
	_ SupportedFn             = Supported
	_ ResolveKeyPrefixFn      = ResolveKeyPrefix

	_ ErrCacheFn         = ErrCache
	_ ErrMissFn          = ErrMiss
	_ ErrTypeAssertionFn = ErrTypeAssertion
	_ ErrNotRegisteredFn = ErrNotRegistered
)

// Cache is the generic cache contract. Implementations must be safe for
// concurrent use by multiple goroutines. The caller is responsible for calling
// Stop when the cache is no longer needed; calls after Stop yield
// implementation-defined behavior.
type Cache[K comparable, V any] interface {
	// Name returns a stable identifier for this cache (used as its registry key).
	Name() string
	// Get returns the value stored at key. It returns an error wrapping
	// ErrCacheMiss when the key is absent. Implementations may also return
	// errors specific to their backend (e.g. ErrCacheTypeAssertion in
	// memoryCache when the stored value's dynamic type does not match V).
	Get(ctx context.Context, key K) (V, error)
	// Set stores value under key with the given TTL. Implementations may ignore
	// ttl when their backend does not support per-entry expiration.
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	// Delete removes the value at key. It is a no-op when the key is absent.
	Delete(ctx context.Context, key K) error
	// Has reports whether key is present. It returns an error if the backend
	// cannot determine presence.
	Has(ctx context.Context, key K) (bool, error)
	// Clear removes every entry.
	Clear(ctx context.Context) error
	// Stop releases the resources held by the cache. Implementations must make
	// Stop idempotent.
	Stop(ctx context.Context) error
}

// Codec encodes and decodes cache values for backends that store raw bytes
// (e.g. redis, memcached, distributed key/value stores). Backends that
// operate on Go values directly (e.g. an in-memory map cache) do not use a
// Codec. The default implementation provided by this package is JSONCodec;
// consumers needing other formats (msgpack, gob, protobuf) supply their own
// Codec.
type Codec interface {
	// Encode marshals v into bytes for backend storage.
	Encode(v any) ([]byte, error)
	// Decode unmarshals data into the target pointed to by v. v must be a
	// non-nil pointer to a value of the cache's V type.
	Decode(data []byte, v any) error
}

// GetFn is the function type for Get.
type GetFn[K comparable, V any] func(ctx context.Context, key K) (V, error)

// SetFn is the function type for Set.
type SetFn[K comparable, V any] func(ctx context.Context, key K, value V, ttl time.Duration) error

// DeleteFn is the function type for Delete.
type DeleteFn[K comparable, V any] func(ctx context.Context, key K) error

// HasFn is the function type for Has.
type HasFn[K comparable, V any] func(ctx context.Context, key K) (bool, error)

// ClearFn is the function type for Clear.
type ClearFn[K comparable, V any] func(ctx context.Context) error

// StopFn is the function type for Stop.
type StopFn[K comparable, V any] func(ctx context.Context) error

// RegisterFn is the function type for Register.
type RegisterFn[K comparable, V any] func(name string, cache Cache[K, V])

// LookupFn is the function type for Lookup.
type LookupFn[K comparable, V any] func(name string) (Cache[K, V], error)

// SupportedFn is the function type for Supported.
type SupportedFn func() []string

// ResolveKeyPrefixFn is the function type for ResolveKeyPrefix.
type ResolveKeyPrefixFn func(name, configured string) string

// ErrCacheFn is the function type for ErrCache.
type ErrCacheFn func(causes ...error) error

// ErrMissFn is the function type for ErrMiss.
type ErrMissFn func(causes ...error) error

// ErrTypeAssertionFn is the function type for ErrTypeAssertion.
type ErrTypeAssertionFn func(causes ...error) error

// ErrNotRegisteredFn is the function type for ErrNotRegistered.
type ErrNotRegisteredFn func(name string) error
