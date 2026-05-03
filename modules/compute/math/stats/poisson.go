package stats

import "math"

// NewPoisson creates a validated Poisson distribution.
func NewPoisson(lambda float64) (Poisson, error) {
	if lambda <= 0 {
		return Poisson{}, ErrInvalidParameter
	}

	return Poisson{Lambda: lambda}, nil
}

// PMF returns the probability mass at k for the Poisson distribution.
func (p Poisson) PMF(k int) float64 {
	if k < 0 {
		return 0
	}

	return math.Exp(float64(k)*math.Log(p.Lambda) - p.Lambda - lnFactorial(k))
}

// CDFDiscrete returns the cumulative probability at k for the Poisson distribution.
func (p Poisson) CDFDiscrete(k int) float64 {
	if k < 0 {
		return 0
	}

	sum := 0.0

	for i := range k + 1 {
		sum += p.PMF(i)
	}

	return sum
}

// Mean returns the expected value of the Poisson distribution.
func (p Poisson) Mean() float64 {
	return p.Lambda
}

// Variance returns the variance of the Poisson distribution.
func (p Poisson) Variance() float64 {
	return p.Lambda
}

// lnFactorial returns the natural logarithm of k! using lnGamma(k+1).
func lnFactorial(k int) float64 {
	return lnGamma(float64(k) + 1)
}
