package stats

import "math"

// NewGamma creates a validated Gamma distribution.
func NewGamma(alpha, beta float64) (Gamma, error) {
	if alpha <= 0 || beta <= 0 {
		return Gamma{}, ErrInvalidParameter
	}

	return Gamma{Alpha: alpha, Bet: beta}, nil
}

// PDF returns the probability density at x for the Gamma distribution.
func (g Gamma) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	if x == 0 {
		if g.Alpha < 1 {
			return math.Inf(1)
		}

		if g.Alpha == 1 {
			return g.Bet
		}

		return 0
	}

	return math.Exp((g.Alpha-1)*math.Log(x) - g.Bet*x + g.Alpha*math.Log(g.Bet) - lnGamma(g.Alpha))
}

// CDF returns the cumulative probability at x for the Gamma distribution.
func (g Gamma) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	return incompleteGamma(g.Alpha, g.Bet*x)
}

// Mean returns the expected value of the Gamma distribution.
func (g Gamma) Mean() float64 {
	return g.Alpha / g.Bet
}

// Variance returns the variance of the Gamma distribution.
func (g Gamma) Variance() float64 {
	return g.Alpha / (g.Bet * g.Bet)
}

// Quantile returns the inverse CDF at probability p using bisection.
func (g Gamma) Quantile(p float64) float64 {
	return bisectQuantile(0, g.Mean()*10, p, g.CDF, true)
}
