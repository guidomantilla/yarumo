package graph

import (
	"errors"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

func TestErrGraph_wraps_sentinels(t *testing.T) {
	t.Parallel()

	err := ErrGraph(ErrNodeNotFound)
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatal("expected error to wrap ErrNodeNotFound")
	}
}

func TestErrGraph_typed_error(t *testing.T) {
	t.Parallel()

	err := ErrGraph(ErrCycleDetected)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}

	if typed.Type != GraphType {
		t.Fatalf("expected type %q, got %q", GraphType, typed.Type)
	}
}

func TestErrGraph_multiple_causes(t *testing.T) {
	t.Parallel()

	err := ErrGraph(ErrInvalidEdge, ErrNodeNotFound)

	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatal("expected error to wrap ErrInvalidEdge")
	}

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatal("expected error to wrap ErrNodeNotFound")
	}
}

func TestErrGraph_error_message(t *testing.T) {
	t.Parallel()

	err := ErrGraph(ErrNodeNotFound)

	expected := "math-graph error: node not found"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestError_type_compliance(t *testing.T) {
	t.Parallel()

	var _ error = &Error{
		TypedError: cerrs.TypedError{
			Type: GraphType,
			Err:  ErrNodeNotFound,
		},
	}
}

func TestGraphType_constant(t *testing.T) {
	t.Parallel()

	if GraphType != "math-graph" {
		t.Fatalf("expected %q, got %q", "math-graph", GraphType)
	}
}

func TestErrGraph_all_sentinels(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		ErrNodeNotFound, ErrEdgeNotFound, ErrCycleDetected,
		ErrInvalidEdge, ErrNotDAG, ErrNotTree,
		ErrNegativeCycle, ErrDisconnected, ErrNoPath,
		ErrNotBipartite, ErrMultipleParents, ErrSelfLoop,
		ErrDuplicateNode,
	}

	for _, s := range sentinels {
		err := ErrGraph(s)

		if !errors.Is(err, s) {
			t.Fatalf("expected error to wrap %v", s)
		}
	}
}
