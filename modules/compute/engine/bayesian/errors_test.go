package bayesian

import (
	"errors"
	"testing"
)

func TestError_implements_error(t *testing.T) {
	t.Parallel()

	err := ErrQuery(ErrQueryNotInNetwork)
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestErrQuery(t *testing.T) {
	t.Parallel()

	err := ErrQuery(ErrQueryNotInNetwork)
	if !errors.Is(err, ErrQueryNotInNetwork) {
		t.Fatal("expected ErrQueryNotInNetwork")
	}
}

func TestErrQuery_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("variable X missing")
	err := ErrQuery(cause, ErrQueryNotInNetwork)

	if !errors.Is(err, ErrQueryNotInNetwork) {
		t.Fatal("expected ErrQueryNotInNetwork")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	err := ErrValidation(ErrNetworkInvalid)
	if !errors.Is(err, ErrNetworkInvalid) {
		t.Fatal("expected ErrNetworkInvalid")
	}
}

func TestErrValidation_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("CPT invalid")
	err := ErrValidation(cause, ErrNetworkInvalid)

	if !errors.Is(err, ErrNetworkInvalid) {
		t.Fatal("expected ErrNetworkInvalid")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrNetworkInvalid", func(t *testing.T) {
		t.Parallel()

		if ErrNetworkInvalid.Error() != "network is invalid" {
			t.Fatalf("unexpected: %s", ErrNetworkInvalid.Error())
		}
	})

	t.Run("ErrCyclicNetwork", func(t *testing.T) {
		t.Parallel()

		if ErrCyclicNetwork.Error() != "network contains a cycle" {
			t.Fatalf("unexpected: %s", ErrCyclicNetwork.Error())
		}
	})

	t.Run("ErrQueryNotInNetwork", func(t *testing.T) {
		t.Parallel()

		if ErrQueryNotInNetwork.Error() != "query variable not in network" {
			t.Fatalf("unexpected: %s", ErrQueryNotInNetwork.Error())
		}
	})
}

func TestError_type(t *testing.T) {
	t.Parallel()

	err := ErrQuery(ErrQueryNotInNetwork)

	var bayesErr *Error

	if !errors.As(err, &bayesErr) {
		t.Fatal("expected *Error type")
	}

	if bayesErr.Type != BayesianType {
		t.Fatalf("expected type %s, got %s", BayesianType, bayesErr.Type)
	}
}

func TestErrQuery_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrQuery()
	if !errors.Is(err, ErrBayesianQueryFailed) {
		t.Fatal("expected ErrBayesianQueryFailed in chain")
	}
}

func TestErrValidation_zeroArgs(t *testing.T) {
	t.Parallel()

	err := ErrValidation()
	if !errors.Is(err, ErrBayesianValidationFailed) {
		t.Fatal("expected ErrBayesianValidationFailed in chain")
	}
}

func TestErrBayesianQueryFailed(t *testing.T) {
	t.Parallel()

	if ErrBayesianQueryFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrBayesianQueryFailed.Error() != "bayesian query failed" {
		t.Fatalf("unexpected: %s", ErrBayesianQueryFailed.Error())
	}
}

func TestErrBayesianValidationFailed(t *testing.T) {
	t.Parallel()

	if ErrBayesianValidationFailed == nil {
		t.Fatal("expected non-nil error")
	}

	if ErrBayesianValidationFailed.Error() != "bayesian validation failed" {
		t.Fatalf("unexpected: %s", ErrBayesianValidationFailed.Error())
	}
}
