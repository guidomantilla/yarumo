package stats

import (
	"math"
	"slices"
)

// Sum returns the sum of the data.
func Sum(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var s float64

	for _, v := range data {
		s += v
	}

	return s, nil
}

// Mean returns the arithmetic mean of the data.
func Mean(data []float64) (float64, error) {
	s, err := Sum(data)
	if err != nil {
		return 0, err
	}

	return s / float64(len(data)), nil
}

// Median returns the median of the data.
func Median(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)

	slices.Sort(sorted)

	n := len(sorted)

	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2, nil
	}

	return sorted[n/2], nil
}

// Variance returns the population variance of the data.
func Variance(data []float64) (float64, error) {
	m, err := Mean(data)
	if err != nil {
		return 0, err
	}

	var sum float64

	for _, v := range data {
		d := v - m
		sum += d * d
	}

	return sum / float64(len(data)), nil
}

// StdDev returns the population standard deviation of the data.
func StdDev(data []float64) (float64, error) {
	v, err := Variance(data)
	if err != nil {
		return 0, err
	}

	return math.Sqrt(v), nil
}

// Min returns the minimum value in the data.
func Min(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	m := data[0]

	for _, v := range data[1:] {
		if v < m {
			m = v
		}
	}

	return m, nil
}

// Max returns the maximum value in the data.
func Max(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	m := data[0]

	for _, v := range data[1:] {
		if v > m {
			m = v
		}
	}

	return m, nil
}

// Range returns the difference between the maximum and minimum values.
func Range(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	minV := data[0]
	maxV := data[0]

	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}

		if v > maxV {
			maxV = v
		}
	}

	return maxV - minV, nil
}

// Mode returns the most frequently occurring value. If multiple modes exist,
// returns the smallest one.
func Mode(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	counts := make(map[float64]int)

	for _, v := range data {
		counts[v]++
	}

	maxCount := 0
	mode := data[0]

	for v, c := range counts {
		if c > maxCount || (c == maxCount && v < mode) {
			maxCount = c
			mode = v
		}
	}

	return mode, nil
}

// Percentile returns the p-th percentile of the data using linear interpolation.
// p must be in (0, 100].
func Percentile(data []float64, p float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	if p <= 0 || p > 100 {
		return 0, ErrInvalidPercentile
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)

	slices.Sort(sorted)

	n := float64(len(sorted))
	rank := p / 100 * (n - 1)
	lower := int(rank)
	frac := rank - float64(lower)

	if lower+1 >= len(sorted) {
		return sorted[len(sorted)-1], nil
	}

	return sorted[lower] + frac*(sorted[lower+1]-sorted[lower]), nil
}

// Covariance returns the population covariance of two data sets.
func Covariance(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, ErrMismatchedLengths
	}

	if len(x) == 0 {
		return 0, ErrEmptyData
	}

	mx, _ := Mean(x)
	my, _ := Mean(y)

	var sum float64

	for i := range x {
		sum += (x[i] - mx) * (y[i] - my)
	}

	return sum / float64(len(x)), nil
}

// SampleVariance returns the sample variance (Bessel-corrected) of the data.
func SampleVariance(data []float64) (float64, error) {
	if len(data) < 2 {
		return 0, ErrInsufficientData
	}

	m, _ := Mean(data)

	var sum float64

	for _, v := range data {
		d := v - m
		sum += d * d
	}

	return sum / float64(len(data)-1), nil
}

// SampleStdDev returns the sample standard deviation (Bessel-corrected) of the data.
func SampleStdDev(data []float64) (float64, error) {
	v, err := SampleVariance(data)
	if err != nil {
		return 0, err
	}

	return math.Sqrt(v), nil
}

// SampleCovariance returns the sample covariance (Bessel-corrected) of two data sets.
func SampleCovariance(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, ErrMismatchedLengths
	}

	if len(x) < 2 {
		return 0, ErrInsufficientData
	}

	mx, _ := Mean(x)
	my, _ := Mean(y)

	var sum float64

	for i := range x {
		sum += (x[i] - mx) * (y[i] - my)
	}

	return sum / float64(len(x)-1), nil
}

// LinearRegression returns the slope and intercept of the least-squares
// regression line y = slope*x + intercept.
func LinearRegression(x, y []float64) (slope, intercept float64, err error) {
	cov, err := Covariance(x, y)
	if err != nil {
		return 0, 0, err
	}

	vx, _ := Variance(x)

	if vx == 0 {
		return 0, 0, ErrZeroVariance
	}

	slope = cov / vx

	mx, _ := Mean(x)
	my, _ := Mean(y)

	intercept = my - slope*mx

	return slope, intercept, nil
}

// RSquared returns the coefficient of determination (R²) for two data sets.
func RSquared(x, y []float64) (float64, error) {
	r, err := Correlation(x, y)
	if err != nil {
		return 0, err
	}

	return r * r, nil
}

// WeightedMean returns the weighted arithmetic mean of the data.
func WeightedMean(data, weights []float64) (float64, error) {
	if len(data) != len(weights) {
		return 0, ErrMismatchedLengths
	}

	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var sumW, sumWX float64

	for i := range data {
		sumW += weights[i]
		sumWX += weights[i] * data[i]
	}

	if sumW == 0 {
		return 0, ErrInvalidWeights
	}

	return sumWX / sumW, nil
}

// WeightedVariance returns the weighted population variance.
func WeightedVariance(data, weights []float64) (float64, error) {
	wm, err := WeightedMean(data, weights)
	if err != nil {
		return 0, err
	}

	var sumW, sumWD float64

	for i := range data {
		sumW += weights[i]
		d := data[i] - wm
		sumWD += weights[i] * d * d
	}

	if sumW == 0 {
		return 0, ErrInvalidWeights
	}

	return sumWD / sumW, nil
}

// IQR returns the interquartile range (Q3 - Q1) of the data.
func IQR(data []float64) (float64, error) {
	q1, err := Percentile(data, 25)
	if err != nil {
		return 0, err
	}

	q3, err := Percentile(data, 75)
	if err != nil {
		return 0, err
	}

	return q3 - q1, nil
}

// Skewness returns the population skewness of the data.
func Skewness(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	m, _ := Mean(data)
	sd, _ := StdDev(data)

	if sd == 0 {
		return 0, nil
	}

	var sum float64

	for _, v := range data {
		d := (v - m) / sd
		sum += d * d * d
	}

	return sum / float64(len(data)), nil
}

// Kurtosis returns the population excess kurtosis of the data.
func Kurtosis(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	m, _ := Mean(data)
	sd, _ := StdDev(data)

	if sd == 0 {
		return 0, nil
	}

	var sum float64

	for _, v := range data {
		d := (v - m) / sd
		sum += d * d * d * d
	}

	return sum/float64(len(data)) - 3, nil
}

// MAD returns the median absolute deviation of the data.
func MAD(data []float64) (float64, error) {
	med, err := Median(data)
	if err != nil {
		return 0, err
	}

	deviations := make([]float64, len(data))

	for i, v := range data {
		deviations[i] = math.Abs(v - med)
	}

	return Median(deviations)
}

// GeometricMean returns the geometric mean of the data.
// All values must be positive.
func GeometricMean(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var sumLog float64

	for _, v := range data {
		if v <= 0 {
			return 0, ErrNonPositiveData
		}

		sumLog += math.Log(v)
	}

	return math.Exp(sumLog / float64(len(data))), nil
}

// HarmonicMean returns the harmonic mean of the data.
// All values must be positive.
func HarmonicMean(data []float64) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var sumRecip float64

	for _, v := range data {
		if v <= 0 {
			return 0, ErrNonPositiveData
		}

		sumRecip += 1 / v
	}

	return float64(len(data)) / sumRecip, nil
}

// Correlation returns the Pearson correlation coefficient of two data sets.
func Correlation(x, y []float64) (float64, error) {
	cov, err := Covariance(x, y)
	if err != nil {
		return 0, err
	}

	sx, _ := StdDev(x)
	sy, _ := StdDev(y)

	if sx == 0 || sy == 0 {
		return 0, nil
	}

	return cov / (sx * sy), nil
}
