package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// TestProperties_kolmogorov verifies Kolmogorov axioms of probability.
func TestProperties_kolmogorov(t *testing.T) {
	t.Parallel()

	t.Run("probabilities are non-negative", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 0.3, "B": 0.5, "C": 0.2}

		for outcome, p := range dist {
			if p < 0 {
				t.Fatalf("negative probability for %s: %v", outcome, p)
			}
		}
	})

	t.Run("valid distribution sums to one", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 0.3, "B": 0.5, "C": 0.2}

		if !stats.IsValid(dist) {
			t.Fatal("expected valid distribution")
		}
	})

	t.Run("complement sums to one", func(t *testing.T) {
		t.Parallel()

		p := stats.Prob(0.7)
		sum := float64(p) + float64(stats.Complement(p))

		if !approx(sum, 1.0) {
			t.Fatalf("expected sum 1.0, got %v", sum)
		}
	})

	t.Run("normalization preserves ratios", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 2, "B": 4, "C": 4}

		normalized, err := stats.Normalize(dist)
		if err != nil {
			t.Fatalf("normalize error: %v", err)
		}

		ratio := float64(normalized["B"]) / float64(normalized["A"])

		if !approx(ratio, 2.0) {
			t.Fatalf("expected ratio 2.0, got %v", ratio)
		}
	})
}

// TestProperties_bayes verifies Bayes theorem properties.
func TestProperties_bayes(t *testing.T) {
	t.Parallel()

	t.Run("posterior equals prior when likelihood equals evidence", func(t *testing.T) {
		t.Parallel()

		prior := stats.Prob(0.3)
		same := stats.Prob(0.5)

		posterior, err := stats.Bayes(prior, same, same)
		if err != nil {
			t.Fatalf("bayes error: %v", err)
		}

		if !approx(float64(posterior), float64(prior)) {
			t.Fatalf("expected %v, got %v", prior, posterior)
		}
	})

	t.Run("total probability matches manual calculation", func(t *testing.T) {
		t.Parallel()

		priors := []stats.Prob{0.4, 0.6}
		likelihoods := []stats.Prob{0.8, 0.2}

		total, err := stats.TotalProbability(priors, likelihoods)
		if err != nil {
			t.Fatalf("total probability error: %v", err)
		}

		manual := 0.8*0.4 + 0.2*0.6

		if !approx(float64(total), manual) {
			t.Fatalf("expected %v, got %v", manual, total)
		}
	})

	t.Run("chain rule with certainties equals first factor", func(t *testing.T) {
		t.Parallel()

		result := stats.ChainRule(0.5, 1.0, 1.0)

		if !approx(float64(result), 0.5) {
			t.Fatalf("expected 0.5, got %v", result)
		}
	})

	t.Run("independence is commutative", func(t *testing.T) {
		t.Parallel()

		ab := stats.Independent(0.3, 0.7)
		ba := stats.Independent(0.7, 0.3)

		if !approx(float64(ab), float64(ba)) {
			t.Fatalf("expected commutative: %v != %v", ab, ba)
		}
	})
}

// TestProperties_entropy verifies Shannon entropy properties.
func TestProperties_entropy(t *testing.T) {
	t.Parallel()

	t.Run("certain outcome has zero entropy", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 1.0}

		h, err := stats.Entropy(dist)
		if err != nil {
			t.Fatalf("entropy error: %v", err)
		}

		if !approx(h, 0.0) {
			t.Fatalf("expected 0.0, got %v", h)
		}
	})

	t.Run("uniform distribution has maximum entropy", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 0.25, "B": 0.25, "C": 0.25, "D": 0.25}

		h, err := stats.Entropy(dist)
		if err != nil {
			t.Fatalf("entropy error: %v", err)
		}

		maxEntropy := math.Log2(4)

		if !approx(h, maxEntropy) {
			t.Fatalf("expected %v, got %v", maxEntropy, h)
		}
	})

	t.Run("entropy is non-negative", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 0.9, "B": 0.1}

		h, err := stats.Entropy(dist)
		if err != nil {
			t.Fatalf("entropy error: %v", err)
		}

		if h < 0 {
			t.Fatalf("expected non-negative entropy, got %v", h)
		}
	})

	t.Run("adding zero-weight outcome does not change entropy", func(t *testing.T) {
		t.Parallel()

		dist := stats.Distribution{"A": 0.5, "B": 0.5}

		h, err := stats.Entropy(dist)
		if err != nil {
			t.Fatalf("entropy error: %v", err)
		}

		if !approx(h, 1.0) {
			t.Fatalf("expected 1.0, got %v", h)
		}
	})
}
