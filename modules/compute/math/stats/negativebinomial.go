package stats

import "math"

// NewNegativeBinomial creates a validated negative binomial distribution.
// R is the number of successes required, P is the probability of success on each trial.
func NewNegativeBinomial(r int, p float64) (NegativeBinomial, error) {
	if r <= 0 {
		return NegativeBinomial{}, ErrInvalidParameter
	}

	if p <= 0 || p > 1 {
		return NegativeBinomial{}, ErrInvalidProb
	}

	return NegativeBinomial{R: r, P: p}, nil
}

// PMF returns the probability mass at k for the negative binomial distribution.
// k is the number of failures before the r-th success.
func (nb NegativeBinomial) PMF(k int) float64 {
	if k < 0 {
		return 0
	}

	// P(X=k) = C(k+r-1, k) * p^r * (1-p)^k.
	lnCoeff := lnBinom(k+nb.R-1, k)
	lnPr := float64(nb.R) * math.Log(nb.P)

	if k == 0 {
		return math.Exp(lnCoeff + lnPr)
	}

	return math.Exp(lnCoeff + lnPr + float64(k)*math.Log(1-nb.P))
}

// CDFDiscrete returns the cumulative probability at k for the negative binomial distribution.
func (nb NegativeBinomial) CDFDiscrete(k int) float64 {
	if k < 0 {
		return 0
	}

	sum := 0.0

	for i := range k + 1 {
		sum += nb.PMF(i)
	}

	return sum
}

// Mean returns the expected value of the negative binomial distribution.
func (nb NegativeBinomial) Mean() float64 {
	return float64(nb.R) * (1 - nb.P) / nb.P
}

// Variance returns the variance of the negative binomial distribution.
func (nb NegativeBinomial) Variance() float64 {
	return float64(nb.R) * (1 - nb.P) / (nb.P * nb.P)
}
