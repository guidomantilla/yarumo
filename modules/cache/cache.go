package cache

import (
	"context"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// cache is the private implementation of the Cache[K, V] interface.
//
// Behaviour is pluggable per-backend via function fields populated at
// construction time by the matching factory in backends.go. The struct
// itself contains no backend-specific logic — each method just delegates
// to its function field after asserting the receiver is non-nil. This
// follows criterion 4 Exception 3 (Pluggable struct pattern) from the
// common coding standards and mirrors crypto's *Method.
//
// Hooks (slog and OTel metrics) are wired into the closures by the
// factory; the recordHit / recordMiss / recordSet / recordEviction
// helpers live on the cache and the closures call them directly.
type cache[K comparable, V any] struct {
	options *Options
	metrics *otelMetrics
	stopped atomic.Bool

	// Pluggable function fields — the cache's behaviour IS these closures.
	getFn    func(ctx context.Context, key K) (V, error)
	setFn    func(ctx context.Context, key K, value V, ttl time.Duration) error
	deleteFn func(ctx context.Context, key K) error
	hasFn    func(ctx context.Context, key K) bool
	clearFn  func(ctx context.Context) error
	stopFn   func(ctx context.Context) error
}

// NewCache constructs a Cache[K, V] with the given options. The returned cache
// is ready for use; callers MUST eventually call Stop to release backend
// resources when they are done with it.
func NewCache[K comparable, V any](opts ...Option) (Cache[K, V], error) {
	options := NewOptions(opts...)

	err := validateOptions(options)
	if err != nil {
		return nil, ErrCache(err)
	}

	switch options.backend {
	case BackendRistretto:
		return newRistrettoCache[K, V](options)
	case BackendBigcache:
		return newBigcacheCache[K, V](options)
	case BackendGoCache:
		return newGoCacheCache[K, V](options)
	default:
		return nil, ErrUnsupported()
	}
}

// Get returns the value stored at key. It returns an ErrCacheMiss-typed error
// when the key is absent or expired.
func (c *cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	cassert.NotNil(c, "cache receiver is nil")

	return c.getFn(ctx, key)
}

// Set stores value under key with the given TTL. A non-positive ttl resolves
// to the cache default configured via WithTTL.
func (c *cache[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	cassert.NotNil(c, "cache receiver is nil")

	return c.setFn(ctx, key, value, ttl)
}

// Delete removes the entry at key. It is a no-op if the key is absent.
func (c *cache[K, V]) Delete(ctx context.Context, key K) error {
	cassert.NotNil(c, "cache receiver is nil")

	return c.deleteFn(ctx, key)
}

// Has reports whether key is present in the cache.
func (c *cache[K, V]) Has(ctx context.Context, key K) bool {
	cassert.NotNil(c, "cache receiver is nil")

	return c.hasFn(ctx, key)
}

// Clear empties the cache.
func (c *cache[K, V]) Clear(ctx context.Context) error {
	cassert.NotNil(c, "cache receiver is nil")

	return c.clearFn(ctx)
}

// Stop releases the backend resources. Safe to call multiple times.
func (c *cache[K, V]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "cache receiver is nil")

	swapped := c.stopped.CompareAndSwap(false, true)
	if !swapped {
		return nil
	}

	return c.stopFn(ctx)
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

// recordStopped optionally logs a debug message when the backend has been released.
func (c *cache[K, V]) recordStopped(ctx context.Context) {
	if c.options.slogEnabled {
		clog.Debug(ctx, "cache stopped", "component", "cache", "backend", string(c.options.backend))
	}
}
