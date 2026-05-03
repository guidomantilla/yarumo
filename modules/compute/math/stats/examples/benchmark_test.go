package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

func BenchmarkBayes(b *testing.B) {
	for b.Loop() {
		_, _ = stats.Bayes(0.01, 0.9, 0.059)
	}
}

func BenchmarkEntropy(b *testing.B) {
	dist := stats.Distribution{"A": 0.25, "B": 0.25, "C": 0.25, "D": 0.25}

	b.ResetTimer()

	for b.Loop() {
		_, _ = stats.Entropy(dist)
	}
}

func BenchmarkNormalize(b *testing.B) {
	dist := stats.Distribution{"A": 2, "B": 3, "C": 5, "D": 7, "E": 1}

	b.ResetTimer()

	for b.Loop() {
		_, _ = stats.Normalize(dist)
	}
}
