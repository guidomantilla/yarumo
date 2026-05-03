package fuzzy

import "math"

// Min is the standard fuzzy AND (t-norm): min(a, b).
func Min(a, b Degree) Degree {
	if a < b {
		return a
	}

	return b
}

// Product is the product t-norm: a * b.
func Product(a, b Degree) Degree {
	return Degree(float64(a) * float64(b))
}

// Lukasiewicz is the Lukasiewicz t-norm: max(a + b - 1, 0).
func Lukasiewicz(a, b Degree) Degree {
	v := float64(a) + float64(b) - 1
	if v < 0 {
		return 0
	}

	return Degree(v)
}

// Max is the standard fuzzy OR (t-conorm): max(a, b).
func Max(a, b Degree) Degree {
	if a > b {
		return a
	}

	return b
}

// ProbabilisticSum is the probabilistic sum t-conorm: a + b - a*b.
func ProbabilisticSum(a, b Degree) Degree {
	return Degree(float64(a) + float64(b) - float64(a)*float64(b))
}

// BoundedSum is the bounded sum t-conorm: min(a + b, 1).
func BoundedSum(a, b Degree) Degree {
	v := float64(a) + float64(b)

	return Degree(math.Min(v, 1))
}

// Complement is the standard fuzzy NOT: 1 - d.
func Complement(d Degree) Degree {
	return 1 - d
}
