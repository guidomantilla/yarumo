package cache

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto/v2"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ccache "github.com/guidomantilla/yarumo/common/cache"
	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// ristrettoCache is a ristretto-backed Cache[string, V] implementation.
// Keys are stored under the resolved key prefix (default "<name>:"); the
// prefix is decorative in ristretto since storage is per-instance, but
// applied uniformly with the redis backend for consistency.
type ristrettoCache[V any] struct {
	name      string
	keyPrefix string
	options   *Options
	client    *ristretto.Cache[string, V]
	stopped   atomic.Bool
}

// BuildRistrettoCache builds a ristretto-backed Cache[string, V] under the
// given name. Options control the per-cache TTL default, the key prefix,
// and the underlying ristretto admission/eviction tuning. Returns an error
// wrapping ErrRistrettoInitFailed when ristretto.NewCache rejects the
// configuration. WithLazyInit, WithRedis*, and WithCodec are accepted but
// ignored by this backend.
func BuildRistrettoCache[V any](name string, opts ...Option) (ccache.Cache[string, V], error) {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	client, err := ristretto.NewCache(&ristretto.Config[string, V]{
		NumCounters: options.ristrettoNumCtrs,
		MaxCost:     options.ristrettoMaxCost,
		BufferItems: options.ristrettoBufItems,
	})
	if err != nil {
		return nil, ErrInit(err)
	}

	return &ristrettoCache[V]{
		name:      name,
		keyPrefix: ccache.ResolveKeyPrefix(name, options.keyPrefix),
		options:   options,
		client:    client,
	}, nil
}

// Name returns the cache name supplied to BuildRistrettoCache.
func (c *ristrettoCache[V]) Name() string {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	return c.name
}

// Get returns the value stored at key or an error wrapping ErrCacheMiss when
// the key is absent or has been evicted by the ristretto admission policy.
func (c *ristrettoCache[V]) Get(_ context.Context, key string) (V, error) {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	value, ok := c.client.Get(c.keyPrefix + key)
	if !ok {
		return cpointer.Zero[V](), ccache.ErrMiss()
	}

	return value, nil
}

// Set stores value under key with the given TTL. A non-positive ttl resolves
// to the cache default configured via WithTTL. Returns an error wrapping
// ErrRistrettoSetRejected when ristretto rejects the write (admission policy).
func (c *ristrettoCache[V]) Set(_ context.Context, key string, value V, ttl time.Duration) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	effective := ttl
	if effective <= 0 {
		effective = c.options.ttl
	}

	ok := c.client.SetWithTTL(c.keyPrefix+key, value, 1, effective)
	if !ok {
		return ErrSet()
	}

	c.client.Wait()
	return nil
}

// Delete removes the entry at key. It is a no-op when the key is absent.
func (c *ristrettoCache[V]) Delete(_ context.Context, key string) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	c.client.Del(c.keyPrefix + key)
	return nil
}

// Has reports whether key is present in the cache.
func (c *ristrettoCache[V]) Has(_ context.Context, key string) (bool, error) {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	_, ok := c.client.Get(c.keyPrefix + key)
	return ok, nil
}

// Clear removes every entry from the underlying ristretto cache. Because
// ristretto storage is per-instance, only this cache's entries are
// affected; the key prefix is irrelevant here.
func (c *ristrettoCache[V]) Clear(_ context.Context) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	c.client.Clear()
	return nil
}

// Stop releases the underlying ristretto client. Safe to call more than once;
// subsequent calls are no-ops.
func (c *ristrettoCache[V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	swapped := c.stopped.CompareAndSwap(false, true)
	if !swapped {
		return nil
	}

	c.client.Close()
	return nil
}
