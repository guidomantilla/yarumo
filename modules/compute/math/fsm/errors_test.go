package fsm

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

func TestErrFSM_wraps_sentinels(t *testing.T) {
	t.Parallel()

	err := ErrFSM(ErrStateNotFound)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatal("expected error to wrap ErrStateNotFound")
	}
}

func TestErrFSM_typed_error(t *testing.T) {
	t.Parallel()

	err := ErrFSM(ErrGuardRejected)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}

	if typed.Type != FSMType {
		t.Fatalf("expected type %q, got %q", FSMType, typed.Type)
	}
}

func TestErrFSM_multiple_causes(t *testing.T) {
	t.Parallel()

	err := ErrFSM(ErrInvalidTransition, ErrStateNotFound)

	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatal("expected error to wrap ErrInvalidTransition")
	}

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatal("expected error to wrap ErrStateNotFound")
	}
}

func TestErrFSM_error_message(t *testing.T) {
	t.Parallel()

	err := ErrFSM(ErrStateNotFound)

	got := err.Error()
	if !strings.Contains(got, "math-fsm") {
		t.Fatalf("expected type prefix in %q", got)
	}
	if !strings.Contains(got, "state not found") {
		t.Fatalf("expected cause in %q", got)
	}
}

func TestErrFSM_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrFSM()
	if !errors.Is(err, ErrFSMFailed) {
		t.Fatal("expected ErrFSMFailed in chain")
	}
}

func TestErrFSMFailed(t *testing.T) {
	t.Parallel()

	if ErrFSMFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrFSMFailed.Error() != "fsm operation failed" {
		t.Fatalf("unexpected message: %s", ErrFSMFailed.Error())
	}
}

func TestError_type_compliance(t *testing.T) {
	t.Parallel()

	var _ error = &Error{
		TypedError: cerrs.TypedError{
			Type: FSMType,
			Err:  ErrStateNotFound,
		},
	}
}

func TestFSMType_constant(t *testing.T) {
	t.Parallel()

	if FSMType != "math-fsm" {
		t.Fatalf("expected %q, got %q", "math-fsm", FSMType)
	}
}

func TestErrFSM_all_sentinels(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrStateNotFound, ErrTransitionNotFound, ErrDuplicateState,
		ErrDuplicateTransition, ErrGuardRejected, ErrInvalidTransition,
		ErrNoInitialState, ErrInvalidEvent,
	}

	for _, s := range sentinels {
		err := ErrFSM(s)

		if !errors.Is(err, s) {
			t.Fatalf("expected error to wrap %v", s)
		}
	}
}
