package cache

import (
	"time"

	ccache "github.com/guidomantilla/yarumo/common/cache"
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied at cache construction time. Not
// every field applies to every backend (e.g. redisPassword is ignored by
// ristretto); each backend reads only what it needs.
type Options struct {
	ttl       time.Duration
	keyPrefix string
	lazyInit  bool

	ristrettoNumCtrs  int64
	ristrettoMaxCost  int64
	ristrettoBufItems int64

	redisAddr     string
	redisPassword string
	redisDB       int

	codec ccache.Codec
}

// NewOptions creates Options with safe defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		ttl:       5 * time.Minute,
		keyPrefix: "",
		lazyInit:  false,

		ristrettoNumCtrs:  1_000_000,
		ristrettoMaxCost:  100 << 20,
		ristrettoBufItems: 64,

		redisAddr:     "",
		redisPassword: "",
		redisDB:       0,

		codec: ccache.JSONCodec{},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
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

// WithKeyPrefix overrides the key prefix used to namespace cache keys.
// Effective prefix is "<name>:" when this option is not provided. Empty
// values are ignored, preserving the default. Applies to every backend.
func WithKeyPrefix(prefix string) Option {
	return func(opts *Options) {
		if prefix != "" {
			opts.keyPrefix = prefix
		}
	}
}

// WithLazyInit opts out of eager connection checks at construction time.
// Currently affects only the redis backend: when present, BuildRedisCache
// skips the post-construction PING. Ignored by the ristretto backend, which
// has no remote connection to verify.
func WithLazyInit() Option {
	return func(opts *Options) {
		opts.lazyInit = true
	}
}

// WithRistrettoCapacity overrides the ristretto counter count, max cost and
// buffer item size. Non-positive values are ignored per-parameter. Ignored
// by the redis backend.
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

// WithRedisAddr sets the redis server address. Empty values are ignored;
// go-redis then defaults Addr to "localhost:6379" at client init. Ignored
// by the ristretto backend.
func WithRedisAddr(addr string) Option {
	return func(opts *Options) {
		if addr != "" {
			opts.redisAddr = addr
		}
	}
}

// WithRedisPassword sets the redis auth password. Empty values are ignored.
// Ignored by the ristretto backend.
func WithRedisPassword(password string) Option {
	return func(opts *Options) {
		if password != "" {
			opts.redisPassword = password
		}
	}
}

// WithRedisDB sets the redis logical DB index. Negative values are ignored.
// Ignored by the ristretto backend.
func WithRedisDB(db int) Option {
	return func(opts *Options) {
		if db >= 0 {
			opts.redisDB = db
		}
	}
}

// WithCodec overrides the codec used by backends that need to serialize
// values (currently: redis). Nil codecs are ignored, preserving the default
// common/cache.JSONCodec. Ignored by the ristretto backend.
func WithCodec(codec ccache.Codec) Option {
	return func(opts *Options) {
		if codec != nil {
			opts.codec = codec
		}
	}
}
