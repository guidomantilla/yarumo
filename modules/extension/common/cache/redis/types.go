// Package redis provides a redis-backed implementation of the
// common/cache.Cache[K, V] contract.
//
// The constructor (NewRedisCache) is infallible and performs no I/O: it
// builds the underlying go-redis client (itself lazy) and stores
// configuration. The PING handshake that verifies the server is reachable
// runs in Start per the common/lifecycle.Component contract; consumers
// wire the cache via lifecycle.Build, which dispatches Start and surfaces
// PING failures through errChan as lifecycle.ErrStart.
//
// Keys: string keys, prefixed with "<name>:" by default so multiple
// caches can share a redis DB without colliding. Override with
// WithKeyPrefix. Prefix resolution is delegated to
// common/cache.ResolveKeyPrefix.
//
// Values: serialized via the configured Codec (default JSONCodec) from
// common/cache.
//
// Lifecycle: each cache owns a go-redis client that must be released by
// calling Stop. Stop is idempotent; Done closes after the first Stop
// completes.
package redis

import (
	ccache "github.com/guidomantilla/yarumo/core/common/cache"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// Type compliance: redisCache implements the canonical Cache[string, any]
// shape from common/cache and the lifecycle.Component contract; Error
// implements the standard error contract.
var (
	_ ccache.Cache[string, any] = (*redisCache[any])(nil)
	_ lifecycle.Component       = (*redisCache[any])(nil)
	_ error                     = (*Error)(nil)
)
