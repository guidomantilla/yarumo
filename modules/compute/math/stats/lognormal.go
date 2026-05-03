package stats

import "math"

// NewLognormal creates a validated log-normal distribution.
func NewLognormal(mu, sigma float64) (Lognormal, error) {
	if sigma <= 0 {
		return Lognormal{}, ErrInvalidParameter
	}

	return Lognormal{Mu: mu, Sigma: sigma}, nil
}

// PDF returns the probability density at x for the log-normal distribution.
func (l Lognormal) PDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	z := (math.Log(x) - l.Mu) / l.Sigma

	return math.Exp(-0.5*z*z) / (x * l.Sigma * math.Sqrt(2*math.Pi))
}

// CDF returns the cumulative probability at x for the log-normal distribution.
// Delegates to the normal CDF applied to log(x).
func (l Lognormal) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	return 0.5 * math.Erfc(-(math.Log(x)-l.Mu)/(l.Sigma*math.Sqrt2))
}

// Mean returns the expected value of the log-normal distribution.
func (l Lognormal) Mean() float64 {
	return math.Exp(l.Mu + l.Sigma*l.Sigma/2)
}

// Variance returns the variance of the log-normal distribution.
func (l Lognormal) Variance() float64 {
	s2 := l.Sigma * l.Sigma

	return (math.Exp(s2) - 1) * math.Exp(2*l.Mu+s2)
}

// Quantile returns the inverse CDF at probability p for the log-normal distribution.
func (l Lognormal) Quantile(p float64) float64 {
	return math.Exp(l.Mu + l.Sigma*normalQuantile(p))
}
