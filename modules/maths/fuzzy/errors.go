package fuzzy

import "errors"

// Error sentinels for the fuzzy package.
var (
	ErrInvalidDegree = errors.New("degree must be in [0,1]")
	ErrEmptySamples  = errors.New("empty sample set")
	ErrInvalidRange  = errors.New("invalid range")
)
