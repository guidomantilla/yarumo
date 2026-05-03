package stats

// NewUniform creates a validated uniform distribution.
func NewUniform(min, max float64) (Uniform, error) {
	if min >= max {
		return Uniform{}, ErrInvalidParameter
	}

	return Uniform{Min: min, Max: max}, nil
}

// PDF returns the probability density at x for the uniform distribution.
func (u Uniform) PDF(x float64) float64 {
	if x < u.Min || x > u.Max {
		return 0
	}

	return 1 / (u.Max - u.Min)
}

// CDF returns the cumulative probability at x for the uniform distribution.
func (u Uniform) CDF(x float64) float64 {
	if x <= u.Min {
		return 0
	}

	if x >= u.Max {
		return 1
	}

	return (x - u.Min) / (u.Max - u.Min)
}

// Mean returns the expected value of the uniform distribution.
func (u Uniform) Mean() float64 {
	return (u.Min + u.Max) / 2
}

// Variance returns the variance of the uniform distribution.
func (u Uniform) Variance() float64 {
	d := u.Max - u.Min
	return d * d / 12
}

// Quantile returns the inverse CDF at probability p.
func (u Uniform) Quantile(p float64) float64 {
	return u.Min + p*(u.Max-u.Min)
}
