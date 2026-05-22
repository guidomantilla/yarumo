package diagnostics

import (
	"context"
	"runtime"
	"sync"
	"testing"
)

// blockProfRateMu serialises tests that mutate runtime.SetBlockProfileRate.
var blockProfRateMu sync.Mutex

func TestNewBlockProfiling(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil sampler", func(t *testing.T) {
		t.Parallel()

		s := NewBlockProfiling("blockprof-1")
		if s == nil {
			t.Fatal("expected non-nil sampler")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		s := NewBlockProfiling("blockprof-named")
		if s.Name() != "blockprof-named" {
			t.Fatalf("expected name %q, got %q", "blockprof-named", s.Name())
		}
	})

	t.Run("default rate", func(t *testing.T) {
		t.Parallel()

		s := NewBlockProfiling("blockprof-rate-default")
		if s.Rate() != blockProfileRateDefault {
			t.Fatalf("expected rate %d, got %d", blockProfileRateDefault, s.Rate())
		}
	})

	t.Run("custom rate", func(t *testing.T) {
		t.Parallel()

		s := NewBlockProfiling("blockprof-rate-custom", WithBlockProfileRate(1000))
		if s.Rate() != 1000 {
			t.Fatalf("expected rate %d, got %d", 1000, s.Rate())
		}
	})

	t.Run("done channel is open at construction", func(t *testing.T) {
		t.Parallel()

		s := NewBlockProfiling("blockprof-open-done")

		select {
		case <-s.Done():
			t.Fatal("expected Done channel to be open before Start/Stop")
		default:
		}
	})
}

func TestBlockProfiling_Lifecycle(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		blockProfRateMu.Lock()
		defer blockProfRateMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		s := NewBlockProfiling("blockprof-lifecycle", WithBlockProfileRate(100))

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("unexpected stop error: %v", err)
		}

		select {
		case <-s.Done():
		default:
			t.Fatal("expected Done channel closed after Stop")
		}
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		blockProfRateMu.Lock()
		defer blockProfRateMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		s := NewBlockProfiling("blockprof-stop-twice")

		err := s.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected start error: %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = s.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})
}

func TestBlockProfiling_Done(t *testing.T) {
	t.Parallel()

	t.Run("unblocks readers after Stop", func(t *testing.T) {
		t.Parallel()

		blockProfRateMu.Lock()
		defer blockProfRateMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		s := NewBlockProfiling("blockprof-done")

		ready := make(chan struct{})
		done := make(chan struct{})

		go func() {
			close(ready)
			<-s.Done()
			close(done)
		}()

		<-ready

		err := s.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		<-done
	})
}
