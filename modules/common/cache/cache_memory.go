package cache

import (
	"context"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// memoryCache is an in-memory Cache[K, V] implementation backed by sync.Map
// for concurrent-safe access without explicit locking. TTL is accepted for
// interface parity but not honored — entries persist until Delete, Clear, or
// process exit.
type memoryCache[K comparable, V any] struct {
	name string
	data sync.Map
	done chan struct{}
	once sync.Once
}

// NewMemoryCache constructs a Cache[K, V] backed by sync.Map under the given
// name. The returned cache reports name via Cache.Name and is safe for
// concurrent use; consumers may register the returned instance via Register
// and retrieve it via Lookup[K, V](name).
func NewMemoryCache[K comparable, V any](name string) Cache[K, V] {
	return &memoryCache[K, V]{
		name: name,
		data: sync.Map{},
		done: make(chan struct{}),
	}
}

// Name returns the cache name supplied to NewMemoryCache.
func (c *memoryCache[K, V]) Name() string {
	cassert.NotNil(c, "memory cache receiver is nil")

	return c.name
}

// Start is a no-op for the memory cache. It holds no external resources to
// initialize and returns nil immediately, satisfying the lifecycle.Component
// worker-style contract; Done is closed after Stop completes.
func (c *memoryCache[K, V]) Start(_ context.Context) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	return nil
}

// Stop closes the Done channel idempotently. The memory cache holds no
// external resources to release, so Stop returns nil; subsequent calls are
// no-ops.
func (c *memoryCache[K, V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "memory cache receiver is nil")

	defer c.once.Do(func() { close(c.done) })

	return nil
}

// Done returns the channel that is closed after Stop has been called.
func (c *memoryCache[K, V]) Done() <-chan struct{} {
	cassert.NotNil(c, "memory cache receiver is nil")

	return c.done
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
