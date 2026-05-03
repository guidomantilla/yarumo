package stats

import (
	"errors"
	"math"
	"testing"
)

// --- Sum ---

func TestSum(t *testing.T) {
	t.Parallel()

	result, err := Sum([]float64{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 15 {
		t.Fatalf("expected 15, got %f", result)
	}
}

func TestSum_empty(t *testing.T) {
	t.Parallel()

	_, err := Sum(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Mean ---

func TestMean(t *testing.T) {
	t.Parallel()

	result, err := Mean([]float64{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestMean_empty(t *testing.T) {
	t.Parallel()

	_, err := Mean(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Median ---

func TestMedian_odd(t *testing.T) {
	t.Parallel()

	result, err := Median([]float64{3, 1, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 2 {
		t.Fatalf("expected 2, got %f", result)
	}
}

func TestMedian_even(t *testing.T) {
	t.Parallel()

	result, err := Median([]float64{4, 1, 3, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 2.5 {
		t.Fatalf("expected 2.5, got %f", result)
	}
}

func TestMedian_empty(t *testing.T) {
	t.Parallel()

	_, err := Median(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Variance ---

func TestVariance(t *testing.T) {
	t.Parallel()

	result, err := Variance([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-4.0) > 1e-9 {
		t.Fatalf("expected 4.0, got %f", result)
	}
}

func TestVariance_empty(t *testing.T) {
	t.Parallel()

	_, err := Variance(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- StdDev ---

func TestStdDev(t *testing.T) {
	t.Parallel()

	result, err := StdDev([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-2.0) > 1e-9 {
		t.Fatalf("expected 2.0, got %f", result)
	}
}

func TestStdDev_empty(t *testing.T) {
	t.Parallel()

	_, err := StdDev(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Min ---

func TestMin(t *testing.T) {
	t.Parallel()

	result, err := Min([]float64{3, 1, 4, 1, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %f", result)
	}
}

func TestMin_empty(t *testing.T) {
	t.Parallel()

	_, err := Min(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Max ---

func TestMax(t *testing.T) {
	t.Parallel()

	result, err := Max([]float64{3, 1, 4, 1, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 5 {
		t.Fatalf("expected 5, got %f", result)
	}
}

func TestMax_empty(t *testing.T) {
	t.Parallel()

	_, err := Max(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Range ---

func TestRange(t *testing.T) {
	t.Parallel()

	result, err := Range([]float64{3, 1, 4, 1, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 4 {
		t.Fatalf("expected 4, got %f", result)
	}
}

func TestRange_empty(t *testing.T) {
	t.Parallel()

	_, err := Range(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Mode ---

func TestMode(t *testing.T) {
	t.Parallel()

	result, err := Mode([]float64{1, 2, 2, 3, 3, 3, 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestMode_tie(t *testing.T) {
	t.Parallel()

	result, err := Mode([]float64{1, 1, 2, 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 1 {
		t.Fatalf("expected 1 (smallest mode), got %f", result)
	}
}

func TestMode_empty(t *testing.T) {
	t.Parallel()

	_, err := Mode(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Percentile ---

func TestPercentile_50(t *testing.T) {
	t.Parallel()

	result, err := Percentile([]float64{1, 2, 3, 4, 5}, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestPercentile_100(t *testing.T) {
	t.Parallel()

	result, err := Percentile([]float64{1, 2, 3, 4, 5}, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 5 {
		t.Fatalf("expected 5, got %f", result)
	}
}

func TestPercentile_25(t *testing.T) {
	t.Parallel()

	result, err := Percentile([]float64{1, 2, 3, 4, 5}, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 2 {
		t.Fatalf("expected 2, got %f", result)
	}
}

func TestPercentile_invalid_zero(t *testing.T) {
	t.Parallel()

	_, err := Percentile([]float64{1, 2, 3}, 0)
	if !errors.Is(err, ErrInvalidPercentile) {
		t.Fatalf("expected ErrInvalidPercentile, got %v", err)
	}
}

func TestPercentile_invalid_negative(t *testing.T) {
	t.Parallel()

	_, err := Percentile([]float64{1, 2, 3}, -10)
	if !errors.Is(err, ErrInvalidPercentile) {
		t.Fatalf("expected ErrInvalidPercentile, got %v", err)
	}
}

func TestPercentile_invalid_over100(t *testing.T) {
	t.Parallel()

	_, err := Percentile([]float64{1, 2, 3}, 101)
	if !errors.Is(err, ErrInvalidPercentile) {
		t.Fatalf("expected ErrInvalidPercentile, got %v", err)
	}
}

func TestPercentile_empty(t *testing.T) {
	t.Parallel()

	_, err := Percentile(nil, 50)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestPercentile_single(t *testing.T) {
	t.Parallel()

	result, err := Percentile([]float64{42}, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 42 {
		t.Fatalf("expected 42, got %f", result)
	}
}

// --- SampleVariance ---

func TestSampleVariance(t *testing.T) {
	t.Parallel()

	// Same data as TestVariance: pop var = 4.0, sample var = 4*8/7 = 32/7.
	result, err := SampleVariance([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := 32.0 / 7.0

	if math.Abs(result-expected) > 1e-9 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestSampleVariance_insufficient(t *testing.T) {
	t.Parallel()

	_, err := SampleVariance([]float64{1})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestSampleVariance_empty(t *testing.T) {
	t.Parallel()

	_, err := SampleVariance(nil)
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

// --- SampleStdDev ---

func TestSampleStdDev(t *testing.T) {
	t.Parallel()

	result, err := SampleStdDev([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := math.Sqrt(32.0 / 7.0)

	if math.Abs(result-expected) > 1e-9 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestSampleStdDev_insufficient(t *testing.T) {
	t.Parallel()

	_, err := SampleStdDev([]float64{1})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

// --- SampleCovariance ---

func TestSampleCovariance(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	result, err := SampleCovariance(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pop cov = 4.0, sample cov = 4.0 * 5/4 = 5.0.
	if math.Abs(result-5.0) > 1e-9 {
		t.Fatalf("expected 5.0, got %f", result)
	}
}

func TestSampleCovariance_mismatch(t *testing.T) {
	t.Parallel()

	_, err := SampleCovariance([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestSampleCovariance_insufficient(t *testing.T) {
	t.Parallel()

	_, err := SampleCovariance([]float64{1}, []float64{2})
	if !errors.Is(err, ErrInsufficientData) {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

// --- LinearRegression ---

func TestLinearRegression(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	slope, intercept, err := LinearRegression(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(slope-2.0) > 1e-9 {
		t.Fatalf("expected slope 2.0, got %f", slope)
	}

	if math.Abs(intercept) > 1e-9 {
		t.Fatalf("expected intercept 0, got %f", intercept)
	}
}

func TestLinearRegression_withIntercept(t *testing.T) {
	t.Parallel()

	// y = 3x + 1.
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{4, 7, 10, 13, 16}

	slope, intercept, err := LinearRegression(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(slope-3.0) > 1e-9 {
		t.Fatalf("expected slope 3.0, got %f", slope)
	}

	if math.Abs(intercept-1.0) > 1e-9 {
		t.Fatalf("expected intercept 1.0, got %f", intercept)
	}
}

func TestLinearRegression_zeroVariance(t *testing.T) {
	t.Parallel()

	x := []float64{5, 5, 5}
	y := []float64{1, 2, 3}

	_, _, err := LinearRegression(x, y)
	if !errors.Is(err, ErrZeroVariance) {
		t.Fatalf("expected ErrZeroVariance, got %v", err)
	}
}

func TestLinearRegression_empty(t *testing.T) {
	t.Parallel()

	_, _, err := LinearRegression(nil, nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- RSquared ---

func TestRSquared_perfect(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	result, err := RSquared(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

func TestRSquared_partial(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 5, 4, 5}

	result, err := RSquared(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result <= 0 || result >= 1 {
		t.Fatalf("expected R² in (0,1), got %f", result)
	}
}

func TestRSquared_mismatch(t *testing.T) {
	t.Parallel()

	_, err := RSquared([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

// --- WeightedMean ---

func TestWeightedMean(t *testing.T) {
	t.Parallel()

	data := []float64{10, 20, 30}
	weights := []float64{1, 2, 1}

	result, err := WeightedMean(data, weights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// (10*1 + 20*2 + 30*1) / 4 = 80/4 = 20.
	if math.Abs(result-20.0) > 1e-9 {
		t.Fatalf("expected 20.0, got %f", result)
	}
}

func TestWeightedMean_equalWeights(t *testing.T) {
	t.Parallel()

	data := []float64{1, 2, 3, 4, 5}
	weights := []float64{1, 1, 1, 1, 1}

	result, err := WeightedMean(data, weights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-3.0) > 1e-9 {
		t.Fatalf("expected 3.0, got %f", result)
	}
}

func TestWeightedMean_mismatch(t *testing.T) {
	t.Parallel()

	_, err := WeightedMean([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestWeightedMean_empty(t *testing.T) {
	t.Parallel()

	_, err := WeightedMean(nil, nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestWeightedMean_zeroWeights(t *testing.T) {
	t.Parallel()

	_, err := WeightedMean([]float64{1, 2, 3}, []float64{0, 0, 0})
	if !errors.Is(err, ErrInvalidWeights) {
		t.Fatalf("expected ErrInvalidWeights, got %v", err)
	}
}

// --- WeightedVariance ---

func TestWeightedVariance(t *testing.T) {
	t.Parallel()

	data := []float64{10, 20, 30}
	weights := []float64{1, 1, 1}

	result, err := WeightedVariance(data, weights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Equal weights → same as population variance.
	popVar, _ := Variance(data)

	if math.Abs(result-popVar) > 1e-9 {
		t.Fatalf("expected %f, got %f", popVar, result)
	}
}

func TestWeightedVariance_mismatch(t *testing.T) {
	t.Parallel()

	_, err := WeightedVariance([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestWeightedVariance_zeroWeights(t *testing.T) {
	t.Parallel()

	_, err := WeightedVariance([]float64{1, 2, 3}, []float64{0, 0, 0})
	if !errors.Is(err, ErrInvalidWeights) {
		t.Fatalf("expected ErrInvalidWeights, got %v", err)
	}
}

// --- Covariance ---

func TestCovariance(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	result, err := Covariance(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-4.0) > 1e-9 {
		t.Fatalf("expected 4.0, got %f", result)
	}
}

func TestCovariance_mismatch(t *testing.T) {
	t.Parallel()

	_, err := Covariance([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestCovariance_empty(t *testing.T) {
	t.Parallel()

	_, err := Covariance(nil, nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Correlation ---

func TestCorrelation_perfect(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	result, err := Correlation(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

func TestCorrelation_negative(t *testing.T) {
	t.Parallel()

	x := []float64{1, 2, 3, 4, 5}
	y := []float64{10, 8, 6, 4, 2}

	result, err := Correlation(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-(-1.0)) > 1e-9 {
		t.Fatalf("expected -1.0, got %f", result)
	}
}

func TestCorrelation_mismatch(t *testing.T) {
	t.Parallel()

	_, err := Correlation([]float64{1, 2}, []float64{1})
	if !errors.Is(err, ErrMismatchedLengths) {
		t.Fatalf("expected ErrMismatchedLengths, got %v", err)
	}
}

func TestCorrelation_zeroVariance(t *testing.T) {
	t.Parallel()

	// Constant data: stddev = 0.
	x := []float64{5, 5, 5}
	y := []float64{1, 2, 3}

	result, err := Correlation(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 0 {
		t.Fatalf("expected 0 for zero variance, got %f", result)
	}
}

// --- IQR ---

func TestIQR(t *testing.T) {
	t.Parallel()

	result, err := IQR([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Q1=3.25, Q3=7.75 → IQR=4.5.
	if math.Abs(result-4.5) > 1e-9 {
		t.Fatalf("expected 4.5, got %f", result)
	}
}

func TestIQR_empty(t *testing.T) {
	t.Parallel()

	_, err := IQR(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- Skewness ---

func TestSkewness_symmetric(t *testing.T) {
	t.Parallel()

	result, err := Skewness([]float64{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result) > 1e-9 {
		t.Fatalf("expected 0 for symmetric data, got %f", result)
	}
}

func TestSkewness_rightSkewed(t *testing.T) {
	t.Parallel()

	result, err := Skewness([]float64{1, 1, 1, 1, 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result <= 0 {
		t.Fatalf("expected positive skewness, got %f", result)
	}
}

func TestSkewness_empty(t *testing.T) {
	t.Parallel()

	_, err := Skewness(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestSkewness_zeroVariance(t *testing.T) {
	t.Parallel()

	result, err := Skewness([]float64{5, 5, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 0 {
		t.Fatalf("expected 0 for constant data, got %f", result)
	}
}

// --- Kurtosis ---

func TestKurtosis_normal(t *testing.T) {
	t.Parallel()

	// For a uniform distribution {1,...,N}, excess kurtosis = -6(N²+1) / (5(N²-1)).
	// For N=5: -6*26 / 5*24 = -156/120 = -1.3.
	result, err := Kurtosis([]float64{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-(-1.3)) > 1e-9 {
		t.Fatalf("expected -1.3, got %f", result)
	}
}

func TestKurtosis_empty(t *testing.T) {
	t.Parallel()

	_, err := Kurtosis(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestKurtosis_zeroVariance(t *testing.T) {
	t.Parallel()

	result, err := Kurtosis([]float64{5, 5, 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 0 {
		t.Fatalf("expected 0 for constant data, got %f", result)
	}
}

// --- MAD ---

func TestMAD(t *testing.T) {
	t.Parallel()

	// Data: {1, 1, 2, 2, 4, 6, 9}. Median=2. Deviations: {1,1,0,0,2,4,7}. Median of deviations=1.
	result, err := MAD([]float64{1, 1, 2, 2, 4, 6, 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

func TestMAD_empty(t *testing.T) {
	t.Parallel()

	_, err := MAD(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

// --- GeometricMean ---

func TestGeometricMean(t *testing.T) {
	t.Parallel()

	// GM(2, 8) = sqrt(16) = 4.
	result, err := GeometricMean([]float64{2, 8})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-4.0) > 1e-9 {
		t.Fatalf("expected 4.0, got %f", result)
	}
}

func TestGeometricMean_single(t *testing.T) {
	t.Parallel()

	result, err := GeometricMean([]float64{7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-7.0) > 1e-9 {
		t.Fatalf("expected 7.0, got %f", result)
	}
}

func TestGeometricMean_empty(t *testing.T) {
	t.Parallel()

	_, err := GeometricMean(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestGeometricMean_nonPositive(t *testing.T) {
	t.Parallel()

	_, err := GeometricMean([]float64{1, 0, 3})
	if !errors.Is(err, ErrNonPositiveData) {
		t.Fatalf("expected ErrNonPositiveData, got %v", err)
	}
}

func TestGeometricMean_negative(t *testing.T) {
	t.Parallel()

	_, err := GeometricMean([]float64{1, -2, 3})
	if !errors.Is(err, ErrNonPositiveData) {
		t.Fatalf("expected ErrNonPositiveData, got %v", err)
	}
}

// --- HarmonicMean ---

func TestHarmonicMean(t *testing.T) {
	t.Parallel()

	// HM(1, 4) = 2 / (1/1 + 1/4) = 2 / 1.25 = 1.6.
	result, err := HarmonicMean([]float64{1, 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-1.6) > 1e-9 {
		t.Fatalf("expected 1.6, got %f", result)
	}
}

func TestHarmonicMean_single(t *testing.T) {
	t.Parallel()

	result, err := HarmonicMean([]float64{5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result-5.0) > 1e-9 {
		t.Fatalf("expected 5.0, got %f", result)
	}
}

func TestHarmonicMean_empty(t *testing.T) {
	t.Parallel()

	_, err := HarmonicMean(nil)
	if !errors.Is(err, ErrEmptyData) {
		t.Fatalf("expected ErrEmptyData, got %v", err)
	}
}

func TestHarmonicMean_nonPositive(t *testing.T) {
	t.Parallel()

	_, err := HarmonicMean([]float64{1, 0, 3})
	if !errors.Is(err, ErrNonPositiveData) {
		t.Fatalf("expected ErrNonPositiveData, got %v", err)
	}
}
