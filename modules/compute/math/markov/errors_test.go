package markov

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrMarkov_wraps_sentinels(t *testing.T) {
	t.Parallel()

	err := ErrMarkov(ErrStateNotFound)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatal("expected error to wrap ErrStateNotFound")
	}
}

func TestErrMarkov_typed_error(t *testing.T) {
	t.Parallel()

	err := ErrMarkov(ErrSingularMatrix)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}

	if typed.Type != MarkovType {
		t.Fatalf("expected type %q, got %q", MarkovType, typed.Type)
	}
}

func TestErrMarkov_multiple_causes(t *testing.T) {
	t.Parallel()

	err := ErrMarkov(ErrInvalidMatrix, ErrStateNotFound)

	if !errors.Is(err, ErrInvalidMatrix) {
		t.Fatal("expected error to wrap ErrInvalidMatrix")
	}

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatal("expected error to wrap ErrStateNotFound")
	}
}

func TestErrMarkov_error_message(t *testing.T) {
	t.Parallel()

	err := ErrMarkov(ErrStateNotFound)

	got := err.Error()
	if !strings.Contains(got, "math-markov") {
		t.Fatalf("expected type prefix in %q", got)
	}
	if !strings.Contains(got, "state not found") {
		t.Fatalf("expected cause in %q", got)
	}
}

func TestErrMarkov_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrMarkov()
	if !errors.Is(err, ErrMarkovFailed) {
		t.Fatal("expected ErrMarkovFailed in chain")
	}
}

func TestErrMarkovFailed(t *testing.T) {
	t.Parallel()

	if ErrMarkovFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrMarkovFailed.Error() != "markov operation failed" {
		t.Fatalf("unexpected message: %s", ErrMarkovFailed.Error())
	}
}

func TestError_type_compliance(t *testing.T) {
	t.Parallel()

	var _ error = &Error{
		TypedError: cerrs.TypedError{
			Type: MarkovType,
			Err:  ErrStateNotFound,
		},
	}
}

func TestMarkovType_constant(t *testing.T) {
	t.Parallel()

	if MarkovType != "math-markov" {
		t.Fatalf("expected %q, got %q", "math-markov", MarkovType)
	}
}

func TestErrMarkov_all_sentinels(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrStateNotFound, ErrDuplicateState, ErrEmptyChain,
		ErrInvalidMatrix, ErrInvalidProbability, ErrInvalidRow,
		ErrNotIrreducible, ErrSingularMatrix, ErrNotTransient,
		ErrNoAbsorbingStates,
	}

	for _, s := range sentinels {
		err := ErrMarkov(s)

		if !errors.Is(err, s) {
			t.Fatalf("expected error to wrap %v", s)
		}
	}
}
