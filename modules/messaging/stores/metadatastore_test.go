package stores

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

func TestNewInMemoryMetadataStore(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil store", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMetadataStore("test")
		if s == nil {
			t.Fatal("expected non-nil store")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMetadataStore("dedup")

		c, ok := s.(lifecycle.Component)
		if !ok {
			t.Fatal("expected lifecycle.Component implementation")
		}

		if c.Name() != "dedup" {
			t.Fatalf("expected name dedup, got %q", c.Name())
		}
	})

	t.Run("uses default sweep interval when no option supplied", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMetadataStore("test").(*inMemoryMetadataStore)
		if s.sweepInterval != defaultSweepInterval {
			t.Fatalf("expected default sweep interval %v, got %v", defaultSweepInterval, s.sweepInterval)
		}
	})

	t.Run("honors WithSweepInterval", func(t *testing.T) {
		t.Parallel()

		want := 50 * time.Millisecond

		s := NewInMemoryMetadataStore("test", WithSweepInterval(want)).(*inMemoryMetadataStore)
		if s.sweepInterval != want {
			t.Fatalf("expected sweep interval %v, got %v", want, s.sweepInterval)
		}
	})

	t.Run("WithSweepInterval(0) preserves default", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryMetadataStore("test", WithSweepInterval(0)).(*inMemoryMetadataStore)
		if s.sweepInterval != defaultSweepInterval {
			t.Fatalf("expected default preserved on zero arg, got %v", s.sweepInterval)
		}
	})
}

func TestInMemoryMetadataStore_Add(t *testing.T) {
	t.Parallel()

	t.Run("records a key that Has reports as present", func(t *testing.T) {
		t.Parallel()

		s := startedStore(t)

		err := s.Add(context.Background(), "k", time.Second)
		if err != nil {
			t.Fatalf("add: %v", err)
		}

		ok, err := s.Has(context.Background(), "k")
		if err != nil {
			t.Fatalf("has: %v", err)
		}

		if !ok {
			t.Fatal("expected key to be present after Add")
		}
	})

	t.Run("rejects non-positive TTL with ErrInvalidTTL", func(t *testing.T) {
		t.Parallel()

		s := startedStore(t)

		err := s.Add(context.Background(), "k", 0)
		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatalf("expected ErrInvalidTTL for zero ttl, got %v", err)
		}

		err = s.Add(context.Background(), "k", -time.Second)
		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatalf("expected ErrInvalidTTL for negative ttl, got %v", err)
		}
	})

	t.Run("refreshes TTL when key already exists", func(t *testing.T) {
		t.Parallel()

		s := startedStore(t, WithSweepInterval(10*time.Millisecond))

		err := s.Add(context.Background(), "k", 30*time.Millisecond)
		if err != nil {
			t.Fatalf("first add: %v", err)
		}

		time.Sleep(20 * time.Millisecond)

		err = s.Add(context.Background(), "k", 100*time.Millisecond)
		if err != nil {
			t.Fatalf("refresh add: %v", err)
		}

		// Past the original TTL deadline; still present because of refresh.
		time.Sleep(25 * time.Millisecond)

		ok, err := s.Has(context.Background(), "k")
		if err != nil {
			t.Fatalf("has: %v", err)
		}

		if !ok {
			t.Fatal("expected refreshed key to still be present")
		}
	})
}

func TestInMemoryMetadataStore_Has(t *testing.T) {
	t.Parallel()

	t.Run("returns false for unknown key", func(t *testing.T) {
		t.Parallel()

		s := startedStore(t)

		ok, err := s.Has(context.Background(), "missing")
		if err != nil {
			t.Fatalf("has: %v", err)
		}

		if ok {
			t.Fatal("expected unknown key to be absent")
		}
	})

	t.Run("returns false after TTL expires (even before sweep)", func(t *testing.T) {
		t.Parallel()

		// Long sweep interval so the per-call freshness check is what
		// is being exercised, not the background sweeper.
		s := startedStore(t, WithSweepInterval(time.Hour))

		err := s.Add(context.Background(), "k", 10*time.Millisecond)
		if err != nil {
			t.Fatalf("add: %v", err)
		}

		time.Sleep(30 * time.Millisecond)

		ok, err := s.Has(context.Background(), "k")
		if err != nil {
			t.Fatalf("has: %v", err)
		}

		if ok {
			t.Fatal("expected key to be reported absent after TTL")
		}
	})

	t.Run("sweeper reclaims expired entries", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(10*time.Millisecond)).(*inMemoryMetadataStore)
		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			_ = raw.Stop(ctx)
		})

		err = raw.Add(context.Background(), "k", 20*time.Millisecond)
		if err != nil {
			t.Fatalf("add: %v", err)
		}

		// Wait long enough for the sweeper to run at least one tick
		// past the TTL.
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			raw.mu.RLock()
			n := len(raw.expiry)
			raw.mu.RUnlock()

			if n == 0 {
				return
			}

			time.Sleep(10 * time.Millisecond)
		}

		t.Fatal("sweeper did not evict expired entry in time")
	})
}

func TestInMemoryMetadataStore_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(time.Hour)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = raw.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("stop: %v", err)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(time.Hour)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop has drained the sweeper", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(10*time.Millisecond)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-raw.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-raw.Done():
		case <-time.After(time.Second):
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop without prior Start closes Done immediately", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test").(*inMemoryMetadataStore)

		err := raw.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-raw.Done():
		default:
			t.Fatal("Done not closed after Stop on never-started store")
		}
	})

	t.Run("Has after Stop returns ErrStoreClosed", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(time.Hour)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		_, err = raw.Has(context.Background(), "k")
		if !errors.Is(err, ErrStoreClosed) {
			t.Fatalf("expected ErrStoreClosed after Stop, got %v", err)
		}
	})

	t.Run("Add after Stop returns ErrStoreClosed", func(t *testing.T) {
		t.Parallel()

		raw := NewInMemoryMetadataStore("test", WithSweepInterval(time.Hour)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = raw.Stop(ctx)
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		err = raw.Add(context.Background(), "k", time.Second)
		if !errors.Is(err, ErrStoreClosed) {
			t.Fatalf("expected ErrStoreClosed after Stop, got %v", err)
		}
	})

	t.Run("Stop returns ErrShutdownTimeout when ctx is already expired and sweeper has not drained", func(t *testing.T) {
		t.Parallel()

		// Long sweep interval so the sweeper does not exit on its own
		// before we observe the timeout. Stop signals via stopCh and
		// the goroutine will wake on the next tick (~1h) — far longer
		// than the cancelled ctx.
		raw := NewInMemoryMetadataStore("test", WithSweepInterval(time.Hour)).(*inMemoryMetadataStore)

		err := raw.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = raw.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}

		// Drain the sweeper cleanly so the t.Cleanup leak detector is happy.
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanCancel()

		_ = raw.Stop(cleanCtx)
	})
}

func TestInMemoryMetadataStore_Concurrent(t *testing.T) {
	t.Parallel()

	t.Run("Add + Has are race-free", func(t *testing.T) {
		t.Parallel()

		s := startedStore(t, WithSweepInterval(5*time.Millisecond))

		const workers = 16
		const ops = 200

		var wg sync.WaitGroup

		for w := range workers {
			wg.Add(1)

			go func(id int) {
				defer wg.Done()

				for i := range ops {
					key := strconv.Itoa(id) + "-" + strconv.Itoa(i)

					err := s.Add(context.Background(), key, 50*time.Millisecond)
					if err != nil {
						t.Errorf("add: %v", err)

						return
					}

					_, err = s.Has(context.Background(), key)
					if err != nil {
						t.Errorf("has: %v", err)

						return
					}
				}
			}(w)
		}

		wg.Wait()
	})
}

// startedStore constructs an in-memory metadata store, applies opts,
// calls Start with a fresh context, and registers a t.Cleanup that
// Stops the store with a bounded timeout.
func startedStore(t *testing.T, opts ...Option) MetadataStore {
	t.Helper()

	s := NewInMemoryMetadataStore("test", opts...)

	c, ok := s.(lifecycle.Component)
	if !ok {
		t.Fatal("expected lifecycle.Component implementation")
	}

	err := c.Start(context.Background())
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = c.Stop(ctx)
	})

	return s
}
