package limiter

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewLimiter(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil limiter with defaults", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter()
		if l == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("returns independent instances per call", func(t *testing.T) {
		t.Parallel()

		l1 := NewLimiter(WithBurst(1))
		l2 := NewLimiter(WithBurst(1))

		// Drain l1.
		if !l1.Allow() {
			t.Fatal("first Allow on l1 should succeed (burst=1)")
		}
		if l1.Allow() {
			t.Fatal("second Allow on l1 should fail (bucket empty)")
		}

		// l2 is independent; its bucket is still full.
		if !l2.Allow() {
			t.Fatal("Allow on l2 should succeed (independent bucket)")
		}
	})
}

func TestLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("returns true while burst tokens remain", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter(WithRate(1, time.Hour), WithBurst(3))
		for i := range 3 {
			if !l.Allow() {
				t.Fatalf("Allow %d should succeed within burst", i)
			}
		}
	})

	t.Run("returns false after burst is exhausted", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter(WithRate(1, time.Hour), WithBurst(1))
		if !l.Allow() {
			t.Fatal("first Allow should succeed")
		}
		if l.Allow() {
			t.Fatal("second Allow should fail (bucket empty, refill rate is 1/hour)")
		}
	})
}

func TestLimiter_Wait(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when a token is immediately available", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter(WithBurst(1))
		err := l.Wait(context.Background())
		if err != nil {
			t.Fatalf("Wait: %v", err)
		}
	})

	t.Run("returns ErrWaitFailed wrapping ctx error on cancellation", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter(WithRate(1, time.Hour), WithBurst(1))

		// Drain the bucket.
		err := l.Wait(context.Background())
		if err != nil {
			t.Fatalf("first Wait should succeed: %v", err)
		}

		// Next Wait would block ~1 hour; cancel quickly.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = l.Wait(ctx)
		if err == nil {
			t.Fatal("expected Wait to fail when ctx times out")
		}
		if !errors.Is(err, ErrWaitFailed) {
			t.Fatalf("expected wrap of ErrWaitFailed, got %v", err)
		}
	})

	t.Run("returns ErrWaitFailed wrapping ErrContextNil when ctx is nil", func(t *testing.T) {
		t.Parallel()

		l := NewLimiter()
		//nolint:staticcheck // intentionally passing nil ctx to exercise the guard
		err := l.Wait(nil)
		if err == nil {
			t.Fatal("expected Wait to fail with nil ctx")
		}
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected wrap of ErrContextNil, got %v", err)
		}
		if !errors.Is(err, ErrWaitFailed) {
			t.Fatalf("expected wrap of ErrWaitFailed, got %v", err)
		}
	})
}
