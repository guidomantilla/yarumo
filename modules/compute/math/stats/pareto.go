package stats

import "math"

// NewPareto creates a validated Pareto (type I) distribution.
// Xm is the minimum value (scale) and Alpha is the shape parameter.
func NewPareto(xm, alpha float64) (Pareto, error) {
	if xm <= 0 || alpha <= 0 {
		return Pareto{}, ErrInvalidParameter
	}

	return Pareto{Xm: xm, Alpha: alpha}, nil
}

// PDF returns the probability density at x for the Pareto distribution.
func (p Pareto) PDF(x float64) float64 {
	if x < p.Xm {
		return 0
	}

	return p.Alpha * math.Pow(p.Xm, p.Alpha) / math.Pow(x, p.Alpha+1)
}

// CDF returns the cumulative probability at x for the Pareto distribution.
func (p Pareto) CDF(x float64) float64 {
	if x < p.Xm {
		return 0
	}

	return 1 - math.Pow(p.Xm/x, p.Alpha)
}

// Mean returns the expected value of the Pareto distribution.
// Returns +Inf when alpha <= 1.
func (p Pareto) Mean() float64 {
	if p.Alpha <= 1 {
		return math.Inf(1)
	}

	return p.Alpha * p.Xm / (p.Alpha - 1)
}

// Variance returns the variance of the Pareto distribution.
// Returns +Inf when alpha <= 2.
func (p Pareto) Variance() float64 {
	if p.Alpha <= 2 {
		return math.Inf(1)
	}

	return p.Xm * p.Xm * p.Alpha / ((p.Alpha - 1) * (p.Alpha - 1) * (p.Alpha - 2))
}

// Quantile returns the inverse CDF at probability p for the Pareto distribution.
func (pa Pareto) Quantile(prob float64) float64 {
	return pa.Xm / math.Pow(1-prob, 1/pa.Alpha)
}
