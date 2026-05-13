package cache

import (
	"context"
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/ristretto/v2"
	gocache "github.com/patrickmn/go-cache"
)

// newRistrettoCache builds a *cache[K, V] backed by dgraph-io/ristretto.
//
// Behaviour is encoded entirely in the function fields populated below; the
// closures capture the ristretto client and any per-backend state (metrics,
// slog hooks via the cache receiver). No internal interface is involved.
func newRistrettoCache[K comparable, V any](opts *Options) (*cache[K, V], error) {
	client, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: opts.ristrettoNumCtrs,
		MaxCost:     opts.ristrettoMaxCost,
		BufferItems: opts.ristrettoBufItems,
	})
	if err != nil {
		return nil, ErrBackend(err)
	}

	c := &cache[K, V]{
		options: opts,
		metrics: newMetricsIfEnabled(opts),
	}

	c.getFn = func(ctx context.Context, key K) (V, error) {
		var zero V

		cacheKey := stringKey(key)

		raw, ok := client.Get(cacheKey)
		if !ok {
			c.recordMiss(ctx, cacheKey)
			return zero, ErrMiss()
		}

		value, aerr := assertValue[V](raw)
		if aerr != nil {
			c.recordMiss(ctx, cacheKey)
			return zero, aerr
		}

		c.recordHit(ctx, cacheKey)
		return value, nil
	}

	c.setFn = func(ctx context.Context, key K, value V, ttl time.Duration) error {
		cacheKey := stringKey(key)
		effective := effectiveTTL(ttl, opts.ttl)

		ok := client.SetWithTTL(cacheKey, any(value), 1, effective)
		if !ok {
			return ErrBackend(errRistrettoSetRejected)
		}

		// Ristretto applies writes asynchronously via a buffered admission
		// channel; Wait blocks until previously buffered writes have been
		// fully applied, matching the synchronous-set semantics the eko
		// adapter previously requested via libstore.WithSynchronousSet.
		client.Wait()

		c.recordSet(ctx, cacheKey)
		return nil
	}

	c.deleteFn = func(ctx context.Context, key K) error {
		cacheKey := stringKey(key)
		client.Del(cacheKey)
		c.recordEviction(ctx, cacheKey)
		return nil
	}

	c.hasFn = func(_ context.Context, key K) bool {
		_, ok := client.Get(stringKey(key))
		return ok
	}

	c.clearFn = func(ctx context.Context) error {
		client.Clear()
		c.recordEviction(ctx, "*")
		return nil
	}

	c.stopFn = func(ctx context.Context) error {
		client.Close()
		c.recordStopped(ctx)
		return nil
	}

	return c, nil
}

// newBigcacheCache builds a *cache[K, V] backed by allegro/bigcache.
//
// bigcache stores raw []byte payloads natively, so the closure asserts that V
// is convertible to []byte via a type-assertion on the boxed value. Mismatched
// types surface as ErrSerialize so the public contract stays consistent across
// backends.
func newBigcacheCache[K comparable, V any](opts *Options) (*cache[K, V], error) {
	cfg := bigcache.DefaultConfig(opts.bigcacheLifeWin)
	cfg.Shards = opts.bigcacheShards
	cfg.CleanWindow = opts.bigcacheCleanWin
	cfg.HardMaxCacheSize = opts.bigcacheMaxSize
	cfg.MaxEntrySize = opts.bigcacheMaxEntry

	client, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, ErrBackend(err)
	}

	c := &cache[K, V]{
		options: opts,
		metrics: newMetricsIfEnabled(opts),
	}

	c.getFn = func(ctx context.Context, key K) (V, error) {
		var zero V

		cacheKey := stringKey(key)

		raw, gerr := client.Get(cacheKey)
		if gerr != nil {
			if errors.Is(gerr, bigcache.ErrEntryNotFound) {
				c.recordMiss(ctx, cacheKey)
				return zero, ErrMiss()
			}
			return zero, ErrCache(gerr)
		}

		// bigcache stores []byte natively; assertValue cross-validates that V
		// is a []byte (the documented bigcache contract).
		value, aerr := assertValue[V](raw)
		if aerr != nil {
			c.recordMiss(ctx, cacheKey)
			return zero, aerr
		}

		c.recordHit(ctx, cacheKey)
		return value, nil
	}

	c.setFn = func(ctx context.Context, key K, value V, _ time.Duration) error {
		cacheKey := stringKey(key)

		// bigcache only stores []byte and applies a global TTL configured at
		// construction (no per-entry TTL). The wrapper's per-call ttl is
		// accepted for API parity but intentionally ignored here.
		payload, ok := any(value).([]byte)
		if !ok {
			return ErrSerialize(errors.New("bigcache requires []byte values"))
		}

		serr := client.Set(cacheKey, payload)
		if serr != nil {
			return ErrCache(serr)
		}

		c.recordSet(ctx, cacheKey)
		return nil
	}

	c.deleteFn = func(ctx context.Context, key K) error {
		cacheKey := stringKey(key)

		derr := client.Delete(cacheKey)
		if derr != nil && !errors.Is(derr, bigcache.ErrEntryNotFound) {
			return ErrCache(derr)
		}

		c.recordEviction(ctx, cacheKey)
		return nil
	}

	c.hasFn = func(_ context.Context, key K) bool {
		_, gerr := client.Get(stringKey(key))
		return gerr == nil
	}

	c.clearFn = func(ctx context.Context) error {
		rerr := client.Reset()
		if rerr != nil {
			return ErrCache(rerr)
		}

		c.recordEviction(ctx, "*")
		return nil
	}

	c.stopFn = func(ctx context.Context) error {
		cerr := client.Close()
		if cerr != nil {
			return ErrCache(cerr)
		}

		c.recordStopped(ctx)
		return nil
	}

	return c, nil
}

// newGoCacheCache builds a *cache[K, V] backed by patrickmn/go-cache.
//
// go-cache stores values as `any`, so the closure performs a direct
// type-assertion on the retrieved interface value to recover V.
func newGoCacheCache[K comparable, V any](opts *Options) (*cache[K, V], error) {
	client := gocache.New(opts.gocacheDefault, opts.gocacheCleanup)

	c := &cache[K, V]{
		options: opts,
		metrics: newMetricsIfEnabled(opts),
	}

	c.getFn = func(ctx context.Context, key K) (V, error) {
		var zero V

		cacheKey := stringKey(key)

		raw, ok := client.Get(cacheKey)
		if !ok {
			c.recordMiss(ctx, cacheKey)
			return zero, ErrMiss()
		}

		value, aerr := assertValue[V](raw)
		if aerr != nil {
			c.recordMiss(ctx, cacheKey)
			return zero, aerr
		}

		c.recordHit(ctx, cacheKey)
		return value, nil
	}

	c.setFn = func(ctx context.Context, key K, value V, ttl time.Duration) error {
		cacheKey := stringKey(key)
		effective := effectiveTTL(ttl, opts.ttl)

		client.Set(cacheKey, any(value), effective)
		c.recordSet(ctx, cacheKey)
		return nil
	}

	c.deleteFn = func(ctx context.Context, key K) error {
		cacheKey := stringKey(key)
		client.Delete(cacheKey)
		c.recordEviction(ctx, cacheKey)
		return nil
	}

	c.hasFn = func(_ context.Context, key K) bool {
		_, ok := client.Get(stringKey(key))
		return ok
	}

	c.clearFn = func(ctx context.Context) error {
		client.Flush()
		c.recordEviction(ctx, "*")
		return nil
	}

	c.stopFn = func(ctx context.Context) error {
		// go-cache holds no external resources; just record the stop.
		c.recordStopped(ctx)
		return nil
	}

	return c, nil
}

// newMetricsIfEnabled returns an otelMetrics adapter when WithOTel was set on
// opts, or nil otherwise. Centralised here so each factory's wiring stays
// compact.
func newMetricsIfEnabled(opts *Options) *otelMetrics {
	if !opts.otelEnabled {
		return nil
	}
	return newOtelMetrics(opts.otelMeterName)
}

// effectiveTTL resolves the per-call ttl, falling back to the cache default
// when ttl is non-positive.
func effectiveTTL(ttl time.Duration, defaultTTL time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}
	return defaultTTL
}

// errRistrettoSetRejected is returned wrapped in ErrBackend when ristretto
// rejects a SetWithTTL call (most commonly because the admission policy
// dropped the entry).
var errRistrettoSetRejected = errors.New("ristretto rejected set")
