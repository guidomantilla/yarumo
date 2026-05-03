package stats

import "math"

// NewBinomial creates a validated binomial distribution.
func NewBinomial(n int, p float64) (Binomial, error) {
	if n <= 0 {
		return Binomial{}, ErrInvalidParameter
	}

	if p < 0 || p > 1 {
		return Binomial{}, ErrInvalidProb
	}

	return Binomial{N: n, P: p}, nil
}

// PMF returns the probability mass at k for the binomial distribution.
func (b Binomial) PMF(k int) float64 {
	if k < 0 || k > b.N {
		return 0
	}

	if b.P == 0 {
		if k == 0 {
			return 1
		}

		return 0
	}

	if b.P == 1 {
		if k == b.N {
			return 1
		}

		return 0
	}

	lnCoeff := lnFactorial(b.N) - lnFactorial(k) - lnFactorial(b.N-k)

	return math.Exp(lnCoeff + float64(k)*math.Log(b.P) + float64(b.N-k)*math.Log(1-b.P))
}

// CDFDiscrete returns the cumulative probability at k for the binomial distribution.
func (b Binomial) CDFDiscrete(k int) float64 {
	if k < 0 {
		return 0
	}

	if k >= b.N {
		return 1
	}

	sum := 0.0

	for i := range k + 1 {
		sum += b.PMF(i)
	}

	return sum
}

// Mean returns the expected value of the binomial distribution.
func (b Binomial) Mean() float64 {
	return float64(b.N) * b.P
}

// Variance returns the variance of the binomial distribution.
func (b Binomial) Variance() float64 {
	return float64(b.N) * b.P * (1 - b.P)
}
