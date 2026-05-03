package stats

import "math"

// NewFDist creates a validated F-distribution.
func NewFDist(d1, d2 float64) (FDist, error) {
	if d1 <= 0 || d2 <= 0 {
		return FDist{}, ErrInvalidDegreesOfFreedom
	}

	return FDist{D1: d1, D2: d2}, nil
}

// PDF returns the probability density at x for the F-distribution.
func (f FDist) PDF(x float64) float64 {
	if x < 0 {
		return 0
	}

	d1 := f.D1
	d2 := f.D2
	halfD1 := d1 / 2

	if x == 0 {
		if halfD1 < 1 {
			return math.Inf(1)
		}

		if halfD1 == 1 {
			return math.Exp(math.Log(2) - math.Log(d2) - lnBeta(1, d2/2))
		}

		return 0
	}

	lnNum := (d1/2)*math.Log(d1) + (d2/2)*math.Log(d2) + ((d1/2)-1)*math.Log(x)
	lnDen := ((d1 + d2) / 2) * math.Log(d1*x+d2)
	lnB := lnBeta(d1/2, d2/2)

	return math.Exp(lnNum - lnDen - lnB)
}

// CDF returns the cumulative probability at x for the F-distribution.
// Uses the regularized incomplete beta function.
func (f FDist) CDF(x float64) float64 {
	if x <= 0 {
		return 0
	}

	d1 := f.D1
	d2 := f.D2
	z := d1 * x / (d1*x + d2)

	return regularizedIncompleteBeta(z, d1/2, d2/2)
}

// Mean returns the expected value of the F-distribution.
// Defined only for d2 > 2.
func (f FDist) Mean() float64 {
	if f.D2 > 2 {
		return f.D2 / (f.D2 - 2)
	}

	return math.NaN()
}

// Variance returns the variance of the F-distribution.
// Defined only for d2 > 4.
func (f FDist) Variance() float64 {
	if f.D2 > 4 {
		d1 := f.D1
		d2 := f.D2

		return 2 * d2 * d2 * (d1 + d2 - 2) / (d1 * (d2 - 2) * (d2 - 2) * (d2 - 4))
	}

	return math.NaN()
}

// Quantile returns the inverse CDF at probability p using bisection.
func (f FDist) Quantile(p float64) float64 {
	return bisectQuantile(0, 100, p, f.CDF, true)
}
