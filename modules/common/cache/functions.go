package cache

import (
	"context"
	"time"

	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

// Get delegates to the cache registered under "default" for the requested
// K and V. It returns an error wrapping ErrCacheMiss when the key is absent,
// or an error wrapping ErrCacheTypeAssertion when the registered "default"
// is not a Cache[K, V].
func Get[K comparable, V any](ctx context.Context, key K) (V, error) {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return cpointer.Zero[V](), err
	}
	return c.Get(ctx, key)
}

// Set delegates to the cache registered under "default" for the requested
// K and V and stores value under key with the given TTL. It returns an error
// wrapping ErrCacheTypeAssertion when the registered "default" is not a
// Cache[K, V].
func Set[K comparable, V any](ctx context.Context, key K, value V, ttl time.Duration) error {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return err
	}
	return c.Set(ctx, key, value, ttl)
}

// Delete delegates to the cache registered under "default" for the requested
// K and V and removes the entry at key. It returns an error wrapping
// ErrCacheTypeAssertion when the registered "default" is not a Cache[K, V].
func Delete[K comparable, V any](ctx context.Context, key K) error {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return err
	}
	return c.Delete(ctx, key)
}

// Has delegates to the cache registered under "default" for the requested
// K and V and reports whether key is present. It returns an error wrapping
// ErrCacheTypeAssertion when the registered "default" is not a Cache[K, V].
func Has[K comparable, V any](ctx context.Context, key K) (bool, error) {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return false, err
	}
	return c.Has(ctx, key)
}

// Clear delegates to the cache registered under "default" for the requested
// K and V and removes every entry. It returns an error wrapping
// ErrCacheTypeAssertion when the registered "default" is not a Cache[K, V].
func Clear[K comparable, V any](ctx context.Context) error {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return err
	}
	return c.Clear(ctx)
}

// Stop delegates to the cache registered under "default" for the requested
// K and V and releases its resources. Subsequent facade calls operate on a
// stopped cache and will return implementation-defined errors. It returns an
// error wrapping ErrCacheTypeAssertion when the registered "default" is not
// a Cache[K, V].
func Stop[K comparable, V any](ctx context.Context) error {
	c, err := Lookup[K, V]("default")
	if err != nil {
		return err
	}
	return c.Stop(ctx)
}

// ResolveKeyPrefix returns the namespace prefix backends should prepend to
// every logical key. When configured is non-empty it wins; otherwise the
// default "<name>:" is returned. Backends that share underlying storage
// (redis sharing a DB, memcached sharing an instance) use this to keep
// caches with different names from colliding; backends with per-instance
// storage (in-memory maps) may still apply the prefix for uniformity.
func ResolveKeyPrefix(name, configured string) string {
	if configured != "" {
		return configured
	}

	return name + ":"
}
