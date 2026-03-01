package bayesian

import (
	"errors"
	"testing"
)

func TestError_implements_error(t *testing.T) {
	t.Parallel()

	err := ErrQuery()
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestErrQuery(t *testing.T) {
	t.Parallel()

	err := ErrQuery()
	if !errors.Is(err, ErrQueryNotInNetwork) {
		t.Fatal("expected ErrQueryNotInNetwork")
	}
}

func TestErrQuery_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("variable X missing")
	err := ErrQuery(cause)

	if !errors.Is(err, ErrQueryNotInNetwork) {
		t.Fatal("expected ErrQueryNotInNetwork")
	}

	if !errors.Is(err, cause) {
		t.Fatal("expected cause error")
	}
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	err := ErrValidation()
	if !errors.Is(err, ErrNetworkInvalid) {
		t.Fatal("expected ErrNetworkInvalid")
	}
}

func TestErrValidation_withCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("CPT invalid")
	err := ErrValidation(cause)

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

	t.Run("ErrNoEvidence", func(t *testing.T) {
		t.Parallel()

		if ErrNoEvidence.Error() != "no evidence provided" {
			t.Fatalf("unexpected: %s", ErrNoEvidence.Error())
		}
	})
}

func TestError_type(t *testing.T) {
	t.Parallel()

	err := ErrQuery()

	var bayesErr *Error

	if !errors.As(err, &bayesErr) {
		t.Fatal("expected *Error type")
	}

	if bayesErr.Type != BayesianType {
		t.Fatalf("expected type %s, got %s", BayesianType, bayesErr.Type)
	}
}
