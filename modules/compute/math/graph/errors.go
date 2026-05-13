package graph

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for graph errors.
const (
	GraphType = "math-graph"
)

var _ error = (*Error)(nil)

// Error is the domain error for graph operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for graph operations.
var (
	ErrNodeNotFound    = errors.New("node not found")
	ErrEdgeNotFound    = errors.New("edge not found")
	ErrCycleDetected   = errors.New("cycle detected")
	ErrInvalidEdge     = errors.New("invalid edge")
	ErrNotDAG          = errors.New("graph is not a DAG")
	ErrNotTree         = errors.New("graph is not a tree")
	ErrNegativeCycle   = errors.New("negative cycle detected")
	ErrDisconnected    = errors.New("graph is disconnected")
	ErrNoPath          = errors.New("no path exists")
	ErrNotBipartite    = errors.New("graph is not bipartite")
	ErrMultipleParents = errors.New("node has multiple parents")
	ErrSelfLoop        = errors.New("self-loop not allowed")
	ErrDuplicateNode   = errors.New("node already exists")
	ErrGraphFailed     = errors.New("graph operation failed")
)

// ErrGraph creates a graph domain error joining the given causes.
func ErrGraph(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: GraphType,
			Err:  errors.Join(append(errs, ErrGraphFailed)...),
		},
	}
}
