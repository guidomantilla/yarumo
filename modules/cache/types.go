// Package cache provides a yarumo-flavored wrapper around eko/gocache/v4 with
// in-memory backends (ristretto, bigcache, go-cache), lifecycle integration
// with modules/managed, OpenTelemetry metrics emission, and slog-based debug logging.
//
// The Cache[K, V] interface keeps a uniform API across backends. Keys must be
// comparable; non-string keys are converted to their fmt.Sprintf("%v", k)
// representation before being handed to the underlying store. Values are stored
// as Go values directly (no serialization is performed).
//
// Backend selection is done via WithBackend; ristretto is the default. Redis
// support is intentionally not included in this release and will land once the
// modules/datasource/goredis module exists.
//
// Concurrency: all caches returned by NewCache are safe for concurrent use.
// Lifecycle: a cache obtained via BuildCache participates in modules/managed
// startup/shutdown and releases its backend resources on Stop.
package cache

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/managed"
)

// Backend identifies a supported in-memory cache backend.
type Backend string

// Supported backends. Redis is intentionally excluded until modules/datasource/goredis lands.
const (
	// BackendRistretto selects the ristretto sampled-LFU in-memory backend (default).
	BackendRistretto Backend = "ristretto"
	// BackendBigcache selects the bigcache GC-friendly in-memory backend.
	BackendBigcache Backend = "bigcache"
	// BackendGoCache selects the go-cache simple TTL-based in-memory backend.
	BackendGoCache Backend = "go-cache"
)

// Metric names emitted when WithOTel is enabled.
const (
	// MetricHits is the OTel counter name for cache hits.
	MetricHits = "cache.hits"
	// MetricMisses is the OTel counter name for cache misses.
	MetricMisses = "cache.misses"
	// MetricSets is the OTel counter name for cache writes.
	MetricSets = "cache.sets"
	// MetricEvictions is the OTel counter name for cache evictions (deletes plus clears).
	MetricEvictions = "cache.evictions"
)

// Cache defines the generic cache interface exposed by this module.
//
// K must be comparable. V is unconstrained. Implementations are safe for
// concurrent use by multiple goroutines. The caller is responsible for calling
// Stop (when the cache was created via BuildCache) or letting it be released
// at program exit when created via NewCache.
type Cache[K comparable, V any] interface {
	// Get returns the value stored at key. It returns ErrCacheMiss when the key
	// is not present or has expired.
	Get(ctx context.Context, key K) (V, error)
	// Set stores value under key with the given TTL. A non-positive ttl means
	// "use the cache default TTL".
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	// Delete removes the value at key. It is a no-op if the key is absent.
	Delete(ctx context.Context, key K) error
	// Has reports whether key is present in the cache. It must not return any error.
	Has(ctx context.Context, key K) bool
	// Clear removes every entry from the cache.
	Clear(ctx context.Context) error
	// Stop releases the resources held by the cache. It is safe to call Stop
	// more than once; subsequent calls are no-ops.
	Stop(ctx context.Context) error
}

// CacheFn is the function type for NewCache. It exists for type-compliance per criterion 2 of CODING_STANDARDS.
type CacheFn[K comparable, V any] func(opts ...Option) (Cache[K, V], error)

// BuildCacheFn is the function type for BuildCache. It mirrors the managed builder shape for cache components.
type BuildCacheFn[K comparable, V any] func(ctx context.Context, name string, opts ...Option) (Cache[K, V], managed.StopFn, error)

// Type-compliance instantiations for the canonical string-keyed []byte cache.
var (
	_ Cache[string, []byte] = (*cache[string, []byte])(nil)

	_ CacheFn[string, []byte]      = NewCache[string, []byte]
	_ BuildCacheFn[string, []byte] = BuildCache[string, []byte]
)
