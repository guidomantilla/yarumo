package fuzzy

import (
	"math"
	"testing"
)

func TestCentroid_basic(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0, 0, 1, 0, 0}

	result := Centroid(xs, ys)

	if result != 3.0 {
		t.Fatalf("expected 3.0, got %f", result)
	}
}

func TestCentroid_uniform(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3}
	ys := []Degree{1, 1, 1}

	result := Centroid(xs, ys)

	if math.Abs(result-2.0) > 1e-9 {
		t.Fatalf("expected 2.0, got %f", result)
	}
}

func TestCentroid_zeroArea(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3}
	ys := []Degree{0, 0, 0}

	result := Centroid(xs, ys)

	if result != 0 {
		t.Fatalf("expected 0 for zero area, got %f", result)
	}
}

func TestCentroid_empty(t *testing.T) {
	t.Parallel()

	result := Centroid(nil, nil)

	if result != 0 {
		t.Fatalf("expected 0 for empty, got %f", result)
	}
}

func TestBisector_basic(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0.2, 0.2, 0.2, 0.2, 0.2}

	result := Bisector(xs, ys)

	// Total = 1.0, half = 0.5. Running sum: 0.2, 0.4, 0.6 -> bisector at x=3.
	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestBisector_leftHeavy(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3}
	ys := []Degree{1, 0, 0}

	result := Bisector(xs, ys)

	if result != 1 {
		t.Fatalf("expected 1 (left heavy), got %f", result)
	}
}

func TestBisector_empty(t *testing.T) {
	t.Parallel()

	result := Bisector(nil, nil)

	if result != 0 {
		t.Fatalf("expected 0 for empty, got %f", result)
	}
}

func TestMeanOfMax_singleMax(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0, 0, 1, 0, 0}

	result := MeanOfMax(xs, ys)

	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestMeanOfMax_multipleMax(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0, 1, 0, 1, 0}

	result := MeanOfMax(xs, ys)

	// Mean of {2, 4} = 3.
	if result != 3 {
		t.Fatalf("expected 3, got %f", result)
	}
}

func TestMeanOfMax_empty(t *testing.T) {
	t.Parallel()

	result := MeanOfMax(nil, nil)

	if result != 0 {
		t.Fatalf("expected 0 for empty, got %f", result)
	}
}

func TestLargestOfMax_basic(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0, 1, 0, 1, 0}

	result := LargestOfMax(xs, ys)

	if result != 4 {
		t.Fatalf("expected 4 (largest of max), got %f", result)
	}
}

func TestLargestOfMax_empty(t *testing.T) {
	t.Parallel()

	result := LargestOfMax(nil, nil)

	if result != 0 {
		t.Fatalf("expected 0 for empty, got %f", result)
	}
}

func TestLargestOfMax_single(t *testing.T) {
	t.Parallel()

	xs := []float64{5}
	ys := []Degree{0.7}

	result := LargestOfMax(xs, ys)

	if result != 5 {
		t.Fatalf("expected 5, got %f", result)
	}
}

func TestSmallestOfMax_basic(t *testing.T) {
	t.Parallel()

	xs := []float64{1, 2, 3, 4, 5}
	ys := []Degree{0, 1, 0, 1, 0}

	result := SmallestOfMax(xs, ys)

	if result != 2 {
		t.Fatalf("expected 2 (smallest of max), got %f", result)
	}
}

func TestSmallestOfMax_empty(t *testing.T) {
	t.Parallel()

	result := SmallestOfMax(nil, nil)

	if result != 0 {
		t.Fatalf("expected 0 for empty, got %f", result)
	}
}

func TestSmallestOfMax_single(t *testing.T) {
	t.Parallel()

	xs := []float64{5}
	ys := []Degree{0.7}

	result := SmallestOfMax(xs, ys)

	if result != 5 {
		t.Fatalf("expected 5, got %f", result)
	}
}
