package cache

import (
	"context"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// INMEMORY_CACHE is the preconfigured Cache[string, any] used as the package
// seed. It is registered under the names "default" and its own Name() value at
// package init, so Lookup[string, any]("default") and the facade resolve to it
// out of the box. TTL is not honored — for TTL/eviction, register a different
// cache via Register and use Lookup to operate on it.
var INMEMORY_CACHE = NewMemoryCache[string, any]("INMEMORY_CACHE")

// memoryCache is an in-memory Cache[K, V] implementation backed by sync.Map
// for concurrent-safe access without explicit locking. TTL is accepted for
// interface parity but not honored — entries persist until Delete, Clear, or
// process exit.
type memoryCache[K comparable, V any] struct {
	name string
	data sync.Map
}

// NewMemoryCache constructs a Cache[K, V] backed by sync.Map under the given
// name. The returned cache reports name via Cache.Name and is safe for
// concurrent use; consumers may register the returned instance via Register
// and retrieve it via Lookup[K, V](name).
func NewMemoryCache[K comparable, V any](name string) Cache[K, V] {
	return &memoryCache[K, V]{
		name: name,
		data: sync.Map{},
	}
}

// Name returns the cache name supplied to NewMemoryCache.
func (c *memoryCache[K, V]) Name() string {
	cassert.NotNil(c, "memory cache receiver is nil")

	return c.name
}

// Get returns the value stored at key. It returns an error wrapping
// ErrCacheMiss when the key is absent, or an error wrapping
// ErrCacheTypeAssertion when the stored value does not match V.
func (c *memoryCache[K, V]) Get(_ context.Context, key K) (V, error) {
	cassert.NotNil(c, "memory cache receiver is nil")

	anyValue, ok := c.data.Load(key)
	if !ok {
		return cpointer.Zero[V](), ErrMiss()
	}

	vValue, ok := anyValue.(V)
	if !ok {
		return cpointer.Zero[V](), ErrTypeAssertion()
	}

	return vValue, nil
}

// Set stores value under key. The ttl argument is accepted for interface
// parity but ignored — entries persist until Delete, Clear, or process exit.
func (c *memoryCache[K, V]) Set(_ context.Context, key K, value V, _ time.Duration) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	c.data.Store(key, value)
	return nil
}

// Delete removes the entry at key. It is a no-op when the key is absent.
func (c *memoryCache[K, V]) Delete(_ context.Context, key K) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	c.data.Delete(key)
	return nil
}

// Has reports whether key is present.
func (c *memoryCache[K, V]) Has(_ context.Context, key K) (bool, error) {
	cassert.NotNil(c, "memory cache receiver is nil")

	_, ok := c.data.Load(key)
	return ok, nil
}

// Clear removes every entry from the cache.
func (c *memoryCache[K, V]) Clear(_ context.Context) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	c.data.Clear()
	return nil
}

// Stop releases the cache's resources. The memory cache holds no external
// resources, so Stop is a no-op and safe to call more than once.
func (c *memoryCache[K, V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	return nil
}
