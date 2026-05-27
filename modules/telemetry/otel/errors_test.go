package otel

import (
	"errors"
	"testing"
)

func TestErrResource(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("resource error")
		err := ErrResource(cause)

		var oe *Error

		ok := errors.As(err, &oe)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if oe.Type != OtelType {
			t.Fatalf("Type = %q, want %q", oe.Type, OtelType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrResourceFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrResource()

		if !errors.Is(err, ErrResourceFailed) {
			t.Fatalf("expected ErrResourceFailed in chain: %v", err)
		}
	})
}

func TestErrTracer(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("tracer error")
		err := ErrTracer(cause)

		var oe *Error

		ok := errors.As(err, &oe)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if oe.Type != OtelType {
			t.Fatalf("Type = %q, want %q", oe.Type, OtelType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrTracerFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrTracer()

		if !errors.Is(err, ErrTracerFailed) {
			t.Fatalf("expected ErrTracerFailed in chain: %v", err)
		}
	})
}

func TestErrMeter(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("meter error")
		err := ErrMeter(cause)

		var oe *Error

		ok := errors.As(err, &oe)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if oe.Type != OtelType {
			t.Fatalf("Type = %q, want %q", oe.Type, OtelType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrMeterFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrMeter()

		if !errors.Is(err, ErrMeterFailed) {
			t.Fatalf("expected ErrMeterFailed in chain: %v", err)
		}
	})
}

func TestErrLogger(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("logger error")
		err := ErrLogger(cause)

		var oe *Error

		ok := errors.As(err, &oe)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if oe.Type != OtelType {
			t.Fatalf("Type = %q, want %q", oe.Type, OtelType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrLoggerFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrLogger()

		if !errors.Is(err, ErrLoggerFailed) {
			t.Fatalf("expected ErrLoggerFailed in chain: %v", err)
		}
	})
}

func TestErrObserve(t *testing.T) {
	t.Parallel()

	t.Run("joins errors with type", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("observe error")
		err := ErrObserve(ErrTracerFailed, cause)

		var oe *Error

		ok := errors.As(err, &oe)
		if !ok {
			t.Fatalf("errors.As to *Error failed: %T", err)
		}

		if oe.Type != OtelType {
			t.Fatalf("Type = %q, want %q", oe.Type, OtelType)
		}

		if !errors.Is(err, cause) || !errors.Is(err, ErrObserveFailed) || !errors.Is(err, ErrTracerFailed) {
			t.Fatalf("joined error does not match components: %v", err)
		}
	})

	t.Run("no args wraps sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrObserve()

		if !errors.Is(err, ErrObserveFailed) {
			t.Fatalf("expected ErrObserveFailed in chain: %v", err)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("matched via errors.Is", func(t *testing.T) {
		t.Parallel()

		joined := errors.Join(ErrResourceFailed, ErrTracerFailed)
		if !errors.Is(joined, ErrResourceFailed) || !errors.Is(joined, ErrTracerFailed) {
			t.Fatalf("sentinel errors are not matched via errors.Is")
		}
	})
}
