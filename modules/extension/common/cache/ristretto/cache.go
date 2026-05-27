package ristretto

import (
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	ccache "github.com/guidomantilla/yarumo/core/common/cache"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	cpointer "github.com/guidomantilla/yarumo/core/common/pointer"
)

// ristrettoCache is a ristretto-backed Cache[string, V] implementation.
// Keys are stored under the resolved key prefix (default "<name>:"); the
// prefix is decorative in ristretto since storage is per-instance, but
// applied uniformly with other cache backends for consistency. The
// underlying ristretto client is constructed lazily in Start (which can
// fail with invalid config) and released in Stop.
type ristrettoCache[V any] struct {
	name      string
	keyPrefix string
	options   *Options
	client    *ristretto.Cache[string, V]

	done chan struct{}
	once sync.Once
}

// NewRistrettoCache constructs a ristretto-backed Cache[string, V] under
// the given name. The constructor performs no I/O and cannot fail: the
// ristretto client is instantiated in Start, where invalid configuration
// surfaces as a lifecycle.ErrStart wrapping ErrRistrettoInitFailed.
func NewRistrettoCache[V any](name string, opts ...Option) ccache.Cache[string, V] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &ristrettoCache[V]{
		name:      name,
		keyPrefix: ccache.ResolveKeyPrefix(name, options.keyPrefix),
		options:   options,
		done:      make(chan struct{}),
	}
}

// Name returns the cache name supplied to NewRistrettoCache.
func (c *ristrettoCache[V]) Name() string {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	return c.name
}

// Start instantiates the underlying ristretto client with the configured
// counter count, max cost and buffer items. It satisfies the
// lifecycle.Component worker-style contract: returns immediately on
// success, or returns a lifecycle.ErrStart wrapping ErrRistrettoInitFailed
// when ristretto rejects the configuration.
func (c *ristrettoCache[V]) Start(_ context.Context) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	client, err := ristretto.NewCache(&ristretto.Config[string, V]{
		NumCounters: c.options.numCtrs,
		MaxCost:     c.options.maxCost,
		BufferItems: c.options.bufItems,
	})
	if err != nil {
		return lifecycle.ErrStart(ErrInit(err))
	}

	c.client = client

	return nil
}

// Stop releases the underlying ristretto client and closes Done. It is
// idempotent: only the first call closes the client; subsequent calls
// are no-ops returning nil.
func (c *ristrettoCache[V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	c.once.Do(func() {
		if c.client != nil {
			c.client.Close()
		}
		close(c.done)
	})

	return nil
}

// Done returns the channel that is closed after Stop has been called.
func (c *ristrettoCache[V]) Done() <-chan struct{} {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	return c.done
}

// Get returns the value stored at key or an error wrapping ErrCacheMiss
// when the key is absent or has been evicted by the ristretto admission
// policy.
func (c *ristrettoCache[V]) Get(_ context.Context, key string) (V, error) {
	cassert.NotNil(c, "ristretto cache receiver is nil")

	value, ok := c.client.Get(c.keyPrefix + key)
	if !ok {
		return cpointer.Zero[V](), ccache.ErrMiss()
	}

	return value, nil
}

// Set stores value under key with the given TTL. A non-positive ttl
// resolves to the cache default configured via WithTTL. Returns an error
// wrapping ErrRistrettoSetRejected when ristretto rejects the write
// (admission policy).
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
