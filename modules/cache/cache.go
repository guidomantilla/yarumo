package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache/v3"
	libstore "github.com/eko/gocache/lib/v4/store"
	cassert "github.com/guidomantilla/yarumo/common/assert"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// cache is the private implementation of the Cache[K, V] interface. It wraps a
// gocache instance, applies the wrapper's TTL semantics, and optionally emits
// OTel metrics and slog debug logs on each event.
type cache[K comparable, V any] struct {
	options *Options
	backend *backendInstance
	metrics *otelMetrics
	stopped atomic.Bool
}

// NewCache constructs a Cache[K, V] with the given options. The returned cache
// is ready for use; callers MUST eventually call Stop to release backend
// resources when they are done with it.
func NewCache[K comparable, V any](opts ...Option) (Cache[K, V], error) {
	options := NewOptions(opts...)

	backend, err := buildBackend(options)
	if err != nil {
		return nil, ErrCache(err)
	}

	var metrics *otelMetrics
	if options.otelEnabled {
		metrics = newOtelMetrics(options.otelMeterName)
	}

	return &cache[K, V]{
		options: options,
		backend: backend,
		metrics: metrics,
	}, nil
}

// Get returns the value stored at key. It returns an ErrCacheMiss-typed error
// when the key is absent or expired.
func (c *cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	cassert.NotNil(c, "cache receiver is nil")

	var zero V

	cacheKey, err := stringKey(key)
	if err != nil {
		return zero, ErrCache(err)
	}

	raw, err := c.backend.cache.Get(ctx, cacheKey)
	if err != nil {
		if isNotFound(err) {
			c.recordMiss(ctx, cacheKey)
			return zero, ErrMiss()
		}
		return zero, ErrCache(err)
	}

	value, ok := raw.(V)
	if !ok {
		c.recordMiss(ctx, cacheKey)
		return zero, ErrSerialize(errors.New("value type mismatch"))
	}

	c.recordHit(ctx, cacheKey)
	return value, nil
}

// Set stores value under key with the given TTL. A non-positive ttl resolves
// to the cache default configured via WithTTL.
func (c *cache[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	cassert.NotNil(c, "cache receiver is nil")

	cacheKey, err := stringKey(key)
	if err != nil {
		return ErrCache(err)
	}

	opts := setOptionsForTTL(ttl, c.options.ttl)

	err = c.backend.cache.Set(ctx, cacheKey, value, opts...)
	if err != nil {
		return ErrCache(err)
	}

	c.recordSet(ctx, cacheKey)
	return nil
}

// Delete removes the entry at key. It is a no-op if the key is absent.
func (c *cache[K, V]) Delete(ctx context.Context, key K) error {
	cassert.NotNil(c, "cache receiver is nil")

	cacheKey, err := stringKey(key)
	if err != nil {
		return ErrCache(err)
	}

	err = c.backend.cache.Delete(ctx, cacheKey)
	if err != nil {
		return ErrCache(err)
	}

	c.recordEviction(ctx, cacheKey)
	return nil
}

// Has reports whether key is present in the cache.
func (c *cache[K, V]) Has(ctx context.Context, key K) bool {
	cassert.NotNil(c, "cache receiver is nil")

	cacheKey, err := stringKey(key)
	if err != nil {
		return false
	}

	raw, err := c.backend.cache.Get(ctx, cacheKey)
	if err != nil {
		return false
	}

	return raw != nil
}

// Clear empties the cache.
func (c *cache[K, V]) Clear(ctx context.Context) error {
	cassert.NotNil(c, "cache receiver is nil")

	err := c.backend.cache.Clear(ctx)
	if err != nil {
		return ErrCache(err)
	}

	c.recordEviction(ctx, "*")
	return nil
}

// Stop releases the backend resources. Safe to call multiple times.
func (c *cache[K, V]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "cache receiver is nil")

	swapped := c.stopped.CompareAndSwap(false, true)
	if !swapped {
		return nil
	}

	err := c.backend.closer.Close()
	if err != nil {
		return ErrCache(err)
	}

	if c.options.slogEnabled {
		clog.Debug(ctx, "cache stopped", "component", "cache", "backend", string(c.options.backend))
	}

	return nil
}

// recordHit emits the OTel hit counter and optionally logs a debug message.
func (c *cache[K, V]) recordHit(ctx context.Context, key string) {
	if c.metrics != nil {
		c.metrics.recordHit(ctx)
	}
	if c.options.slogEnabled {
		clog.Debug(ctx, "cache hit", "component", "cache", "backend", string(c.options.backend), "key", key)
	}
}

// recordMiss emits the OTel miss counter and optionally logs a debug message.
func (c *cache[K, V]) recordMiss(ctx context.Context, key string) {
	if c.metrics != nil {
		c.metrics.recordMiss(ctx)
	}
	if c.options.slogEnabled {
		clog.Debug(ctx, "cache miss", "component", "cache", "backend", string(c.options.backend), "key", key)
	}
}

// recordSet emits the OTel set counter and optionally logs a debug message.
func (c *cache[K, V]) recordSet(ctx context.Context, key string) {
	if c.metrics != nil {
		c.metrics.recordSet(ctx)
	}
	if c.options.slogEnabled {
		clog.Debug(ctx, "cache set", "component", "cache", "backend", string(c.options.backend), "key", key)
	}
}

// recordEviction emits the OTel eviction counter and optionally logs a debug message.
func (c *cache[K, V]) recordEviction(ctx context.Context, key string) {
	if c.metrics != nil {
		c.metrics.recordEviction(ctx)
	}
	if c.options.slogEnabled {
		clog.Debug(ctx, "cache eviction", "component", "cache", "backend", string(c.options.backend), "key", key)
	}
}

// isNotFound reports whether err is a backend's "not found" sentinel.
//
// gocache and most store adapters wrap a libstore.NotFound, but a handful of
// backends (notably allegro/bigcache for direct lookups) surface their own
// native sentinel. This helper recognises both.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	var nf *libstore.NotFound
	if errors.As(err, &nf) {
		return true
	}

	// Some store backends return store.NotFound by value via Is(); cross-check.
	probe := libstore.NotFoundWithCause(nil)
	if errors.Is(err, probe) {
		return true
	}

	// allegro/bigcache returns its own native error before gocache wraps it.
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return true
	}

	return false
}
