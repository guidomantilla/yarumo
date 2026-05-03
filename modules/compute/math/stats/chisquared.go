package stats

import "math"

// NewChiSquared creates a validated chi-squared distribution.
func NewChiSquared(k float64) (ChiSquared, error) {
	if k <= 0 {
		return ChiSquared{}, ErrInvalidDegreesOfFreedom
	}

	return ChiSquared{K: k}, nil
}

// PDF returns the probability density at x for the chi-squared distribution.
func (c ChiSquared) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	halfK := c.K / 2

	if x == 0 {
		if halfK < 1 {
			return math.Inf(1)
		}

		if halfK == 1 {
			return math.Exp(-halfK*math.Log(2) - lnGamma(halfK))
		}

		return 0
	}

	return math.Exp((halfK-1)*math.Log(x) - x/2 - halfK*math.Log(2) - lnGamma(halfK))
}

// CDF returns the cumulative probability at x for the chi-squared distribution.
// Delegates to the regularized lower incomplete gamma function P(k/2, x/2).
func (c ChiSquared) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	return incompleteGamma(c.K/2, x/2)
}

// Mean returns the expected value of the chi-squared distribution.
func (c ChiSquared) Mean() float64 {
	return c.K
}

// Variance returns the variance of the chi-squared distribution.
func (c ChiSquared) Variance() float64 {
	return 2 * c.K
}

// Quantile returns the inverse CDF at probability p using bisection.
func (c ChiSquared) Quantile(p float64) float64 {
	hi := c.K * 10
	if hi < 10 {
		hi = 10
	}

	return bisectQuantile(0, hi, p, c.CDF, true)
}
