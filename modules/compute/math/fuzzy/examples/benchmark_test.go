package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
)

func BenchmarkFuzzify_Triangular(b *testing.B) {
	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()

	for b.Loop() {
		fuzzy.Fuzzify(fn, 3.7)
	}
}

func BenchmarkFuzzify_Gaussian(b *testing.B) {
	fn, err := fuzzy.Gaussian(5, 1)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()

	for b.Loop() {
		fuzzy.Fuzzify(fn, 3.7)
	}
}

func BenchmarkSample(b *testing.B) {
	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()

	for b.Loop() {
		_, _, _ = fuzzy.Sample(fn, 0, 10, 100)
	}
}

func BenchmarkCentroid(b *testing.B) {
	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	xs, ys, _ := fuzzy.Sample(fn, 0, 10, 100)

	b.ResetTimer()

	for b.Loop() {
		fuzzy.Centroid(xs, ys)
	}
}

func BenchmarkAggregateMax(b *testing.B) {
	low, err := fuzzy.Triangular(0, 0, 5)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	med, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	high, err := fuzzy.Triangular(5, 10, 10)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	agg := fuzzy.AggregateMax(low, med, high)

	b.ResetTimer()

	for b.Loop() {
		fuzzy.Fuzzify(agg, 3.7)
	}
}
