package stats

import "math"

// NewStudentT creates a validated Student's t-distribution.
func NewStudentT(nu float64) (StudentT, error) {
	if nu <= 0 {
		return StudentT{}, ErrInvalidDegreesOfFreedom
	}

	return StudentT{Nu: nu}, nil
}

// PDF returns the probability density at x for the Student's t-distribution.
func (s StudentT) PDF(x float64) float64 {
	coeff := math.Exp(lnGamma((s.Nu+1)/2) - lnGamma(s.Nu/2))
	denom := math.Sqrt(s.Nu*math.Pi) * math.Pow(1+x*x/s.Nu, (s.Nu+1)/2)

	return coeff / denom
}

// CDF returns the cumulative probability at x for the Student's t-distribution.
// Uses the regularized incomplete beta function.
func (s StudentT) CDF(x float64) float64 {
	xt := s.Nu / (s.Nu + x*x)
	ib := regularizedIncompleteBeta(xt, s.Nu/2, 0.5)

	if x >= 0 {
		return 1 - 0.5*ib
	}

	return 0.5 * ib
}

// Mean returns the expected value of the Student's t-distribution.
// Defined only for nu > 1; returns 0 when defined.
func (s StudentT) Mean() float64 {
	if s.Nu > 1 {
		return 0
	}

	return math.NaN()
}

// Variance returns the variance of the Student's t-distribution.
// Defined only for nu > 2.
func (s StudentT) Variance() float64 {
	if s.Nu > 2 {
		return s.Nu / (s.Nu - 2)
	}

	if s.Nu > 1 {
		return math.Inf(1)
	}

	return math.NaN()
}

// Quantile returns the inverse CDF at probability p using bisection.
func (s StudentT) Quantile(p float64) float64 {
	return bisectQuantile(-1000, 1000, p, s.CDF, false)
}
