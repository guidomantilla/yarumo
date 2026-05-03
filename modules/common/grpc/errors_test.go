package grpc

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error message with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrServer(errors.New("root-cause"))
		if !strings.Contains(err.Error(), "grpc server grpc-server error: root-cause") {
			t.Fatalf("unexpected error string: %s", err.Error())
		}
	})

	t.Run("wraps with ErrGrpcServerFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrServer(errors.New("inner"))
		if !errors.Is(err, ErrGrpcServerFailed) {
			t.Fatal("expected error to wrap ErrGrpcServerFailed")
		}
	})

	t.Run("wraps the original cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("original-cause")

		err := ErrServer(cause)
		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap original cause")
		}
	})
}

func TestErrServer(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrServer(errors.New("fail"))

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("wraps multiple errors", func(t *testing.T) {
		t.Parallel()

		cause1 := errors.New("cause1")
		cause2 := errors.New("cause2")

		err := ErrServer(cause1, cause2)

		if !errors.Is(err, cause1) {
			t.Fatal("expected error to wrap cause1")
		}

		if !errors.Is(err, cause2) {
			t.Fatal("expected error to wrap cause2")
		}

		if !errors.Is(err, ErrGrpcServerFailed) {
			t.Fatal("expected error to wrap ErrGrpcServerFailed")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrGrpcServerFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrGrpcServerFailed == nil {
			t.Fatal("ErrGrpcServerFailed should not be nil")
		}

		if ErrGrpcServerFailed.Error() == "" {
			t.Fatal("ErrGrpcServerFailed message should not be empty")
		}
	})
}
