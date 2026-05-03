package stats

import "math"

// NewHypergeometric creates a validated hypergeometric distribution.
// N is the population size, K is the number of success states, n is the number of draws.
func NewHypergeometric(populationN, successK, drawsN int) (Hypergeometric, error) {
	if populationN <= 0 {
		return Hypergeometric{}, ErrInvalidParameter
	}

	if successK < 0 || successK > populationN {
		return Hypergeometric{}, ErrInvalidParameter
	}

	if drawsN <= 0 || drawsN > populationN {
		return Hypergeometric{}, ErrInvalidParameter
	}

	return Hypergeometric{N: populationN, K: successK, Draws: drawsN}, nil
}

// PMF returns the probability mass at k for the hypergeometric distribution.
func (h Hypergeometric) PMF(k int) float64 {
	minK := max(0, h.Draws-(h.N-h.K))
	maxK := min(h.K, h.Draws)

	if k < minK || k > maxK {
		return 0
	}

	// P(X=k) = C(K,k) * C(N-K, n-k) / C(N, n).
	lnNum := lnBinom(h.K, k) + lnBinom(h.N-h.K, h.Draws-k)
	lnDen := lnBinom(h.N, h.Draws)

	return math.Exp(lnNum - lnDen)
}

// CDFDiscrete returns the cumulative probability at k for the hypergeometric distribution.
func (h Hypergeometric) CDFDiscrete(k int) float64 {
	minK := max(0, h.Draws-(h.N-h.K))

	if k < minK {
		return 0
	}

	maxK := min(h.K, h.Draws)

	if k >= maxK {
		return 1
	}

	sum := 0.0

	for i := minK; i <= k; i++ {
		sum += h.PMF(i)
	}

	return sum
}

// Mean returns the expected value of the hypergeometric distribution.
func (h Hypergeometric) Mean() float64 {
	return float64(h.Draws) * float64(h.K) / float64(h.N)
}

// Variance returns the variance of the hypergeometric distribution.
func (h Hypergeometric) Variance() float64 {
	n := float64(h.N)
	k := float64(h.K)
	d := float64(h.Draws)

	return d * (k / n) * ((n - k) / n) * ((n - d) / (n - 1))
}

// lnBinom returns the natural logarithm of C(n, k).
func lnBinom(n, k int) float64 {
	return lnFactorial(n) - lnFactorial(k) - lnFactorial(n-k)
}
