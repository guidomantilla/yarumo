package redis

import (
	"time"

	ccache "github.com/guidomantilla/yarumo/common/cache"
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied at cache construction time.
type Options struct {
	ttl       time.Duration
	keyPrefix string

	addr     string
	password string
	db       int

	codec ccache.Codec
}

// NewOptions creates Options with safe defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		ttl:       5 * time.Minute,
		keyPrefix: "",

		addr:     "",
		password: "",
		db:       0,

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
// values are ignored, preserving the default.
func WithKeyPrefix(prefix string) Option {
	return func(opts *Options) {
		if prefix != "" {
			opts.keyPrefix = prefix
		}
	}
}

// WithAddr sets the redis server address. Empty values are ignored;
// go-redis then defaults Addr to "localhost:6379" at client init.
func WithAddr(addr string) Option {
	return func(opts *Options) {
		if addr != "" {
			opts.addr = addr
		}
	}
}

// WithPassword sets the redis auth password. Empty values are ignored.
func WithPassword(password string) Option {
	return func(opts *Options) {
		if password != "" {
			opts.password = password
		}
	}
}

// WithDB sets the redis logical DB index. Negative values are ignored.
func WithDB(db int) Option {
	return func(opts *Options) {
		if db >= 0 {
			opts.db = db
		}
	}
}

// WithCodec overrides the codec used to serialize values for storage.
// Nil codecs are ignored, preserving the default common/cache.JSONCodec.
func WithCodec(codec ccache.Codec) Option {
	return func(opts *Options) {
		if codec != nil {
			opts.codec = codec
		}
	}
}
