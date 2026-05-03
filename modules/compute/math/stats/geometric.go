package stats

import "math"

// NewGeometric creates a validated geometric distribution.
// P is the probability of success on each trial.
func NewGeometric(p float64) (Geometric, error) {
	if p <= 0 || p > 1 {
		return Geometric{}, ErrInvalidProb
	}

	return Geometric{P: p}, nil
}

// PMF returns the probability mass at k for the geometric distribution.
// Uses the convention P(X=k) = (1-p)^k * p for k = 0, 1, 2, ...
// where k is the number of failures before the first success.
func (g Geometric) PMF(k int) float64 {
	if k < 0 {
		return 0
	}

	return math.Pow(1-g.P, float64(k)) * g.P
}

// CDFDiscrete returns the cumulative probability at k for the geometric distribution.
func (g Geometric) CDFDiscrete(k int) float64 {
	if k < 0 {
		return 0
	}

	return 1 - math.Pow(1-g.P, float64(k+1))
}

// Mean returns the expected value of the geometric distribution.
func (g Geometric) Mean() float64 {
	return (1 - g.P) / g.P
}

// Variance returns the variance of the geometric distribution.
func (g Geometric) Variance() float64 {
	return (1 - g.P) / (g.P * g.P)
}
