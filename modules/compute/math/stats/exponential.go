package stats

import "math"

// NewExponential creates a validated exponential distribution.
func NewExponential(lambda float64) (Exponential, error) {
	if lambda <= 0 {
		return Exponential{}, ErrInvalidParameter
	}

	return Exponential{Lambda: lambda}, nil
}

// PDF returns the probability density at x for the exponential distribution.
func (e Exponential) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	return e.Lambda * math.Exp(-e.Lambda*x)
}

// CDF returns the cumulative probability at x for the exponential distribution.
func (e Exponential) CDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	return 1 - math.Exp(-e.Lambda*x)
}

// Mean returns the expected value of the exponential distribution.
func (e Exponential) Mean() float64 {
	return 1 / e.Lambda
}

// Variance returns the variance of the exponential distribution.
func (e Exponential) Variance() float64 {
	return 1 / (e.Lambda * e.Lambda)
}

// Quantile returns the inverse CDF at probability p.
func (e Exponential) Quantile(p float64) float64 {
	return -math.Log(1-p) / e.Lambda
}
