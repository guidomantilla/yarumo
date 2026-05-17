package cache

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Package-level registry storage. caches holds the named heterogeneous
// registry (values are Cache[K, V] for varying K and V, stored as any); lock
// guards every access to the map. INMEMORY_CACHE is preregistered under two
// aliases — "default" so the package-level facade resolves to it out of the
// box, and INMEMORY_CACHE.Name() so consumers can look it up by its own name.
// Both keys point to the same instance.
var (
	caches = map[string]any{
		"default":             INMEMORY_CACHE,
		INMEMORY_CACHE.Name(): INMEMORY_CACHE,
	}
	lock = new(sync.RWMutex)
)

// Register adds a Cache[K, V] to the registry under the given name. K and V
// are inferred from c. Re-registering replaces the previous entry under name
// (regardless of its prior K/V). The package-level facade is not affected.
func Register[K comparable, V any](name string, c Cache[K, V]) {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(c, "cache is nil")

	lock.Lock()
	defer lock.Unlock()

	caches[name] = c
}

// Lookup retrieves a registered Cache[K, V] by name. It returns an error
// wrapping ErrCacheNotRegistered when no cache is registered under name, or
// an error wrapping ErrCacheTypeAssertion when the registered cache's type
// parameters do not match the requested K and V.
func Lookup[K comparable, V any](name string) (Cache[K, V], error) {
	cassert.NotEmpty(name, "name is empty")

	lock.RLock()
	defer lock.RUnlock()

	raw, ok := caches[name]
	if !ok {
		return nil, ErrNotRegistered(name)
	}

	c, ok := raw.(Cache[K, V])
	if !ok {
		return nil, ErrTypeAssertion()
	}

	return c, nil
}

// Supported returns the names of all registered caches in unspecified order.
func Supported() []string {
	lock.RLock()
	defer lock.RUnlock()

	names := make([]string, 0, len(caches))
	for name := range caches {
		names = append(names, name)
	}
	return names
}
