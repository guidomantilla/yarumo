package cache

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/common/log"
	"github.com/guidomantilla/yarumo/managed"
)

// BuildCache creates a Cache[K, V] and returns it together with a
// managed.StopFn that releases its backend resources on shutdown. This mirrors
// the BuildXxx builder shape used across modules/managed: it logs lifecycle
// events and the returned StopFn disposes of the backend with a
// timeout-bounded context.
//
// Unlike the workers in modules/managed, BuildCache does not start a goroutine;
// the cache is fully ready when this function returns and no asynchronous
// error reporting channel is needed.
func BuildCache[K comparable, V any](ctx context.Context, name string, opts ...Option) (Cache[K, V], managed.StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	cacheValue, err := NewCache[K, V](opts...)
	if err != nil {
		clog.Error(ctx, "failed to build cache", "stage", "startup", "component", name, "error", err)
		return nil, nil, err
	}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		stopErr := cacheValue.Stop(timeoutCtx)
		if stopErr != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", stopErr)
		}
	}

	return cacheValue, stopFn, nil
}
