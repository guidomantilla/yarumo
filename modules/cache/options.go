package cache

import (
	"time"
)

// Option is a functional option for configuring cache Options.
type Option func(opts *Options)

// Options holds the configuration for cache construction.
type Options struct {
	backend           Backend
	ttl               time.Duration
	otelEnabled       bool
	otelMeterName     string
	slogEnabled       bool
	ristrettoNumCtrs  int64
	ristrettoMaxCost  int64
	ristrettoBufItems int64
	bigcacheShards    int
	bigcacheLifeWin   time.Duration
	bigcacheCleanWin  time.Duration
	bigcacheMaxSize   int
	bigcacheMaxEntry  int
	gocacheDefault    time.Duration
	gocacheCleanup    time.Duration
}

// NewOptions creates Options with safe defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		backend:       BackendRistretto,
		ttl:           5 * time.Minute,
		otelEnabled:   false,
		otelMeterName: "github.com/guidomantilla/yarumo/cache",
		slogEnabled:   false,

		ristrettoNumCtrs:  1_000_000,
		ristrettoMaxCost:  100 << 20, // 100 MiB
		ristrettoBufItems: 64,

		bigcacheShards:   1024,
		bigcacheLifeWin:  10 * time.Minute,
		bigcacheCleanWin: 5 * time.Minute,
		bigcacheMaxSize:  256,
		bigcacheMaxEntry: 4096,

		gocacheDefault: 5 * time.Minute,
		gocacheCleanup: 10 * time.Minute,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithBackend selects the cache backend. Unknown backends are silently ignored
// and the default (ristretto) is preserved.
func WithBackend(backend Backend) Option {
	return func(opts *Options) {
		switch backend {
		case BackendRistretto, BackendBigcache, BackendGoCache:
			opts.backend = backend
		}
	}
}

// WithTTL sets the default time-to-live applied to entries when Set is called
// with a non-positive ttl. Values less than or equal to zero are ignored.
func WithTTL(ttl time.Duration) Option {
	return func(opts *Options) {
		if ttl > 0 {
			opts.ttl = ttl
		}
	}
}

// WithOTel enables OpenTelemetry metrics emission for cache hits, misses, sets,
// and evictions. The counters are created from the global meter provider lazily
// on the first call to a tracked method.
func WithOTel() Option {
	return func(opts *Options) {
		opts.otelEnabled = true
	}
}

// WithOTelMeterName overrides the OpenTelemetry meter name used when WithOTel
// is enabled. Empty values are ignored.
func WithOTelMeterName(name string) Option {
	return func(opts *Options) {
		if name != "" {
			opts.otelMeterName = name
		}
	}
}

// WithSlog enables debug logging on cache hits, misses, sets, and evictions
// through the common/log facade.
func WithSlog() Option {
	return func(opts *Options) {
		opts.slogEnabled = true
	}
}

// WithRistrettoCapacity overrides the Ristretto counter count, max cost and
// buffer item size. Non-positive values are ignored per-parameter.
func WithRistrettoCapacity(numCounters int64, maxCost int64, bufferItems int64) Option {
	return func(opts *Options) {
		if numCounters > 0 {
			opts.ristrettoNumCtrs = numCounters
		}
		if maxCost > 0 {
			opts.ristrettoMaxCost = maxCost
		}
		if bufferItems > 0 {
			opts.ristrettoBufItems = bufferItems
		}
	}
}

// WithBigcacheCapacity overrides Bigcache shard count, life window, clean
// window, hard max size (in MB) and per-entry max size (in bytes). Non-positive
// values are ignored per-parameter.
func WithBigcacheCapacity(shards int, lifeWindow time.Duration, cleanWindow time.Duration, hardMaxCacheSizeMB int, maxEntrySize int) Option {
	return func(opts *Options) {
		if shards > 0 {
			opts.bigcacheShards = shards
		}
		if lifeWindow > 0 {
			opts.bigcacheLifeWin = lifeWindow
		}
		if cleanWindow > 0 {
			opts.bigcacheCleanWin = cleanWindow
		}
		if hardMaxCacheSizeMB > 0 {
			opts.bigcacheMaxSize = hardMaxCacheSizeMB
		}
		if maxEntrySize > 0 {
			opts.bigcacheMaxEntry = maxEntrySize
		}
	}
}

// WithGoCacheCapacity overrides the go-cache default expiration and janitor
// cleanup interval. Non-positive values are ignored per-parameter.
func WithGoCacheCapacity(defaultExpiration time.Duration, cleanupInterval time.Duration) Option {
	return func(opts *Options) {
		if defaultExpiration > 0 {
			opts.gocacheDefault = defaultExpiration
		}
		if cleanupInterval > 0 {
			opts.gocacheCleanup = cleanupInterval
		}
	}
}
