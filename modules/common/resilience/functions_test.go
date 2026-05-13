package resilience

import (
	"errors"
	"testing"

	"github.com/sony/gobreaker"
)

func TestValidateName(t *testing.T) {
	t.Parallel()

	t.Run("empty returns sentinel", func(t *testing.T) {
		t.Parallel()

		err := validateName("")
		if err == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(err, ErrRegistryNameEmpty) {
			t.Fatalf("expected errors.Is to match ErrRegistryNameEmpty, got %v", err)
		}
	})

	t.Run("non-empty returns nil", func(t *testing.T) {
		t.Parallel()

		err := validateName("foo")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestValidateExecute(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns sentinel", func(t *testing.T) {
		t.Parallel()

		fn := func() (any, error) { return "noop", errSentinel }

		err := validateExecute(nil, fn)
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("nil fn returns sentinel", func(t *testing.T) {
		t.Parallel()

		err := validateExecute(t.Context(), nil)
		if !errors.Is(err, ErrCircuitBreakerExecuteFnNil) {
			t.Fatalf("expected ErrCircuitBreakerExecuteFnNil, got %v", err)
		}
	})

	t.Run("valid inputs return nil", func(t *testing.T) {
		t.Parallel()

		fn := func() (any, error) { return "noop", errSentinel }

		err := validateExecute(t.Context(), fn)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

// errSentinel is a placeholder error used to satisfy linter rules in tests
// where the function return value is never invoked.
var errSentinel = errors.New("test sentinel")

func TestValidateWait(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns sentinel", func(t *testing.T) {
		t.Parallel()

		err := validateWait(nil)
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("context returns nil", func(t *testing.T) {
		t.Parallel()

		err := validateWait(t.Context())
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestTranslateBreakerError(t *testing.T) {
	t.Parallel()

	t.Run("ErrOpenState mapped to ErrCircuitBreakerOpen", func(t *testing.T) {
		t.Parallel()

		got := translateBreakerError(gobreaker.ErrOpenState)
		if !errors.Is(got, ErrCircuitBreakerOpen) {
			t.Fatalf("expected ErrCircuitBreakerOpen, got %v", got)
		}
	})

	t.Run("ErrTooManyRequests mapped", func(t *testing.T) {
		t.Parallel()

		got := translateBreakerError(gobreaker.ErrTooManyRequests)
		if !errors.Is(got, ErrCircuitBreakerTooManyRequests) {
			t.Fatalf("expected ErrCircuitBreakerTooManyRequests, got %v", got)
		}
	})

	t.Run("other error passed through", func(t *testing.T) {
		t.Parallel()

		original := errors.New("call failed")

		got := translateBreakerError(original)
		if !errors.Is(got, original) {
			t.Fatalf("expected pass-through, got %v", got)
		}
	})
}

func TestFromGobreakerState(t *testing.T) {
	t.Parallel()

	t.Run("closed", func(t *testing.T) {
		t.Parallel()

		got := fromGobreakerState(gobreaker.StateClosed)
		if got != StateClosed {
			t.Fatalf("expected StateClosed, got %v", got)
		}
	})

	t.Run("half-open", func(t *testing.T) {
		t.Parallel()

		got := fromGobreakerState(gobreaker.StateHalfOpen)
		if got != StateHalfOpen {
			t.Fatalf("expected StateHalfOpen, got %v", got)
		}
	})

	t.Run("open", func(t *testing.T) {
		t.Parallel()

		got := fromGobreakerState(gobreaker.StateOpen)
		if got != StateOpen {
			t.Fatalf("expected StateOpen, got %v", got)
		}
	})

	t.Run("unknown defaults to closed", func(t *testing.T) {
		t.Parallel()

		got := fromGobreakerState(gobreaker.State(99))
		if got != StateClosed {
			t.Fatalf("expected StateClosed fallback, got %v", got)
		}
	})
}

func TestSettingsFor(t *testing.T) {
	t.Parallel()

	t.Run("trips at configured failure count", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerConsecutiveFailures(3))
		settings := settingsFor("svc", opts)

		if settings.Name != "svc" {
			t.Fatalf("Name=%q want svc", settings.Name)
		}

		if settings.ReadyToTrip(gobreaker.Counts{ConsecutiveFailures: 2}) {
			t.Fatal("did not expect trip at 2 failures")
		}

		if !settings.ReadyToTrip(gobreaker.Counts{ConsecutiveFailures: 3}) {
			t.Fatal("expected trip at 3 failures")
		}
	})
}

