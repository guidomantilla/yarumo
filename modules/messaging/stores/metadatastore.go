package stores

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// inMemoryMetadataStore is the canonical MetadataStore backed by a
// mutex-guarded map of expiration timestamps plus a background
// sweeper goroutine that evicts expired entries at the configured
// interval. It implements lifecycle.Component (worker-style) because
// of the sweeper: Start spawns the goroutine; Stop signals it to
// exit and waits for it to finish.
type inMemoryMetadataStore struct {
	name          string
	sweepInterval time.Duration

	mu     sync.RWMutex
	expiry map[string]time.Time

	started atomic.Bool
	closed  atomic.Bool

	stopCh chan struct{}
	done   chan struct{}

	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	workerWG sync.WaitGroup
}

// NewInMemoryMetadataStore constructs an in-memory MetadataStore.
// name is used in lifecycle logs and must be non-empty. The returned
// store is not running; call lifecycle.Build (or Start directly) to
// spawn the sweeper goroutine. Heavy-dep backends (Redis SETEX,
// Postgres TTL tables, …) live in extension/messaging/stores/<backend>/.
func NewInMemoryMetadataStore(name string, opts ...Option) MetadataStore {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &inMemoryMetadataStore{
		name:          name,
		sweepInterval: options.sweepInterval,
		expiry:        map[string]time.Time{},
		stopCh:        make(chan struct{}),
		done:          make(chan struct{}),
	}
}

// Name returns the store's identity used in lifecycle logs.
func (s *inMemoryMetadataStore) Name() string {
	cassert.NotNil(s, "in-memory metadata store is nil")

	return s.name
}

// Start spawns the sweeper goroutine. Start is idempotent — a second
// invocation returns nil without spawning a second sweeper. Per the
// lifecycle.Component contract, Start is worker-style: it returns
// immediately and Done is closed after Stop completes.
func (s *inMemoryMetadataStore) Start(_ context.Context) error {
	cassert.NotNil(s, "in-memory metadata store is nil")

	s.startOnce.Do(func() {
		s.started.Store(true)
		s.workerWG.Go(s.sweepLoop)

		go s.awaitDrain()
	})

	return nil
}

// Stop signals the sweeper goroutine to exit and waits for it to
// finish, bounded by ctx's deadline. Stop is idempotent per the
// lifecycle.Component contract. Returns lifecycle.ErrShutdown
// wrapping lifecycle.ErrShutdownTimeout if the sweeper does not exit
// before ctx expires.
func (s *inMemoryMetadataStore) Stop(ctx context.Context) error {
	cassert.NotNil(s, "in-memory metadata store is nil")

	s.stopOnce.Do(func() {
		s.closed.Store(true)
		close(s.stopCh)
	})

	// If Start was never called, there is no sweeper to wait for;
	// close done directly so Done observers can converge.
	if !s.started.Load() {
		s.doneOnce.Do(func() { close(s.done) })

		return nil
	}

	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	}
}

// Done returns the channel that is closed after Stop has drained
// the sweeper goroutine (or immediately on Stop if Start was never
// called).
func (s *inMemoryMetadataStore) Done() <-chan struct{} {
	cassert.NotNil(s, "in-memory metadata store is nil")

	return s.done
}

// Has reports whether key is currently in the store. A key whose TTL
// has expired returns false even before the sweeper has reclaimed
// its slot — the freshness check happens on every call. Has returns
// ErrStore(ErrStoreClosed) when invoked after Stop.
func (s *inMemoryMetadataStore) Has(_ context.Context, key string) (bool, error) {
	cassert.NotNil(s, "in-memory metadata store is nil")

	if s.closed.Load() {
		return false, ErrStore(ErrStoreClosed)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	expiresAt, ok := s.expiry[key]
	if !ok {
		return false, nil
	}

	if time.Now().After(expiresAt) {
		return false, nil
	}

	return true, nil
}

// Add records key with the given TTL. If key already exists, the TTL
// is refreshed (last writer wins). Add returns ErrStore(ErrInvalidTTL)
// when ttl is non-positive and ErrStore(ErrStoreClosed) when invoked
// after Stop.
func (s *inMemoryMetadataStore) Add(_ context.Context, key string, ttl time.Duration) error {
	cassert.NotNil(s, "in-memory metadata store is nil")

	if ttl <= 0 {
		return ErrStore(ErrInvalidTTL)
	}

	if s.closed.Load() {
		return ErrStore(ErrStoreClosed)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.expiry[key] = time.Now().Add(ttl)

	return nil
}

// awaitDrain closes done exactly once after the sweeper goroutine
// has exited.
func (s *inMemoryMetadataStore) awaitDrain() {
	s.workerWG.Wait()
	s.doneOnce.Do(func() { close(s.done) })
}

// sweepLoop runs in its own goroutine and evicts expired entries on
// each tick of the sweep interval. It exits when stopCh is closed.
func (s *inMemoryMetadataStore) sweepLoop() {
	ticker := time.NewTicker(s.sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.sweep()
		}
	}
}

// sweep removes every entry whose expiration timestamp is in the
// past. The pass is single-shot under the write lock; on a busy
// store with many entries this is O(n) per sweep — pick the sweep
// interval accordingly.
func (s *inMemoryMetadataStore) sweep() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, expiresAt := range s.expiry {
		if now.After(expiresAt) {
			delete(s.expiry, key)
		}
	}
}
