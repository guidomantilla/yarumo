package probability

import "math"

const epsilon = 1e-9

// IsValid reports whether a distribution sums to approximately 1.0.
func IsValid(d Distribution) bool {
	if len(d) == 0 {
		return false
	}

	var sum float64

	for _, p := range d {
		if p < 0 || p > 1 {
			return false
		}

		sum += float64(p)
	}

	return math.Abs(sum-1.0) < epsilon
}

// Normalize rescales a distribution so its probabilities sum to 1.
func Normalize(d Distribution) (Distribution, error) {
	if len(d) == 0 {
		return nil, ErrEmptyDist
	}

	var sum float64

	for _, p := range d {
		sum += float64(p)
	}

	if sum == 0 {
		return nil, ErrNotNormalized
	}

	result := make(Distribution, len(d))

	for o, p := range d {
		result[o] = Prob(float64(p) / sum)
	}

	return result, nil
}

// Complement returns 1 - p.
func Complement(p Prob) Prob {
	return 1 - p
}

// Entropy computes the Shannon entropy of a distribution: -Sum p*log2(p).
func Entropy(d Distribution) (float64, error) {
	if len(d) == 0 {
		return 0, ErrEmptyDist
	}

	var h float64

	for _, p := range d {
		if p < 0 || p > 1 {
			return 0, ErrInvalidProb
		}

		if p > 0 {
			h -= float64(p) * math.Log2(float64(p))
		}
	}

	return h, nil
}
