package stats

import "math"

// NewNormal creates a validated normal distribution.
func NewNormal(mu, sigma float64) (Normal, error) {
	if sigma <= 0 {
		return Normal{}, ErrInvalidParameter
	}

	return Normal{Mu: mu, Sigma: sigma}, nil
}

// PDF returns the probability density at x for the normal distribution.
func (n Normal) PDF(x float64) float64 {
	z := (x - n.Mu) / n.Sigma
	return math.Exp(-0.5*z*z) / (n.Sigma * math.Sqrt(2*math.Pi))
}

// CDF returns the cumulative probability at x for the normal distribution.
func (n Normal) CDF(x float64) float64 {
	return 0.5 * math.Erfc(-(x-n.Mu)/(n.Sigma*math.Sqrt2))
}

// Mean returns the expected value of the normal distribution.
func (n Normal) Mean() float64 {
	return n.Mu
}

// Variance returns the variance of the normal distribution.
func (n Normal) Variance() float64 {
	return n.Sigma * n.Sigma
}

// Quantile returns the inverse CDF at probability p using the rational approximation.
func (n Normal) Quantile(p float64) float64 {
	return n.Mu + n.Sigma*normalQuantile(p)
}

// normalQuantile computes the quantile of the standard normal distribution.
// Uses the rational approximation by Peter Acklam.
func normalQuantile(p float64) float64 {
	const (
		a1 = -3.969683028665376e+01
		a2 = 2.209460984245205e+02
		a3 = -2.759285104469687e+02
		a4 = 1.383577518672690e+02
		a5 = -3.066479806614716e+01
		a6 = 2.506628277459239e+00

		b1 = -5.447609879822406e+01
		b2 = 1.615858368580409e+02
		b3 = -1.556989798598866e+02
		b4 = 6.680131188771972e+01
		b5 = -1.328068155288572e+01

		c1 = -7.784894002430293e-03
		c2 = -3.223964580411365e-01
		c3 = -2.400758277161838e+00
		c4 = -2.549732539343734e+00
		c5 = 4.374664141464968e+00
		c6 = 2.938163982698783e+00

		d1 = 7.784695709041462e-03
		d2 = 3.224671290700398e-01
		d3 = 2.445134137142996e+00
		d4 = 3.754408661907416e+00

		pLow  = 0.02425
		pHigh = 1 - pLow
	)

	if p < pLow {
		q := math.Sqrt(-2 * math.Log(p))

		return (((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6) /
			((((d1*q+d2)*q+d3)*q+d4)*q + 1)
	}

	if p <= pHigh {
		q := p - 0.5
		r := q * q

		return (((((a1*r+a2)*r+a3)*r+a4)*r+a5)*r + a6) * q /
			(((((b1*r+b2)*r+b3)*r+b4)*r+b5)*r + 1)
	}

	q := math.Sqrt(-2 * math.Log(1-p))

	return -(((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6) /
		((((d1*q+d2)*q+d3)*q+d4)*q + 1)
}
