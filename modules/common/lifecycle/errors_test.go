package lifecycle

import (
	"errors"
	"strings"
	"testing"
)

func TestErrStart(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrStartFailed", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("root-cause")

		err := ErrStart(cause)
		if !errors.Is(err, ErrStartFailed) {
			t.Fatalf("expected error to wrap ErrStartFailed, got %v", err)
		}
	})

	t.Run("wraps the original cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("original")

		err := ErrStart(cause)
		if !errors.Is(err, cause) {
			t.Fatalf("expected error to wrap original cause, got %v", err)
		}
	})

	t.Run("returns *Error with LifecycleType", func(t *testing.T) {
		t.Parallel()

		err := ErrStart(errors.New("fail"))

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if domErr.Type != LifecycleType {
			t.Fatalf("expected Type=%q, got %q", LifecycleType, domErr.Type)
		}
	})

	t.Run("joins multiple causes", func(t *testing.T) {
		t.Parallel()

		cause1 := errors.New("cause1")
		cause2 := errors.New("cause2")

		err := ErrStart(cause1, cause2)

		if !errors.Is(err, cause1) {
			t.Fatal("expected error to wrap cause1")
		}

		if !errors.Is(err, cause2) {
			t.Fatal("expected error to wrap cause2")
		}

		if !errors.Is(err, ErrStartFailed) {
			t.Fatal("expected error to wrap ErrStartFailed")
		}
	})

	t.Run("error message contains lifecycle type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrStart(errors.New("disk-full"))
		msg := err.Error()

		if !strings.Contains(msg, LifecycleType) {
			t.Fatalf("expected error message to contain %q, got %q", LifecycleType, msg)
		}
	})
}

func TestErrShutdown(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrShutdownFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrShutdown(errors.New("root-cause"))
		if !errors.Is(err, ErrShutdownFailed) {
			t.Fatalf("expected error to wrap ErrShutdownFailed, got %v", err)
		}
	})

	t.Run("wraps ErrShutdownTimeout when given as cause", func(t *testing.T) {
		t.Parallel()

		err := ErrShutdown(ErrShutdownTimeout)
		if !errors.Is(err, ErrShutdownTimeout) {
			t.Fatalf("expected error to wrap ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("returns *Error with LifecycleType", func(t *testing.T) {
		t.Parallel()

		err := ErrShutdown(errors.New("fail"))

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if domErr.Type != LifecycleType {
			t.Fatalf("expected Type=%q, got %q", LifecycleType, domErr.Type)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrStartFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrStartFailed == nil {
			t.Fatal("ErrStartFailed should not be nil")
		}

		if ErrStartFailed.Error() == "" {
			t.Fatal("ErrStartFailed message should not be empty")
		}
	})

	t.Run("ErrShutdownFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrShutdownFailed == nil {
			t.Fatal("ErrShutdownFailed should not be nil")
		}

		if ErrShutdownFailed.Error() == "" {
			t.Fatal("ErrShutdownFailed message should not be empty")
		}
	})

	t.Run("ErrShutdownTimeout is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrShutdownTimeout == nil {
			t.Fatal("ErrShutdownTimeout should not be nil")
		}

		if ErrShutdownTimeout.Error() == "" {
			t.Fatal("ErrShutdownTimeout message should not be empty")
		}
	})
}
