package cache

import (
	"context"
	"io"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/ristretto/v2"
	gocachelib "github.com/eko/gocache/lib/v4/cache"
	libstore "github.com/eko/gocache/lib/v4/store"
	bigcachestore "github.com/eko/gocache/store/bigcache/v4"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	ristrettostore "github.com/eko/gocache/store/ristretto/v4"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	gocache "github.com/patrickmn/go-cache"
)

// backendCache is the minimal interface the wrapper consumes from gocache.
// It mirrors the relevant subset of gocache.CacheInterface[any].
type backendCache interface {
	Get(ctx context.Context, key any) (any, error)
	Set(ctx context.Context, key any, value any, options ...libstore.Option) error
	Delete(ctx context.Context, key any) error
	Clear(ctx context.Context) error
}

// backendInstance bundles a gocache cache with an io.Closer used to release the
// backend's underlying resources during Stop.
type backendInstance struct {
	cache  backendCache
	closer io.Closer
}

// closerFn adapts a no-arg function into an io.Closer.
type closerFn func() error

func (f closerFn) Close() error { return f() }

// noopCloser implements io.Closer with a no-op Close.
type noopCloser struct{}

func (noopCloser) Close() error { return nil }

// buildBackend creates the gocache instance for the configured backend.
//
// Each branch instantiates the concrete in-memory backend, wraps it in the
// gocache adapter, and packages the result with an io.Closer that the public
// Stop method calls to release resources.
func buildBackend(opts *Options) (*backendInstance, error) {
	err := validateOptions(opts)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}

	switch opts.backend {
	case BackendRistretto:
		return buildRistretto(opts)
	case BackendBigcache:
		return buildBigcache(opts)
	case BackendGoCache:
		return buildGoCache(opts)
	default:
		return nil, cerrs.Wrap(ErrUnsupportedBackend)
	}
}

// buildRistretto creates a ristretto-backed gocache instance.
func buildRistretto(opts *Options) (*backendInstance, error) {
	client, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: opts.ristrettoNumCtrs,
		MaxCost:     opts.ristrettoMaxCost,
		BufferItems: opts.ristrettoBufItems,
	})
	if err != nil {
		return nil, cerrs.Wrap(ErrBackendUnavailable, err)
	}

	store := ristrettostore.NewRistretto(client, libstore.WithCost(1), libstore.WithSynchronousSet())
	wrapped := gocachelib.New[any](store)

	closer := closerFn(func() error {
		client.Close()
		return nil
	})

	return &backendInstance{cache: wrapped, closer: closer}, nil
}

// buildBigcache creates a bigcache-backed gocache instance.
func buildBigcache(opts *Options) (*backendInstance, error) {
	cfg := bigcache.DefaultConfig(opts.bigcacheLifeWin)
	cfg.Shards = opts.bigcacheShards
	cfg.CleanWindow = opts.bigcacheCleanWin
	cfg.HardMaxCacheSize = opts.bigcacheMaxSize
	cfg.MaxEntrySize = opts.bigcacheMaxEntry

	client, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, cerrs.Wrap(ErrBackendUnavailable, err)
	}

	store := bigcachestore.NewBigcache(client)
	wrapped := gocachelib.New[any](store)

	return &backendInstance{cache: wrapped, closer: client}, nil
}

// buildGoCache creates a go-cache-backed gocache instance.
func buildGoCache(opts *Options) (*backendInstance, error) {
	client := gocache.New(opts.gocacheDefault, opts.gocacheCleanup)

	store := gocachestore.NewGoCache(client)
	wrapped := gocachelib.New[any](store)

	return &backendInstance{cache: wrapped, closer: noopCloser{}}, nil
}

// setOptionsForTTL returns the libstore options that the wrapper applies on
// every Set call, mapping the wrapper's ttl semantics onto each backend.
func setOptionsForTTL(ttl time.Duration, defaultTTL time.Duration) []libstore.Option {
	effective := ttl
	if effective <= 0 {
		effective = defaultTTL
	}

	return []libstore.Option{
		libstore.WithExpiration(effective),
		libstore.WithCost(1),
		libstore.WithSynchronousSet(),
	}
}
