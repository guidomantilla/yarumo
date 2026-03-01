package probability

import "errors"

// Error sentinels for the probability package.
var (
	ErrInvalidProb      = errors.New("probability must be in [0,1]")
	ErrNotNormalized    = errors.New("distribution does not sum to 1")
	ErrEmptyDist        = errors.New("distribution is empty")
	ErrOutcomeNotFound  = errors.New("outcome not found in distribution")
	ErrVariableNotFound = errors.New("variable not found")
)
