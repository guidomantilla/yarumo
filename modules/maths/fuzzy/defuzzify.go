package fuzzy

// Centroid computes the center of gravity: Sum(x*y) / Sum(y).
func Centroid(xs []float64, ys []Degree) float64 {
	var num, den float64

	for i := range xs {
		num += xs[i] * float64(ys[i])
		den += float64(ys[i])
	}

	if den == 0 {
		return 0
	}

	return num / den
}

// Bisector finds the x value that divides the area under the curve into two equal halves.
func Bisector(xs []float64, ys []Degree) float64 {
	var total float64

	for _, y := range ys {
		total += float64(y)
	}

	half := total / 2

	var running float64

	for i := range xs {
		running += float64(ys[i])

		if running >= half {
			return xs[i]
		}
	}

	return 0
}

// MeanOfMax returns the mean of all x values where the membership degree is maximal.
func MeanOfMax(xs []float64, ys []Degree) float64 {
	if len(xs) == 0 {
		return 0
	}

	maxD := ys[0]

	for _, y := range ys[1:] {
		if y > maxD {
			maxD = y
		}
	}

	var sum float64

	var count int

	for i := range xs {
		if ys[i] == maxD {
			sum += xs[i]
			count++
		}
	}

	return sum / float64(count)
}

// LargestOfMax returns the largest x value where the membership degree is maximal.
func LargestOfMax(xs []float64, ys []Degree) float64 {
	if len(xs) == 0 {
		return 0
	}

	maxD := ys[0]
	maxX := xs[0]

	for i := 1; i < len(xs); i++ {
		if ys[i] > maxD {
			maxD = ys[i]
			maxX = xs[i]
		} else if ys[i] == maxD && xs[i] > maxX {
			maxX = xs[i]
		}
	}

	return maxX
}

// SmallestOfMax returns the smallest x value where the membership degree is maximal.
func SmallestOfMax(xs []float64, ys []Degree) float64 {
	if len(xs) == 0 {
		return 0
	}

	maxD := ys[0]
	minX := xs[0]

	for i := 1; i < len(xs); i++ {
		if ys[i] > maxD {
			maxD = ys[i]
			minX = xs[i]
		}
	}

	return minX
}
