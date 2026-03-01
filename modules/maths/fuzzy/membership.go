package fuzzy

import "math"

// Triangular returns a triangular membership function with vertices at a, b, c.
// The function peaks at b with degree 1, and is 0 outside [a, c].
func Triangular(a, b, c float64) MembershipFn {
	return func(x float64) Degree {
		if x <= a || x >= c {
			return 0
		}

		if x <= b {
			return Degree((x - a) / (b - a))
		}

		return Degree((c - x) / (c - b))
	}
}

// Trapezoidal returns a trapezoidal membership function.
// The function is 1 between b and c, and linearly rises from a to b and falls from c to d.
func Trapezoidal(a, b, c, d float64) MembershipFn {
	return func(x float64) Degree {
		if x <= a || x >= d {
			return 0
		}

		if x >= b && x <= c {
			return 1
		}

		if x < b {
			return Degree((x - a) / (b - a))
		}

		return Degree((d - x) / (d - c))
	}
}

// Gaussian returns a Gaussian membership function centered at center with width sigma.
func Gaussian(center, sigma float64) MembershipFn {
	return func(x float64) Degree {
		d := (x - center) / sigma
		return Degree(math.Exp(-0.5 * d * d))
	}
}

// Sigmoid returns a sigmoidal membership function.
// Positive slope gives an S-curve rising around center.
func Sigmoid(center, slope float64) MembershipFn {
	return func(x float64) Degree {
		return Degree(1.0 / (1.0 + math.Exp(-slope*(x-center))))
	}
}

// Constant returns a membership function that always returns the given degree.
func Constant(d Degree) MembershipFn {
	return func(_ float64) Degree {
		return d
	}
}
