package stats

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrEmptyData", func(t *testing.T) {
		t.Parallel()

		if ErrEmptyData.Error() != "data set is empty" {
			t.Fatalf("unexpected: %s", ErrEmptyData.Error())
		}
	})

	t.Run("ErrInvalidPercentile", func(t *testing.T) {
		t.Parallel()

		if ErrInvalidPercentile.Error() != "percentile must be in (0,100]" {
			t.Fatalf("unexpected: %s", ErrInvalidPercentile.Error())
		}
	})

	t.Run("ErrMismatchedLengths", func(t *testing.T) {
		t.Parallel()

		if ErrMismatchedLengths.Error() != "data sets must have the same length" {
			t.Fatalf("unexpected: %s", ErrMismatchedLengths.Error())
		}
	})

	t.Run("ErrInsufficientData", func(t *testing.T) {
		t.Parallel()

		if ErrInsufficientData.Error() != "at least two data points required" {
			t.Fatalf("unexpected: %s", ErrInsufficientData.Error())
		}
	})

	t.Run("ErrZeroVariance", func(t *testing.T) {
		t.Parallel()

		if ErrZeroVariance.Error() != "variance is zero" {
			t.Fatalf("unexpected: %s", ErrZeroVariance.Error())
		}
	})

	t.Run("ErrInvalidWeights", func(t *testing.T) {
		t.Parallel()

		if ErrInvalidWeights.Error() != "weights must sum to a positive value" {
			t.Fatalf("unexpected: %s", ErrInvalidWeights.Error())
		}
	})
}

func TestErrInvalidProb(t *testing.T) {
	t.Parallel()

	if ErrInvalidProb == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrInvalidProb.Error() != "probability must be in [0,1]" {
		t.Fatalf("unexpected message: %s", ErrInvalidProb.Error())
	}
}

func TestErrNotNormalized(t *testing.T) {
	t.Parallel()

	if ErrNotNormalized == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrNotNormalized.Error() != "distribution does not sum to 1" {
		t.Fatalf("unexpected message: %s", ErrNotNormalized.Error())
	}
}

func TestErrEmptyDist(t *testing.T) {
	t.Parallel()

	if ErrEmptyDist == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrEmptyDist.Error() != "distribution is empty" {
		t.Fatalf("unexpected message: %s", ErrEmptyDist.Error())
	}
}

func TestErrOutcomeNotFound(t *testing.T) {
	t.Parallel()

	if ErrOutcomeNotFound == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrOutcomeNotFound.Error() != "outcome not found in distribution" {
		t.Fatalf("unexpected message: %s", ErrOutcomeNotFound.Error())
	}
}

func TestErrZeroEvidence(t *testing.T) {
	t.Parallel()

	if ErrZeroEvidence == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrZeroEvidence.Error() != "evidence probability is zero" {
		t.Fatalf("unexpected message: %s", ErrZeroEvidence.Error())
	}
}

func TestErrInvalidParameter(t *testing.T) {
	t.Parallel()

	if ErrInvalidParameter.Error() != "invalid distribution parameter" {
		t.Fatalf("unexpected: %s", ErrInvalidParameter.Error())
	}
}

func TestErrOutOfRange(t *testing.T) {
	t.Parallel()

	if ErrOutOfRange.Error() != "value is out of range" {
		t.Fatalf("unexpected: %s", ErrOutOfRange.Error())
	}
}

func TestErrStats_wrapsSentinel(t *testing.T) {
	t.Parallel()

	err := ErrStats(ErrEmptyData)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatal("expected ErrEmptyData")
	}
}

func TestErrStats_wrapsMultipleSentinels(t *testing.T) {
	t.Parallel()

	err := ErrStats(ErrEmptyData, ErrInvalidParameter)

	if !errors.Is(err, ErrEmptyData) {
		t.Fatal("expected ErrEmptyData")
	}

	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatal("expected ErrInvalidParameter")
	}
}

func TestErrStats_wrapsAdditionalCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("underlying issue")
	err := ErrStats(ErrInsufficientData, cause)

	if !errors.Is(err, ErrInsufficientData) {
		t.Fatal("expected ErrInsufficientData")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestErrStats_isErrorType(t *testing.T) {
	t.Parallel()

	err := ErrStats(ErrEmptyData)

	var statsErr *Error

	if !errors.As(err, &statsErr) {
		t.Fatal("expected *Error type")
	}

	if statsErr.Type != StatsType {
		t.Fatalf("expected type %s, got %s", StatsType, statsErr.Type)
	}
}
