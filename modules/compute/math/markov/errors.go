package markov

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for markov errors.
const (
	MarkovType = "math-markov"
)

var _ error = (*Error)(nil)

// Error is the domain error for markov operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for markov operations.
var (
	ErrStateNotFound      = errors.New("state not found")
	ErrDuplicateState     = errors.New("duplicate state")
	ErrEmptyChain         = errors.New("chain must have at least one state")
	ErrInvalidMatrix      = errors.New("invalid transition matrix dimensions")
	ErrInvalidProbability = errors.New("probability must be non-negative")
	ErrInvalidRow         = errors.New("row probabilities must sum to 1")
	ErrNotIrreducible     = errors.New("chain is not irreducible")
	ErrSingularMatrix     = errors.New("singular matrix")
	ErrNotTransient       = errors.New("state is not transient")
	ErrNoAbsorbingStates  = errors.New("no absorbing states")
	ErrMarkovFailed       = errors.New("markov operation failed")
)

// ErrMarkov creates a markov domain error joining the given causes.
func ErrMarkov(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MarkovType,
			Err:  errors.Join(append(errs, ErrMarkovFailed)...),
		},
	}
}
