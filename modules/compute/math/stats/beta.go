package stats

import "math"

// NewBeta creates a validated Beta distribution.
func NewBeta(alpha, beta float64) (Beta, error) {
	if alpha <= 0 || beta <= 0 {
		return Beta{}, ErrInvalidParameter
	}

	return Beta{Alpha: alpha, Bet: beta}, nil
}

// PDF returns the probability density at x for the Beta distribution.
func (b Beta) PDF(x float64) float64 {
	if x < 0 || x > 1 {
		return 0
	}

	if x == 0 {
		if b.Alpha < 1 {
			return math.Inf(1)
		}

		if b.Alpha == 1 {
			return 1 / math.Exp(lnGamma(b.Alpha)+lnGamma(b.Bet)-lnGamma(b.Alpha+b.Bet))
		}

		return 0
	}

	if x == 1 {
		if b.Bet < 1 {
			return math.Inf(1)
		}

		if b.Bet == 1 {
			return 1 / math.Exp(lnGamma(b.Alpha)+lnGamma(b.Bet)-lnGamma(b.Alpha+b.Bet))
		}

		return 0
	}

	logB := lnGamma(b.Alpha) + lnGamma(b.Bet) - lnGamma(b.Alpha+b.Bet)

	return math.Exp((b.Alpha-1)*math.Log(x) + (b.Bet-1)*math.Log(1-x) - logB)
}

// CDF returns the cumulative probability at x for the Beta distribution.
func (b Beta) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	if x >= 1 {
		return 1
	}

	return regularizedIncompleteBeta(x, b.Alpha, b.Bet)
}

// Mean returns the expected value of the Beta distribution.
func (b Beta) Mean() float64 {
	return b.Alpha / (b.Alpha + b.Bet)
}

// Variance returns the variance of the Beta distribution.
func (b Beta) Variance() float64 {
	sum := b.Alpha + b.Bet
	return (b.Alpha * b.Bet) / (sum * sum * (sum + 1))
}

// Quantile returns the inverse CDF at probability p using bisection.
func (b Beta) Quantile(p float64) float64 {
	return bisectQuantile(0, 1, p, b.CDF, false)
}
