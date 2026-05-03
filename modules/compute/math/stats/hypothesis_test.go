package stats

import (
	"errors"
	"math"
	"testing"
)

// --- TTest ---

func TestTTest_basic(t *testing.T) {
	t.Parallel()

	stat, p, err := TTest([]float64{5, 6, 7, 8, 9}, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stat <= 0 {
		t.Fatalf("expected positive t-statistic, got %f", stat)
	}

	if p >= 1 || p <= 0 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestTTest_noSignificance(t *testing.T) {
	t.Parallel()

	stat, p, err := TTest([]float64{5, 5, 5, 5, 5}, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(stat) > 1e-9 {
		t.Fatalf("expected t-statistic 0, got %f", stat)
	}

	if math.Abs(p-1) > 1e-9 {
		t.Fatalf("expected p-value 1, got %f", p)
	}
}

func TestTTest_insufficient(t *testing.T) {
	t.Parallel()

	_, _, err := TTest([]float64{1}, 0)
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestTTest_empty(t *testing.T) {
	t.Parallel()

	_, _, err := TTest(nil, 0)
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

// --- TTestTwoSample ---

func TestTTestTwoSample_basic(t *testing.T) {
	t.Parallel()

	stat, p, err := TTestTwoSample([]float64{1, 2, 3}, []float64{4, 5, 6})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stat >= 0 {
		t.Fatalf("expected negative t-statistic, got %f", stat)
	}

	if p >= 1 || p <= 0 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestTTestTwoSample_equal(t *testing.T) {
	t.Parallel()

	stat, p, err := TTestTwoSample([]float64{5, 5, 5}, []float64{5, 5, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(stat) > 1e-9 {
		t.Fatalf("expected t-statistic 0, got %f", stat)
	}

	if math.Abs(p-1) > 1e-9 {
		t.Fatalf("expected p-value 1, got %f", p)
	}
}

func TestTTestTwoSample_xInsufficient(t *testing.T) {
	t.Parallel()

	_, _, err := TTestTwoSample([]float64{1}, []float64{4, 5, 6})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestTTestTwoSample_yInsufficient(t *testing.T) {
	t.Parallel()

	_, _, err := TTestTwoSample([]float64{1, 2, 3}, []float64{1})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

// --- ChiSquaredTest ---

func TestChiSquaredTest_basic(t *testing.T) {
	t.Parallel()

	stat, p, err := ChiSquaredTest([]float64{10, 20, 30}, []float64{20, 20, 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// chi2 = (10-20)^2/20 + (20-20)^2/20 + (30-20)^2/20 = 5 + 0 + 5 = 10.
	if math.Abs(stat-10.0) > 1e-9 {
		t.Fatalf("expected chi2 10.0, got %f", stat)
	}

	if p >= 1 || p <= 0 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestChiSquaredTest_perfect(t *testing.T) {
	t.Parallel()

	stat, p, err := ChiSquaredTest([]float64{10, 20, 30}, []float64{10, 20, 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(stat) > 1e-9 {
		t.Fatalf("expected chi2 0, got %f", stat)
	}

	if math.Abs(p-1) > 1e-9 {
		t.Fatalf("expected p-value 1, got %f", p)
	}
}

func TestChiSquaredTest_mismatch(t *testing.T) {
	t.Parallel()

	_, _, err := ChiSquaredTest([]float64{10, 20}, []float64{10})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestChiSquaredTest_empty(t *testing.T) {
	t.Parallel()

	_, _, err := ChiSquaredTest(nil, nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestChiSquaredTest_zeroExpected(t *testing.T) {
	t.Parallel()

	_, _, err := ChiSquaredTest([]float64{10, 20, 30}, []float64{10, 0, 20})
	if !errors.Is(err, ErrZeroExpected) {
		t.Fatalf("expected ErrZeroExpected, got %v", err)
	}
}

// --- KSTest ---

func TestKSTest_uniform(t *testing.T) {
	t.Parallel()

	// Perfect uniform sample on [0,1].
	sample := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
	cdf := func(x float64) float64 {
		if x < 0 {
			return 0
		}

		if x > 1 {
			return 1
		}

		return x
	}

	d, p, err := KSTest(sample, cdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d < 0 || d > 1 {
		t.Fatalf("expected D in [0,1], got %f", d)
	}

	// Should not reject: p should be high.
	if p < 0.05 {
		t.Fatalf("expected high p-value for uniform data, got %f", p)
	}
}

func TestKSTest_reject(t *testing.T) {
	t.Parallel()

	// Data clearly not from N(0,1).
	sample := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	normal := Normal{Mu: 0, Sigma: 1}

	d, p, err := KSTest(sample, normal.CDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d <= 0 {
		t.Fatalf("expected positive D, got %f", d)
	}

	if p > 0.05 {
		t.Fatalf("expected low p-value for mismatched distribution, got %f", p)
	}
}

func TestKSTest_empty(t *testing.T) {
	t.Parallel()

	_, _, err := KSTest(nil, func(x float64) float64 { return x })
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestKSTest_perfectMatch(t *testing.T) {
	t.Parallel()

	// Single point at the median of N(0,1).
	sample := []float64{0}
	normal := Normal{Mu: 0, Sigma: 1}

	d, _, err := KSTest(sample, normal.CDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// D should be 0.5 (CDF(0)=0.5, empirical CDF jumps from 0 to 1 at x=0).
	if math.Abs(d-0.5) > 1e-9 {
		t.Fatalf("expected D=0.5, got %f", d)
	}
}

// --- ANOVA ---

func TestANOVA_basic(t *testing.T) {
	t.Parallel()

	g1 := []float64{1, 2, 3}
	g2 := []float64{4, 5, 6}
	g3 := []float64{7, 8, 9}

	f, p, err := ANOVA(g1, g2, g3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f <= 0 {
		t.Fatalf("expected positive F-statistic, got %f", f)
	}

	if p >= 1 || p <= 0 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestANOVA_equalGroups(t *testing.T) {
	t.Parallel()

	g1 := []float64{5, 5, 5}
	g2 := []float64{5, 5, 5}

	f, p, err := ANOVA(g1, g2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(f) > 1e-9 {
		t.Fatalf("expected F=0 for equal groups, got %f", f)
	}

	if math.Abs(p-1) > 1e-9 {
		t.Fatalf("expected p=1 for equal groups, got %f", p)
	}
}

func TestANOVA_significantDifference(t *testing.T) {
	t.Parallel()

	g1 := []float64{1, 1, 1, 1, 1}
	g2 := []float64{100, 100, 100, 100, 100}

	_, p, err := ANOVA(g1, g2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p > 0.01 {
		t.Fatalf("expected very low p-value, got %f", p)
	}
}

func TestANOVA_zeroWithinVariance(t *testing.T) {
	t.Parallel()

	g1 := []float64{1, 1}
	g2 := []float64{2, 2}

	f, _, err := ANOVA(g1, g2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !math.IsInf(f, 1) {
		t.Fatalf("expected +Inf F for zero within-group variance, got %f", f)
	}
}

func TestANOVA_insufficientDFWithin(t *testing.T) {
	t.Parallel()

	// 3 groups of 1 element each: dfWithin = 3 - 3 = 0.
	_, _, err := ANOVA([]float64{1}, []float64{2}, []float64{3})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestANOVA_insufficientGroups(t *testing.T) {
	t.Parallel()

	_, _, err := ANOVA([]float64{1, 2, 3})
	if !errors.Is(err, ErrInsufficientGroups) {
		t.Fatalf("expected ErrInsufficientGroups, got %v", err)
	}
}

func TestANOVA_emptyGroup(t *testing.T) {
	t.Parallel()

	_, _, err := ANOVA([]float64{1, 2}, []float64{})
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- MannWhitneyU ---

func TestMannWhitneyU_basic(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{6, 7, 8, 9, 10}

	u, p, err := MannWhitneyU(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// U = min(U1, U2). Complete separation → U = 0.
	if math.Abs(u) > 1e-9 {
		t.Fatalf("expected U=0, got %f", u)
	}

	if p > 0.05 {
		t.Fatalf("expected low p-value for separated groups, got %f", p)
	}
}

func TestMannWhitneyU_equalGroups(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3}
	y := []float64{1, 2, 3}

	u, p, err := MannWhitneyU(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Equal groups → U ≈ n1*n2/2.
	expected := float64(len(x)*len(y)) / 2.0

	if math.Abs(u-expected) > 1e-9 {
		t.Fatalf("expected U=%f, got %f", expected, u)
	}

	if p < 0.9 {
		t.Fatalf("expected high p-value for equal groups, got %f", p)
	}
}

func TestMannWhitneyU_allTied(t *testing.T) {
	t.Parallel()

	x := []float64{5, 5, 5}
	y := []float64{5, 5, 5}

	_, p, err := MannWhitneyU(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(p-1.0) > 1e-9 {
		t.Fatalf("expected p=1.0 for all-tied, got %f", p)
	}
}

func TestMannWhitneyU_emptyX(t *testing.T) {
	t.Parallel()

	_, _, err := MannWhitneyU(nil, []float64{1, 2, 3})
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestMannWhitneyU_emptyY(t *testing.T) {
	t.Parallel()

	_, _, err := MannWhitneyU([]float64{1, 2, 3}, nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- FisherExact ---

func TestFisherExact_basic(t *testing.T) {
	t.Parallel()

	// Classic Lady tasting tea: [[3, 1], [1, 3]].
	or, p, err := FisherExact(3, 1, 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// OR = 3*3 / (1*1) = 9.
	if math.Abs(or-9.0) > 1e-9 {
		t.Fatalf("expected OR=9, got %f", or)
	}

	if p <= 0 || p >= 1 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestFisherExact_noAssociation(t *testing.T) {
	t.Parallel()

	// Balanced table: [[5, 5], [5, 5]].
	or, p, err := FisherExact(5, 5, 5, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(or-1.0) > 1e-9 {
		t.Fatalf("expected OR=1, got %f", or)
	}

	if p < 0.9 {
		t.Fatalf("expected high p-value for no association, got %f", p)
	}
}

func TestFisherExact_zeroCells(t *testing.T) {
	t.Parallel()

	// [[0, 5], [5, 0]].
	or, p, err := FisherExact(0, 5, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// b*c = 25 ≠ 0, a*d = 0 → OR = 0.
	if or != 0 {
		t.Fatalf("expected OR=0, got %f", or)
	}

	if p <= 0 || p >= 1 {
		t.Fatalf("expected p-value in (0,1), got %f", p)
	}
}

func TestFisherExact_infiniteOR(t *testing.T) {
	t.Parallel()

	// b=0 → OR = Inf.
	or, _, err := FisherExact(3, 0, 2, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !math.IsInf(or, 1) {
		t.Fatalf("expected +Inf OR, got %f", or)
	}
}

func TestFisherExact_negativeInput(t *testing.T) {
	t.Parallel()

	_, _, err := FisherExact(-1, 2, 3, 4)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestFisherExact_allZero(t *testing.T) {
	t.Parallel()

	_, _, err := FisherExact(0, 0, 0, 0)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}
