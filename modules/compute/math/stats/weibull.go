package stats

import "math"

// NewWeibull creates a validated Weibull distribution.
func NewWeibull(k, lambda float64) (Weibull, error) {
	if k <= 0 || lambda <= 0 {
		return Weibull{}, ErrInvalidParameter
	}

	return Weibull{K: k, Lambda: lambda}, nil
}

// PDF returns the probability density at x for the Weibull distribution.
func (w Weibull) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	if x == 0 {
		if w.K < 1 {
			return math.Inf(1)
		}

		if w.K == 1 {
			return w.K / w.Lambda
		}

		return 0
	}

	ratio := x / w.Lambda

	return (w.K / w.Lambda) * math.Pow(ratio, w.K-1) * math.Exp(-math.Pow(ratio, w.K))
}

// CDF returns the cumulative probability at x for the Weibull distribution.
func (w Weibull) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	return 1 - math.Exp(-math.Pow(x/w.Lambda, w.K))
}

// Mean returns the expected value of the Weibull distribution.
func (w Weibull) Mean() float64 {
	return w.Lambda * math.Gamma(1+1/w.K)
}

// Variance returns the variance of the Weibull distribution.
func (w Weibull) Variance() float64 {
	g1 := math.Gamma(1 + 1/w.K)
	g2 := math.Gamma(1 + 2/w.K)

	return w.Lambda * w.Lambda * (g2 - g1*g1)
}

// Quantile returns the inverse CDF at probability p for the Weibull distribution.
func (w Weibull) Quantile(p float64) float64 {
	return w.Lambda * math.Pow(-math.Log(1-p), 1/w.K)
}
