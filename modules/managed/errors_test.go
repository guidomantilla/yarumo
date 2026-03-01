package managed

import (
	"errors"
	"testing"
)

func TestErrListen(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("port in use")
		err := ErrListen(cause)

		var me *Error

		ok := errors.As(err, &me)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if me.Type != ManagedType {
			t.Fatalf("Type = %q, want %q", me.Type, ManagedType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrListenFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrListen()

		if !errors.Is(err, ErrListenFailed) {
			t.Fatalf("expected ErrListenFailed in chain: %v", err)
		}
	})
}

func TestErrServe(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("bind error")
		err := ErrServe(cause)

		var me *Error

		ok := errors.As(err, &me)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if me.Type != ManagedType {
			t.Fatalf("Type = %q, want %q", me.Type, ManagedType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrServeFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrServe()

		if !errors.Is(err, ErrServeFailed) {
			t.Fatalf("expected ErrServeFailed in chain: %v", err)
		}
	})
}

func TestErrShutdown(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("context deadline exceeded")
		err := ErrShutdown(ErrShutdownTimeout, cause)

		var me *Error

		ok := errors.As(err, &me)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if me.Type != ManagedType {
			t.Fatalf("Type = %q, want %q", me.Type, ManagedType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrShutdownFailed) || !errors.Is(err, ErrShutdownTimeout) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrShutdown()

		if !errors.Is(err, ErrShutdownFailed) {
			t.Fatalf("expected ErrShutdownFailed in chain: %v", err)
		}
	})
}

func TestErrStart(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("recorder error")
		err := ErrStart(cause)

		var me *Error

		ok := errors.As(err, &me)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if me.Type != ManagedType {
			t.Fatalf("Type = %q, want %q", me.Type, ManagedType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrStartFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrStart()

		if !errors.Is(err, ErrStartFailed) {
			t.Fatalf("expected ErrStartFailed in chain: %v", err)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("matched via errors.Is", func(t *testing.T) {
		t.Parallel()

		joined := errors.Join(ErrShutdownTimeout, ErrShutdownFailed)
		if !errors.Is(joined, ErrShutdownTimeout) || !errors.Is(joined, ErrShutdownFailed) {
			t.Fatalf("sentinel errors are not matched via errors.Is")
		}
	})
}
