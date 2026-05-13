package temporal

import (
	"errors"
	"testing"
)

func TestErrEventNotFound(t *testing.T) {
	t.Parallel()

	got := ErrEventNotFound.Error()
	if got != "event not found in trace" {
		t.Fatalf("expected %q, got %q", "event not found in trace", got)
	}
}

func TestErrTemporal(t *testing.T) {
	t.Parallel()

	err := ErrTemporal(ErrEventNotFound)
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatal("expected ErrEventNotFound")
	}
}

func TestErrTemporal_type(t *testing.T) {
	t.Parallel()

	err := ErrTemporal(ErrEventNotFound)

	var tempErr *Error

	if !errors.As(err, &tempErr) {
		t.Fatal("expected *Error type")
	}

	if tempErr.Type != TemporalType {
		t.Fatalf("expected type %s, got %s", TemporalType, tempErr.Type)
	}
}

func TestErrTemporal_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrTemporal()
	if !errors.Is(err, ErrTemporalFailed) {
		t.Fatal("expected ErrTemporalFailed in chain")
	}
}

func TestErrTemporalFailed(t *testing.T) {
	t.Parallel()

	if ErrTemporalFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrTemporalFailed.Error() != "temporal operation failed" {
		t.Fatalf("unexpected message: %s", ErrTemporalFailed.Error())
	}
}
