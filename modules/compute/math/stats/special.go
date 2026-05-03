package stats

import "math"

// regularizedIncompleteBeta computes the regularized incomplete beta function I_x(a, b)
// using the Lentz continued fraction algorithm (Numerical Recipes, section 6.4).
func regularizedIncompleteBeta(x, a, b float64) float64 {
	if x <= 0 {
		return 0
	}

	if x >= 1 {
		return 1
	}

	// Use symmetry relation: I_x(a,b) = 1 - I_{1-x}(b,a) when x > (a+1)/(a+b+2).
	// Strict inequality avoids infinite recursion when a == b and x == 0.5.
	if x > (a+1)/(a+b+2) {
		return 1 - regularizedIncompleteBeta(1-x, b, a)
	}

	lnPrefactor := a*math.Log(x) + b*math.Log(1-x) - lnBeta(a, b)
	prefactor := math.Exp(lnPrefactor)

	// Evaluate the continued fraction using the modified Lentz algorithm.
	cf := betaContinuedFraction(x, a, b)

	return prefactor * cf / a
}

// betaContinuedFraction evaluates the continued fraction for the incomplete beta function
// using the modified Lentz algorithm. Returns the value of:
//
//	1/(1 + d1/(1 + d2/(1 + ...)))
//
// where d_{2m+1} = -(a+m)(a+b+m)x / ((a+2m)(a+2m+1))
// and   d_{2m}   = m(b-m)x / ((a+2m-1)(a+2m))
//
// This is equivalent to computing the CF f = b0 + a1/(b1 + a2/(b2 + ...))
// where b_i = 1 for all i, a_1 = d1, a_2 = d2, etc.
// Then the function returns 1/f.
func betaContinuedFraction(x, a, b float64) float64 {
	const (
		maxIterations = 200
		epsilon       = 1e-14
		tiny          = 1e-30
	)

	// Lentz algorithm: f_n = b0 + a1/(b1 + a2/(b2+...))
	// with b0 = 1, b_i = 1, and a_i = d_i.
	// Initialize: f = b0 = 1, C = b0 = 1, D = 0.
	f := 1.0
	c := 1.0
	d := 0.0

	for i := 1; i <= maxIterations; i++ {
		var ai float64

		if i%2 == 1 {
			// Odd index: m = (i-1)/2.
			m := float64((i - 1) / 2)
			ai = -(a + m) * (a + b + m) * x / ((a + 2*m) * (a + 2*m + 1))
		} else {
			// Even index: m = i/2.
			m := float64(i / 2)
			ai = m * (b - m) * x / ((a + 2*m - 1) * (a + 2*m))
		}

		// b_i = 1 for all i.
		bi := 1.0

		d = bi + ai*d
		if math.Abs(d) < tiny {
			d = tiny
		}

		d = 1 / d

		c = bi + ai/c
		if math.Abs(c) < tiny {
			c = tiny
		}

		delta := c * d
		f *= delta

		if math.Abs(delta-1) < epsilon {
			break
		}
	}

	// f now equals 1 + d1/(1 + d2/(1+...)), and we need 1/f.
	return 1 / f
}

// incompleteGamma computes the regularized lower incomplete gamma function P(a, x).
func incompleteGamma(a, x float64) float64 {
	if x <= 0 {
		return 0
	}

	// Use series expansion when x < a+1, otherwise use upper gamma complement.
	if x < a+1 {
		return incompleteGammaSeries(a, x)
	}

	return 1 - incompleteGammaUpperCF(a, x)
}

// incompleteGammaSeries computes P(a,x) via series expansion.
func incompleteGammaSeries(a, x float64) float64 {
	const (
		maxIterations = 200
		epsilon       = 1e-14
	)

	lnPrefactor := a*math.Log(x) - x - lnGamma(a)

	sum := 1.0 / a
	term := 1.0 / a

	for n := 1; n < maxIterations; n++ {
		term *= x / (a + float64(n))
		sum += term

		if math.Abs(term) < math.Abs(sum)*epsilon {
			break
		}
	}

	return sum * math.Exp(lnPrefactor)
}

// incompleteGammaUpperCF computes Q(a,x) = 1 - P(a,x) via continued fraction.
func incompleteGammaUpperCF(a, x float64) float64 {
	const (
		maxIterations = 200
		epsilon       = 1e-14
		tiny          = 1e-30
	)

	lnPrefactor := a*math.Log(x) - x - lnGamma(a)

	// Modified Lentz's algorithm for the CF representation of Q(a,x).
	f := tiny
	c := tiny
	d := 1.0

	for i := range maxIterations {
		an, bn := gammaCFCoefficients(i, a, x)

		d = bn + an*d
		if math.Abs(d) < tiny {
			d = tiny
		}

		c = bn + an/c
		if math.Abs(c) < tiny {
			c = tiny
		}

		d = 1 / d

		delta := c * d
		f *= delta

		if math.Abs(delta-1) < epsilon {
			break
		}
	}

	return math.Exp(lnPrefactor) * f
}

// gammaCFCoefficients returns the continued fraction coefficients (an, bn) for iteration i.
func gammaCFCoefficients(i int, a, x float64) (float64, float64) {
	if i == 0 {
		return 1.0, x + 1 - a
	}

	fi := float64(i)

	return -fi * (fi - a), x + 2*fi + 1 - a
}

// lnGamma returns the natural logarithm of the gamma function.
func lnGamma(x float64) float64 {
	v, _ := math.Lgamma(x)
	return v
}

// lnBeta returns the log of the beta function: lnGamma(a) + lnGamma(b) - lnGamma(a+b).
func lnBeta(a, b float64) float64 {
	return lnGamma(a) + lnGamma(b) - lnGamma(a+b)
}

// bisectQuantile computes the inverse CDF at probability p using bisection.
// If expand is true, hi is doubled until cdf(hi) >= p before bisecting.
func bisectQuantile(lo, hi, p float64, cdf func(float64) float64, expand bool) float64 {
	if expand {
		for cdf(hi) < p {
			hi *= 2
		}
	}

	for range 100 {
		mid := (lo + hi) / 2

		if cdf(mid) < p {
			lo = mid
		} else {
			hi = mid
		}
	}

	return (lo + hi) / 2
}
