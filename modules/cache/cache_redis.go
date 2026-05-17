package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ccache "github.com/guidomantilla/yarumo/common/cache"
	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// redisCache is a redis-backed Cache[string, V] implementation. Keys are
// stored under the resolved key prefix (default "<name>:") so that multiple
// caches can share a redis DB without colliding. Values are serialized via
// the configured Codec (default JSONCodec). Safe for concurrent use; Stop
// is idempotent and closes the underlying go-redis client.
type redisCache[V any] struct {
	name      string
	keyPrefix string
	options   *Options
	client    *redis.Client
	codec     ccache.Codec
	stopped   atomic.Bool
}

// BuildRedisCache builds a redis-backed Cache[string, V] under the given
// name. The redis address comes from WithRedisAddr (defaulting to go-redis'
// "localhost:6379" when absent). Unless WithLazyInit is provided, Build
// issues a PING after constructing the go-redis client and fails fast with
// an error wrapping ErrRedisCommandFailed if the server does not respond.
func BuildRedisCache[V any](name string, opts ...Option) (ccache.Cache[string, V], error) {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	client := redis.NewClient(&redis.Options{
		Addr:     options.redisAddr,
		Password: options.redisPassword,
		DB:       options.redisDB,
	})

	if !options.lazyInit {
		pingErr := client.Ping(context.Background()).Err()
		if pingErr != nil {
			_ = client.Close()
			return nil, ErrCommand(pingErr)
		}
	}

	return &redisCache[V]{
		name:      name,
		keyPrefix: ccache.ResolveKeyPrefix(name, options.keyPrefix),
		options:   options,
		client:    client,
		codec:     options.codec,
	}, nil
}

// Name returns the cache name supplied to BuildRedisCache.
func (c *redisCache[V]) Name() string {
	cassert.NotNil(c, "redis cache receiver is nil")

	return c.name
}

// Get returns the value stored at key or an error wrapping ErrCacheMiss when
// the key is absent. Returns an error wrapping ErrRedisCommandFailed for
// transport/protocol errors, or ErrRedisDecodeFailed when the stored bytes
// cannot be decoded into V via the configured codec.
func (c *redisCache[V]) Get(ctx context.Context, key string) (V, error) {
	cassert.NotNil(c, "redis cache receiver is nil")

	raw, err := c.client.Get(ctx, c.keyPrefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return cpointer.Zero[V](), ccache.ErrMiss()
	}
	if err != nil {
		return cpointer.Zero[V](), ErrCommand(err)
	}

	var value V
	decErr := c.codec.Decode(raw, &value)
	if decErr != nil {
		return cpointer.Zero[V](), ErrDecode(decErr)
	}

	return value, nil
}

// Set stores value under key with the given TTL. A non-positive ttl resolves
// to the cache default configured via WithTTL. Returns an error wrapping
// ErrRedisEncodeFailed when the codec cannot encode value, or
// ErrRedisCommandFailed when redis rejects the command.
func (c *redisCache[V]) Set(ctx context.Context, key string, value V, ttl time.Duration) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	raw, err := c.codec.Encode(value)
	if err != nil {
		return ErrEncode(err)
	}

	effective := ttl
	if effective <= 0 {
		effective = c.options.ttl
	}

	setErr := c.client.Set(ctx, c.keyPrefix+key, raw, effective).Err()
	if setErr != nil {
		return ErrCommand(setErr)
	}

	return nil
}

// Delete removes the entry at key. It is a no-op when the key is absent.
// Returns an error wrapping ErrRedisCommandFailed when redis rejects the
// command.
func (c *redisCache[V]) Delete(ctx context.Context, key string) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	err := c.client.Del(ctx, c.keyPrefix+key).Err()
	if err != nil {
		return ErrCommand(err)
	}

	return nil
}

// Has reports whether key is present in the cache. Returns an error wrapping
// ErrRedisCommandFailed when redis rejects the command.
func (c *redisCache[V]) Has(ctx context.Context, key string) (bool, error) {
	cassert.NotNil(c, "redis cache receiver is nil")

	n, err := c.client.Exists(ctx, c.keyPrefix+key).Result()
	if err != nil {
		return false, ErrCommand(err)
	}

	return n > 0, nil
}

// Clear removes every entry registered under this cache's key prefix. Scans
// keys matching "<keyPrefix>*" and deletes them; other caches sharing the
// same DB (with different prefixes) are left untouched. Returns an error
// wrapping ErrRedisCommandFailed on any underlying scan/del failure.
func (c *redisCache[V]) Clear(ctx context.Context) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	iter := c.client.Scan(ctx, 0, c.keyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		err := c.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			return ErrCommand(err)
		}
	}

	err := iter.Err()
	if err != nil {
		return ErrCommand(err)
	}

	return nil
}

// Stop closes the underlying go-redis client. Safe to call more than once;
// subsequent calls are no-ops.
func (c *redisCache[V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	swapped := c.stopped.CompareAndSwap(false, true)
	if !swapped {
		return nil
	}

	err := c.client.Close()
	if err != nil {
		return ErrCommand(err)
	}

	return nil
}
