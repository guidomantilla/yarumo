package fuzzy

// Fuzzify evaluates a membership function at a crisp value.
func Fuzzify(fn MembershipFn, x float64) Degree {
	return fn(x)
}

// Clip returns a membership function clipped at the given level (alpha-cut).
// Output degree is min(fn(x), level).
func Clip(fn MembershipFn, level Degree) MembershipFn {
	return func(x float64) Degree {
		d := fn(x)
		if d > level {
			return level
		}

		return d
	}
}

// Scale returns a membership function scaled by the given level.
// Output degree is fn(x) * level.
func Scale(fn MembershipFn, level Degree) MembershipFn {
	return func(x float64) Degree {
		return Degree(float64(fn(x)) * float64(level))
	}
}

// AggregateMax combines multiple membership functions using max (union).
func AggregateMax(fns ...MembershipFn) MembershipFn {
	return func(x float64) Degree {
		var maxD Degree

		for _, fn := range fns {
			d := fn(x)
			if d > maxD {
				maxD = d
			}
		}

		return maxD
	}
}

// Sample evaluates a membership function at n evenly-spaced points in [min, max].
func Sample(fn MembershipFn, lo, hi float64, n int) ([]float64, []Degree, error) {
	if n <= 0 {
		return nil, nil, ErrEmptySamples
	}

	if lo >= hi {
		return nil, nil, ErrInvalidRange
	}

	xs := make([]float64, n)
	ys := make([]Degree, n)
	step := (hi - lo) / float64(n-1)

	if n == 1 {
		xs[0] = lo
		ys[0] = fn(lo)

		return xs, ys, nil
	}

	for i := range n {
		xs[i] = lo + float64(i)*step
		ys[i] = fn(xs[i])
	}

	return xs, ys, nil
}
