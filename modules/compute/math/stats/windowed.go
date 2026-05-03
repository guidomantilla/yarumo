package stats

import (
	"math"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// WindowedStats computes statistics over a sliding window using a circular buffer.
type WindowedStats struct {
	buf       []float64
	size      int
	count     int
	pos       int
	sum       float64
	min       float64
	max       float64
	pushCount int
}

// NewWindowedStats creates a WindowedStats with the given window size.
// Size must be positive; returns nil for invalid size.
func NewWindowedStats(size int) *WindowedStats {
	if size <= 0 {
		return nil
	}

	return &WindowedStats{
		buf:  make([]float64, size),
		size: size,
		min:  math.Inf(1),
		max:  math.Inf(-1),
	}
}

// Push adds a new value to the window.
func (w *WindowedStats) Push(x float64) {
	cassert.NotNil(w, "windowed stats is nil")

	if w.count >= w.size {
		w.sum -= w.buf[w.pos]
	}

	w.buf[w.pos] = x
	w.sum += x
	w.pos = (w.pos + 1) % w.size

	if w.count < w.size {
		w.count++
	}

	w.pushCount++

	// Periodically recompute sum to prevent floating-point drift.
	if w.pushCount%w.size == 0 {
		w.recomputeSum()
	}

	// Recompute min/max from buffer (simple approach for correctness).
	w.recomputeMinMax()
}

// Count returns the number of values currently in the window.
func (w *WindowedStats) Count() int {
	cassert.NotNil(w, "windowed stats is nil")

	return w.count
}

// Mean returns the mean of values in the window. Returns 0 if empty.
func (w *WindowedStats) Mean() float64 {
	cassert.NotNil(w, "windowed stats is nil")

	if w.count == 0 {
		return 0
	}

	return w.sum / float64(w.count)
}

// Variance returns the population variance of values in the window. Returns 0 if empty.
func (w *WindowedStats) Variance() float64 {
	cassert.NotNil(w, "windowed stats is nil")

	if w.count == 0 {
		return 0
	}

	mean := w.Mean()

	var sum float64

	for i := range w.count {
		idx := (w.pos - w.count + i + w.size) % w.size
		d := w.buf[idx] - mean
		sum += d * d
	}

	return sum / float64(w.count)
}

// Min returns the minimum value in the window. Returns +Inf if empty.
func (w *WindowedStats) Min() float64 {
	cassert.NotNil(w, "windowed stats is nil")

	return w.min
}

// Max returns the maximum value in the window. Returns -Inf if empty.
func (w *WindowedStats) Max() float64 {
	cassert.NotNil(w, "windowed stats is nil")

	return w.max
}

func (w *WindowedStats) recomputeSum() {
	cassert.NotNil(w, "windowed stats is nil")

	w.sum = 0

	for i := range w.count {
		idx := (w.pos - w.count + i + w.size) % w.size
		w.sum += w.buf[idx]
	}
}

func (w *WindowedStats) recomputeMinMax() {
	cassert.NotNil(w, "windowed stats is nil")

	w.min = math.Inf(1)
	w.max = math.Inf(-1)

	for i := range w.count {
		idx := (w.pos - w.count + i + w.size) % w.size

		if w.buf[idx] < w.min {
			w.min = w.buf[idx]
		}

		if w.buf[idx] > w.max {
			w.max = w.buf[idx]
		}
	}
}
