package stats

import (
	"math"
	"testing"
)

// --- NewWindowedStats ---

func TestNewWindowedStats_valid(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(5)
	if ws == nil {
		t.Fatalf("expected non-nil WindowedStats")
	}
}

func TestNewWindowedStats_zeroSize(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(0)
	if ws != nil {
		t.Fatalf("expected nil for zero size, got %v", ws)
	}
}

func TestNewWindowedStats_negativeSize(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(-1)
	if ws != nil {
		t.Fatalf("expected nil for negative size, got %v", ws)
	}
}

// --- WindowedStats ---

func TestWindowedStats_empty(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(5)

	if ws.Count() != 0 {
		t.Fatalf("expected Count 0, got %d", ws.Count())
	}

	if ws.Mean() != 0 {
		t.Fatalf("expected Mean 0, got %f", ws.Mean())
	}

	if ws.Variance() != 0 {
		t.Fatalf("expected Variance 0, got %f", ws.Variance())
	}
}

func TestWindowedStats_singleValue(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(5)
	ws.Push(7)

	if ws.Count() != 1 {
		t.Fatalf("expected Count 1, got %d", ws.Count())
	}

	if ws.Mean() != 7 {
		t.Fatalf("expected Mean 7, got %f", ws.Mean())
	}

	if ws.Variance() != 0 {
		t.Fatalf("expected Variance 0, got %f", ws.Variance())
	}
}

func TestWindowedStats_fillWindow(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(5)

	for _, v := range []float64{1, 2, 3, 4, 5} {
		ws.Push(v)
	}

	if ws.Count() != 5 {
		t.Fatalf("expected Count 5, got %d", ws.Count())
	}

	if math.Abs(ws.Mean()-3.0) > 1e-9 {
		t.Fatalf("expected Mean 3.0, got %f", ws.Mean())
	}
}

func TestWindowedStats_overflow(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(3)

	for _, v := range []float64{1, 2, 3, 4, 5, 6} {
		ws.Push(v)
	}

	// Window should contain {4, 5, 6}.
	if ws.Count() != 3 {
		t.Fatalf("expected Count 3, got %d", ws.Count())
	}

	if math.Abs(ws.Mean()-5.0) > 1e-9 {
		t.Fatalf("expected Mean 5.0, got %f", ws.Mean())
	}
}

func TestWindowedStats_minMax(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(10)

	for _, v := range []float64{3, 1, 4, 1, 5} {
		ws.Push(v)
	}

	if ws.Min() != 1 {
		t.Fatalf("expected Min 1, got %f", ws.Min())
	}

	if ws.Max() != 5 {
		t.Fatalf("expected Max 5, got %f", ws.Max())
	}
}

func TestWindowedStats_minMaxAfterOverflow(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(3)

	for _, v := range []float64{10, 1, 5, 8, 3} {
		ws.Push(v)
	}

	// Window should contain {5, 8, 3}.
	if ws.Min() != 3 {
		t.Fatalf("expected Min 3, got %f", ws.Min())
	}

	if ws.Max() != 8 {
		t.Fatalf("expected Max 8, got %f", ws.Max())
	}
}

func TestWindowedStats_variance(t *testing.T) {
	t.Parallel()

	ws := NewWindowedStats(8)

	for _, v := range []float64{2, 4, 4, 4, 5, 5, 7, 9} {
		ws.Push(v)
	}

	if math.Abs(ws.Variance()-4.0) > 1e-9 {
		t.Fatalf("expected Variance 4.0, got %f", ws.Variance())
	}
}
