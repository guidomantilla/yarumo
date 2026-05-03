package stats

import "math"

// NewGumbel creates a validated Gumbel (type I extreme value) distribution.
// Mu is the location parameter and Beta is the scale parameter.
func NewGumbel(mu, beta float64) (Gumbel, error) {
	if beta <= 0 {
		return Gumbel{}, ErrInvalidParameter
	}

	return Gumbel{Mu: mu, Beta: beta}, nil
}

// PDF returns the probability density at x for the Gumbel distribution.
func (g Gumbel) PDF(x float64) float64 {
	z := (x - g.Mu) / g.Beta

	return math.Exp(-(z + math.Exp(-z))) / g.Beta
}

// CDF returns the cumulative probability at x for the Gumbel distribution.
func (g Gumbel) CDF(x float64) float64 {
	z := (x - g.Mu) / g.Beta

	return math.Exp(-math.Exp(-z))
}

// Mean returns the expected value of the Gumbel distribution.
func (g Gumbel) Mean() float64 {
	return g.Mu + g.Beta*eulerMascheroni
}

// Variance returns the variance of the Gumbel distribution.
func (g Gumbel) Variance() float64 {
	return math.Pi * math.Pi * g.Beta * g.Beta / 6
}

// Quantile returns the inverse CDF at probability p for the Gumbel distribution.
func (g Gumbel) Quantile(p float64) float64 {
	return g.Mu - g.Beta*math.Log(-math.Log(p))
}

// eulerMascheroni is the Euler-Mascheroni constant γ ≈ 0.5772.
const eulerMascheroni = 0.5772156649015329
