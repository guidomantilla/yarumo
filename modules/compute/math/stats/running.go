package stats

import cassert "github.com/guidomantilla/yarumo/common/assert"

// RunningStats computes mean and variance incrementally using Welford's online algorithm.
type RunningStats struct {
	n    int
	mean float64
	m2   float64
}

// Push adds a new value to the running statistics.
func (r *RunningStats) Push(x float64) {
	cassert.NotNil(r, "running stats is nil")

	r.n++

	delta := x - r.mean
	r.mean += delta / float64(r.n)

	delta2 := x - r.mean
	r.m2 += delta * delta2
}

// Count returns the number of values observed.
func (r *RunningStats) Count() int {
	cassert.NotNil(r, "running stats is nil")

	return r.n
}

// Mean returns the current mean. Returns 0 if no values have been pushed.
func (r *RunningStats) Mean() float64 {
	cassert.NotNil(r, "running stats is nil")

	return r.mean
}

// Variance returns the population variance. Returns 0 if fewer than 1 value.
func (r *RunningStats) Variance() float64 {
	cassert.NotNil(r, "running stats is nil")

	if r.n < 1 {
		return 0
	}

	return r.m2 / float64(r.n)
}

// SampleVariance returns the sample variance (Bessel-corrected). Returns 0 if fewer than 2 values.
func (r *RunningStats) SampleVariance() float64 {
	cassert.NotNil(r, "running stats is nil")

	if r.n < 2 {
		return 0
	}

	return r.m2 / float64(r.n-1)
}
