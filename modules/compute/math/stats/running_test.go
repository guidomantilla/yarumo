package stats

import (
	"math"
	"testing"
)

// --- RunningStats ---

func TestRunningStats_empty(t *testing.T) {
	t.Parallel()

	var rs RunningStats

	if rs.Count() != 0 {
		t.Fatalf("expected Count 0, got %d", rs.Count())
	}

	if rs.Mean() != 0 {
		t.Fatalf("expected Mean 0, got %f", rs.Mean())
	}

	if rs.Variance() != 0 {
		t.Fatalf("expected Variance 0, got %f", rs.Variance())
	}

	if rs.SampleVariance() != 0 {
		t.Fatalf("expected SampleVariance 0, got %f", rs.SampleVariance())
	}
}

func TestRunningStats_singleValue(t *testing.T) {
	t.Parallel()

	var rs RunningStats
	rs.Push(5)

	if rs.Count() != 1 {
		t.Fatalf("expected Count 1, got %d", rs.Count())
	}

	if rs.Mean() != 5 {
		t.Fatalf("expected Mean 5, got %f", rs.Mean())
	}

	if rs.Variance() != 0 {
		t.Fatalf("expected Variance 0, got %f", rs.Variance())
	}

	if rs.SampleVariance() != 0 {
		t.Fatalf("expected SampleVariance 0, got %f", rs.SampleVariance())
	}
}

func TestRunningStats_multipleValues(t *testing.T) {
	t.Parallel()

	var rs RunningStats

	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}

	for _, v := range data {
		rs.Push(v)
	}

	if rs.Count() != 8 {
		t.Fatalf("expected Count 8, got %d", rs.Count())
	}

	if math.Abs(rs.Mean()-5.0) > 1e-9 {
		t.Fatalf("expected Mean 5.0, got %f", rs.Mean())
	}

	if math.Abs(rs.Variance()-4.0) > 1e-9 {
		t.Fatalf("expected Variance 4.0, got %f", rs.Variance())
	}

	expectedSampleVar := 32.0 / 7.0

	if math.Abs(rs.SampleVariance()-expectedSampleVar) > 1e-9 {
		t.Fatalf("expected SampleVariance %f, got %f", expectedSampleVar, rs.SampleVariance())
	}
}

func TestRunningStats_matchesBatch(t *testing.T) {
	t.Parallel()

	data := []float64{1, 3, 5, 7, 9, 11, 13}

	var rs RunningStats

	for _, v := range data {
		rs.Push(v)
	}

	batchMean, err := Mean(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batchVar, err := Variance(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(rs.Mean()-batchMean) > 1e-9 {
		t.Fatalf("Mean mismatch: running %f, batch %f", rs.Mean(), batchMean)
	}

	if math.Abs(rs.Variance()-batchVar) > 1e-9 {
		t.Fatalf("Variance mismatch: running %f, batch %f", rs.Variance(), batchVar)
	}
}

func TestRunningStats_incrementalCorrectness(t *testing.T) {
	t.Parallel()

	var rs RunningStats

	data := []float64{10, 20, 30, 40, 50}
	expectedMeans := []float64{10, 15, 20, 25, 30}

	for i, v := range data {
		rs.Push(v)

		if math.Abs(rs.Mean()-expectedMeans[i]) > 1e-9 {
			t.Fatalf("after Push(%f): expected Mean %f, got %f", v, expectedMeans[i], rs.Mean())
		}
	}
}
