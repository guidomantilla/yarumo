package stats

import (
	"math"
	"slices"
)

// TTest performs a one-sample t-test, testing whether the sample mean
// differs significantly from mu. Returns the t-statistic and two-tailed p-value.
func TTest(sample []float64, mu float64) (float64, float64, error) {
	n := len(sample)

	if n < 2 {
		return 0, 0, ErrInsufficientData
	}

	m, _ := Mean(sample)
	sd, _ := SampleStdDev(sample)

	se := sd / math.Sqrt(float64(n))
	if se == 0 {
		if m == mu {
			return 0, 1, nil
		}

		return math.Inf(1), 0, nil
	}

	t := (m - mu) / se

	dist := StudentT{Nu: float64(n - 1)}
	cdf := dist.CDF(t)

	p := 2 * math.Min(cdf, 1-cdf)

	return t, p, nil
}

// TTestTwoSample performs Welch's two-sample t-test, testing whether
// x and y have different means. Returns the t-statistic and two-tailed p-value.
func TTestTwoSample(x, y []float64) (float64, float64, error) {
	nX := len(x)
	nY := len(y)

	if nX < 2 {
		return 0, 0, ErrInsufficientData
	}

	if nY < 2 {
		return 0, 0, ErrInsufficientData
	}

	mX, _ := Mean(x)
	mY, _ := Mean(y)
	vX, _ := SampleVariance(x)
	vY, _ := SampleVariance(y)

	fnX := float64(nX)
	fnY := float64(nY)

	se := math.Sqrt(vX/fnX + vY/fnY)

	if se == 0 {
		if mX == mY {
			return 0, 1, nil
		}

		return math.Inf(1), 0, nil
	}

	t := (mX - mY) / se

	// Welch-Satterthwaite degrees of freedom.
	num := (vX/fnX + vY/fnY) * (vX/fnX + vY/fnY)
	denom := (vX/fnX)*(vX/fnX)/(fnX-1) + (vY/fnY)*(vY/fnY)/(fnY-1)
	df := num / denom

	dist := StudentT{Nu: df}
	cdf := dist.CDF(t)

	p := 2 * math.Min(cdf, 1-cdf)

	return t, p, nil
}

// ChiSquaredTest performs a chi-squared goodness-of-fit test.
// Returns the chi-squared statistic and the p-value.
func ChiSquaredTest(observed, expected []float64) (float64, float64, error) {
	if len(observed) != len(expected) {
		return 0, 0, ErrMismatchedLengths
	}

	if len(observed) == 0 {
		return 0, 0, ErrEmptyData
	}

	for _, e := range expected {
		if e <= 0 {
			return 0, 0, ErrZeroExpected
		}
	}

	var chi2 float64

	for i := range observed {
		d := observed[i] - expected[i]
		chi2 += d * d / expected[i]
	}

	df := float64(len(observed) - 1)

	dist := ChiSquared{K: df}
	p := 1 - dist.CDF(chi2)

	return chi2, p, nil
}

// KSTest performs a one-sample Kolmogorov-Smirnov test, comparing the sample
// against a reference CDF. Returns the KS statistic and approximate p-value.
func KSTest(sample []float64, cdf func(float64) float64) (float64, float64, error) {
	n := len(sample)

	if n == 0 {
		return 0, 0, ErrEmptyData
	}

	sorted := make([]float64, n)
	copy(sorted, sample)

	slices.Sort(sorted)

	fn := float64(n)
	d := 0.0

	for i, x := range sorted {
		fi := float64(i)
		cdfX := cdf(x)

		// D+ = max(i/n - F(x_i)).
		dPlus := (fi+1)/fn - cdfX

		// D- = max(F(x_i) - (i-1)/n).
		dMinus := cdfX - fi/fn

		d = math.Max(d, math.Max(dPlus, dMinus))
	}

	// Approximate p-value using the asymptotic formula.
	p := ksP(d, n)

	return d, p, nil
}

// ksP computes an approximate p-value for the KS statistic using the
// asymptotic Kolmogorov distribution.
func ksP(d float64, n int) float64 {
	z := (math.Sqrt(float64(n)) + 0.12 + 0.11/math.Sqrt(float64(n))) * d

	if z < 1e-10 {
		return 1.0
	}

	// Kolmogorov asymptotic formula: P(D > d) ≈ 2 * Σ (-1)^(k-1) * exp(-2k²z²).
	sum := 0.0

	for k := 1; k <= 100; k++ {
		fk := float64(k)
		sign := 1.0

		if k%2 == 0 {
			sign = -1.0
		}

		term := sign * math.Exp(-2*fk*fk*z*z)
		sum += term

		if math.Abs(term) < 1e-14 {
			break
		}
	}

	p := 2 * sum

	return math.Max(0, math.Min(1, p))
}

// ANOVA performs a one-way analysis of variance, testing whether the means
// of two or more groups are significantly different.
// Returns the F-statistic and p-value.
func ANOVA(groups ...[]float64) (float64, float64, error) {
	if len(groups) < 2 {
		return 0, 0, ErrInsufficientGroups
	}

	totalN := 0

	for _, g := range groups {
		if len(g) == 0 {
			return 0, 0, ErrEmptyData
		}

		totalN += len(g)
	}

	// Grand mean.
	grandSum := 0.0

	for _, g := range groups {
		for _, v := range g {
			grandSum += v
		}
	}

	grandMean := grandSum / float64(totalN)

	// Between-group sum of squares.
	ssb := 0.0

	for _, g := range groups {
		gm, _ := Mean(g)
		d := gm - grandMean
		ssb += float64(len(g)) * d * d
	}

	// Within-group sum of squares.
	ssw := 0.0

	for _, g := range groups {
		gm, _ := Mean(g)

		for _, v := range g {
			d := v - gm
			ssw += d * d
		}
	}

	k := float64(len(groups))
	n := float64(totalN)

	dfBetween := k - 1
	dfWithin := n - k

	if dfWithin <= 0 {
		return 0, 0, ErrInsufficientData
	}

	msb := ssb / dfBetween
	msw := ssw / dfWithin

	if msw == 0 {
		if msb == 0 {
			return 0, 1, nil
		}

		return math.Inf(1), 0, nil
	}

	f := msb / msw

	dist := FDist{D1: dfBetween, D2: dfWithin}
	p := 1 - dist.CDF(f)

	return f, p, nil
}

// MannWhitneyU performs a two-sample Mann-Whitney U test (Wilcoxon rank-sum test),
// testing whether two independent samples come from the same distribution.
// Returns the U statistic and approximate two-tailed p-value using normal approximation.
func MannWhitneyU(x, y []float64) (float64, float64, error) {
	nX := len(x)
	nY := len(y)

	if nX == 0 || nY == 0 {
		return 0, 0, ErrEmptyData
	}

	type ranked struct {
		value float64
		group int // 0 = x, 1 = y
	}

	combined := make([]ranked, 0, nX+nY)

	for _, v := range x {
		combined = append(combined, ranked{value: v, group: 0})
	}

	for _, v := range y {
		combined = append(combined, ranked{value: v, group: 1})
	}

	slices.SortFunc(combined, func(a, b ranked) int {
		if a.value < b.value {
			return -1
		}

		if a.value > b.value {
			return 1
		}

		return 0
	})

	// Assign average ranks for ties.
	n := len(combined)
	ranks := make([]float64, n)

	for i := 0; i < n; {
		j := i + 1

		for j < n && combined[j].value == combined[i].value {
			j++
		}

		avgRank := float64(i+j+1) / 2.0

		for k := i; k < j; k++ {
			ranks[k] = avgRank
		}

		i = j
	}

	// Sum ranks for group x.
	r1 := 0.0

	for i, r := range ranks {
		if combined[i].group == 0 {
			r1 += r
		}
	}

	fnX := float64(nX)
	fnY := float64(nY)

	u1 := r1 - fnX*(fnX+1)/2
	u2 := fnX*fnY - u1
	u := math.Min(u1, u2)

	// Normal approximation for p-value.
	mu := fnX * fnY / 2
	sigma := math.Sqrt(fnX * fnY * (fnX + fnY + 1) / 12)

	if sigma == 0 {
		return u, 1.0, nil
	}

	z := (u - mu) / sigma

	normal := Normal{Mu: 0, Sigma: 1}
	p := 2 * normal.CDF(z)

	return u, p, nil
}

// FisherExact performs a Fisher exact test on a 2x2 contingency table.
// The table is specified as [[a, b], [c, d]].
// Returns the odds ratio and the two-tailed p-value.
func FisherExact(a, b, c, d int) (float64, float64, error) {
	if a < 0 || b < 0 || c < 0 || d < 0 {
		return 0, 0, ErrInvalidParameter
	}

	n := a + b + c + d

	if n == 0 {
		return 0, 0, ErrEmptyData
	}

	// P(table) = C(a+b,a) * C(c+d,c) / C(n,a+c).
	observedP := fisherP(a, b, c, d)

	// Two-tailed: sum all tables with P <= observedP.
	r1 := a + b
	c1 := a + c

	pValue := 0.0
	minA := max(0, c1-(c+d))
	maxA := min(r1, c1)

	for ai := minA; ai <= maxA; ai++ {
		bi := r1 - ai
		ci := c1 - ai
		di := (c + d) - ci

		p := fisherP(ai, bi, ci, di)

		if p <= observedP+1e-12 {
			pValue += p
		}
	}

	pValue = math.Min(1.0, pValue)

	// Odds ratio.
	var or float64

	if b == 0 || c == 0 {
		or = math.Inf(1)
	} else {
		or = float64(a) * float64(d) / (float64(b) * float64(c))
	}

	return or, pValue, nil
}

// fisherP computes the hypergeometric probability of a 2x2 contingency table.
func fisherP(a, b, c, d int) float64 {
	n := a + b + c + d

	return math.Exp(lnFactorial(a+b) + lnFactorial(c+d) + lnFactorial(a+c) + lnFactorial(b+d) -
		lnFactorial(a) - lnFactorial(b) - lnFactorial(c) - lnFactorial(d) - lnFactorial(n))
}
