package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ccache "github.com/guidomantilla/yarumo/common/cache"
	"github.com/guidomantilla/yarumo/common/lifecycle"
	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// redisCache is a redis-backed Cache[string, V] implementation. Keys are
// stored under the resolved key prefix (default "<name>:") so that
// multiple caches can share a redis DB without colliding. Values are
// serialized via the configured Codec (default JSONCodec). Safe for
// concurrent use; Start performs a PING handshake against the configured
// server; Stop is idempotent and closes the underlying go-redis client.
type redisCache[V any] struct {
	name      string
	keyPrefix string
	options   *Options
	client    *redis.Client
	codec     ccache.Codec

	done chan struct{}
	once sync.Once
}

// NewRedisCache constructs a redis-backed Cache[string, V] under the
// given name. The constructor performs no I/O: it builds the underlying
// go-redis client (which is itself lazy) and stores configuration. The
// PING handshake that verifies the server is reachable runs in Start,
// per the lifecycle.Component contract; callers wire the cache through
// lifecycle.Build so PING failures surface via errChan and the standard
// lifecycle.ErrStart wrapper.
func NewRedisCache[V any](name string, opts ...Option) ccache.Cache[string, V] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	client := redis.NewClient(&redis.Options{
		Addr:     options.redisAddr,
		Password: options.redisPassword,
		DB:       options.redisDB,
	})

	return &redisCache[V]{
		name:      name,
		keyPrefix: ccache.ResolveKeyPrefix(name, options.keyPrefix),
		options:   options,
		client:    client,
		codec:     options.codec,
		done:      make(chan struct{}),
	}
}

// Name returns the cache name supplied to NewRedisCache.
func (c *redisCache[V]) Name() string {
	cassert.NotNil(c, "redis cache receiver is nil")

	return c.name
}

// Start verifies connectivity to the configured redis server by issuing
// a PING. It satisfies the lifecycle.Component worker-style contract:
// returns immediately on success, or returns a lifecycle.ErrStart
// wrapping ErrRedisCommandFailed when the server does not respond.
func (c *redisCache[V]) Start(ctx context.Context) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	err := c.client.Ping(ctx).Err()
	if err != nil {
		return lifecycle.ErrStart(ErrCommand(err))
	}

	return nil
}

// Stop closes the underlying go-redis client and closes Done. It is
// idempotent: only the first call closes the client; subsequent calls
// are no-ops returning nil.
func (c *redisCache[V]) Stop(_ context.Context) error {
	cassert.NotNil(c, "redis cache receiver is nil")

	var closeErr error

	c.once.Do(func() {
		closeErr = c.client.Close()
		close(c.done)
	})

	if closeErr != nil {
		return ErrCommand(closeErr)
	}

	return nil
}

// Done returns the channel that is closed after Stop has been called.
func (c *redisCache[V]) Done() <-chan struct{} {
	cassert.NotNil(c, "redis cache receiver is nil")

	return c.done
}

// Get returns the value stored at key or an error wrapping ErrCacheMiss
// when the key is absent. Returns an error wrapping ErrRedisCommandFailed
// for transport/protocol errors, or ErrRedisDecodeFailed when the stored
// bytes cannot be decoded into V via the configured codec.
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

// Set stores value under key with the given TTL. A non-positive ttl
// resolves to the cache default configured via WithTTL. Returns an error
// wrapping ErrRedisEncodeFailed when the codec cannot encode value, or
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

// Has reports whether key is present in the cache. Returns an error
// wrapping ErrRedisCommandFailed when redis rejects the command.
func (c *redisCache[V]) Has(ctx context.Context, key string) (bool, error) {
	cassert.NotNil(c, "redis cache receiver is nil")

	n, err := c.client.Exists(ctx, c.keyPrefix+key).Result()
	if err != nil {
		return false, ErrCommand(err)
	}

	return n > 0, nil
}

// Clear removes every entry registered under this cache's key prefix.
// Scans keys matching "<keyPrefix>*" and deletes them; other caches
// sharing the same DB (with different prefixes) are left untouched.
// Returns an error wrapping ErrRedisCommandFailed on any underlying
// scan/del failure.
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
