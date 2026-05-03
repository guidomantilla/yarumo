package stats

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for stats errors.
const (
	StatsType = "math-stats"
)

var _ error = (*Error)(nil)

// Error is the domain error for stats operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for descriptive statistics.
var (
	ErrEmptyData         = errors.New("data set is empty")
	ErrInsufficientData  = errors.New("at least two data points required")
	ErrInvalidPercentile = errors.New("percentile must be in (0,100]")
	ErrMismatchedLengths = errors.New("data sets must have the same length")
	ErrZeroVariance      = errors.New("variance is zero")
	ErrInvalidWeights    = errors.New("weights must sum to a positive value")
)

// Error sentinels for discrete probability.
var (
	ErrInvalidProb     = errors.New("probability must be in [0,1]")
	ErrNotNormalized   = errors.New("distribution does not sum to 1")
	ErrEmptyDist       = errors.New("distribution is empty")
	ErrOutcomeNotFound = errors.New("outcome not found in distribution")
	ErrZeroEvidence    = errors.New("evidence probability is zero")
)

// Error sentinels for continuous distributions.
var (
	ErrInvalidParameter        = errors.New("invalid distribution parameter")
	ErrInvalidDegreesOfFreedom = errors.New("degrees of freedom must be positive")
	ErrOutOfRange              = errors.New("value is out of range")
)

// Error sentinels for hypothesis testing.
var (
	ErrZeroExpected       = errors.New("expected frequency is zero")
	ErrInsufficientGroups = errors.New("at least two groups required")
)

// Error sentinels for descriptive statistics (extended).
var (
	ErrNonPositiveData = errors.New("all values must be positive")
)

// ErrStats creates a stats domain error joining the given causes.
func ErrStats(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: StatsType,
			Err:  errors.Join(errs...),
		},
	}
}
