package breaker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewBreaker(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil breaker with defaults", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		if b == nil {
			t.Fatal("expected non-nil breaker")
		}
	})

	t.Run("starts in closed state", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		if b.State() != StateClosed {
			t.Fatalf("State = %s, want %s", b.State(), StateClosed)
		}
	})

	t.Run("returns independent instances per call", func(t *testing.T) {
		t.Parallel()

		b1 := NewBreaker(WithConsecutiveFailures(2))
		b2 := NewBreaker(WithConsecutiveFailures(2))

		// Trip b1.
		for range 2 {
			_ = b1.Execute(context.Background(), func() error { return errors.New("fail") })
		}

		if b1.State() != StateOpen {
			t.Fatalf("b1 State = %s, want open", b1.State())
		}
		if b2.State() != StateClosed {
			t.Fatalf("b2 State = %s, want closed (independent)", b2.State())
		}
	})
}

func TestBreaker_Execute(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when fn succeeds and breaker is closed", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		err := b.Execute(context.Background(), func() error { return nil })
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if b.State() != StateClosed {
			t.Fatalf("State = %s, want closed", b.State())
		}
	})

	t.Run("forwards fn error wrapped in ErrBreakerFailed while still closed", func(t *testing.T) {
		t.Parallel()

		want := errors.New("boom")
		b := NewBreaker(WithConsecutiveFailures(5))

		err := b.Execute(context.Background(), func() error { return want })
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if !errors.Is(err, want) {
			t.Fatalf("expected wrap of original cause, got %v", err)
		}
		if !errors.Is(err, ErrBreakerFailed) {
			t.Fatalf("expected wrap of ErrBreakerFailed, got %v", err)
		}
		if b.State() != StateClosed {
			t.Fatal("breaker should still be closed (one failure < threshold)")
		}
	})

	t.Run("trips to open after consecutive-failures threshold", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker(WithConsecutiveFailures(3))

		for range 3 {
			_ = b.Execute(context.Background(), func() error { return errors.New("fail") })
		}

		if b.State() != StateOpen {
			t.Fatalf("State = %s, want open", b.State())
		}
	})

	t.Run("rejects fast with ErrBreakerOpen while open", func(t *testing.T) {
		t.Parallel()

		var ran atomic.Bool
		b := NewBreaker(WithConsecutiveFailures(2))

		// Trip.
		for range 2 {
			_ = b.Execute(context.Background(), func() error { return errors.New("fail") })
		}

		err := b.Execute(context.Background(), func() error {
			ran.Store(true)
			return nil
		})

		if !errors.Is(err, ErrBreakerOpen) {
			t.Fatalf("expected wrap of ErrBreakerOpen, got %v", err)
		}
		if ran.Load() {
			t.Fatal("fn should not have been invoked while breaker is open")
		}
	})

	t.Run("transitions to half-open after timeout and recovers on success", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker(
			WithConsecutiveFailures(2),
			WithTimeout(50*time.Millisecond),
			WithMaxRequests(1),
		)

		for range 2 {
			_ = b.Execute(context.Background(), func() error { return errors.New("fail") })
		}
		if b.State() != StateOpen {
			t.Fatalf("expected open before timeout, got %s", b.State())
		}

		time.Sleep(70 * time.Millisecond)

		// First call probes (half-open), succeeds → breaker closes.
		err := b.Execute(context.Background(), func() error { return nil })
		if err != nil {
			t.Fatalf("probe call: %v", err)
		}
		if b.State() != StateClosed {
			t.Fatalf("State = %s, want closed after successful probe", b.State())
		}
	})

	t.Run("returns ErrBreakerFailed wrapping ErrContextNil when ctx is nil", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		//nolint:staticcheck // intentionally passing nil ctx to exercise the guard
		err := b.Execute(nil, func() error { return nil })
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected wrap of ErrContextNil, got %v", err)
		}
		if !errors.Is(err, ErrBreakerFailed) {
			t.Fatalf("expected wrap of ErrBreakerFailed, got %v", err)
		}
	})

	t.Run("returns ErrBreakerFailed wrapping ErrFnNil when fn is nil", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		err := b.Execute(context.Background(), nil)
		if !errors.Is(err, ErrFnNil) {
			t.Fatalf("expected wrap of ErrFnNil, got %v", err)
		}
		if !errors.Is(err, ErrBreakerFailed) {
			t.Fatalf("expected wrap of ErrBreakerFailed, got %v", err)
		}
	})

	t.Run("records ctx cancellation as a failure", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := b.Execute(ctx, func() error { return nil })
		if err == nil {
			t.Fatal("expected error when ctx is canceled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected wrap of context.Canceled, got %v", err)
		}
	})
}

func TestBreaker_State(t *testing.T) {
	t.Parallel()

	t.Run("returns closed initially", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker()
		if b.State() != StateClosed {
			t.Fatalf("State = %s, want closed", b.State())
		}
	})

	t.Run("returns open after tripping", func(t *testing.T) {
		t.Parallel()

		b := NewBreaker(WithConsecutiveFailures(1))
		_ = b.Execute(context.Background(), func() error { return errors.New("fail") })

		if b.State() != StateOpen {
			t.Fatalf("State = %s, want open", b.State())
		}
	})
}

func TestBreaker_OnStateChangeHook(t *testing.T) {
	t.Parallel()

	var transitions atomic.Int32
	var lastFrom, lastTo atomic.Int32

	b := NewBreaker(
		WithName("payments"),
		WithConsecutiveFailures(1),
		WithOnStateChange(func(_ string, from, to State) {
			transitions.Add(1)
			lastFrom.Store(int32(from))
			lastTo.Store(int32(to))
		}),
	)

	_ = b.Execute(context.Background(), func() error { return errors.New("fail") })

	if transitions.Load() < 1 {
		t.Fatalf("transitions = %d, want at least 1 (closed → open)", transitions.Load())
	}
	if State(lastTo.Load()) != StateOpen {
		t.Fatalf("lastTo = %s, want open", State(lastTo.Load()))
	}
}
