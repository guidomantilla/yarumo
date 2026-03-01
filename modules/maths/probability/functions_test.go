package probability

import (
	"errors"
	"math"
	"testing"
)

func TestIsValid_validDistribution(t *testing.T) {
	t.Parallel()

	d := Distribution{"heads": 0.5, "tails": 0.5}

	if !IsValid(d) {
		t.Fatal("expected valid distribution")
	}
}

func TestIsValid_emptyDistribution(t *testing.T) {
	t.Parallel()

	if IsValid(Distribution{}) {
		t.Fatal("expected invalid for empty distribution")
	}
}

func TestIsValid_notNormalized(t *testing.T) {
	t.Parallel()

	d := Distribution{"heads": 0.3, "tails": 0.3}

	if IsValid(d) {
		t.Fatal("expected invalid for non-normalized distribution")
	}
}

func TestIsValid_negativeProb(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": -0.5, "b": 1.5}

	if IsValid(d) {
		t.Fatal("expected invalid for negative probability")
	}
}

func TestIsValid_probGreaterThanOne(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 1.5}

	if IsValid(d) {
		t.Fatal("expected invalid for probability > 1")
	}
}

func TestNormalize_valid(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 2, "b": 3}

	result, err := Normalize(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(float64(result["a"])-0.4) > epsilon {
		t.Fatalf("expected 0.4 for a, got %f", float64(result["a"]))
	}

	if math.Abs(float64(result["b"])-0.6) > epsilon {
		t.Fatalf("expected 0.6 for b, got %f", float64(result["b"]))
	}
}

func TestNormalize_empty(t *testing.T) {
	t.Parallel()

	_, err := Normalize(Distribution{})
	if !errors.Is(err, ErrEmptyDist) {
		t.Fatalf("expected ErrEmptyDist, got %v", err)
	}
}

func TestNormalize_zeroSum(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 0, "b": 0}

	_, err := Normalize(d)
	if !errors.Is(err, ErrNotNormalized) {
		t.Fatalf("expected ErrNotNormalized, got %v", err)
	}
}

func TestComplement(t *testing.T) {
	t.Parallel()

	result := Complement(0.3)
	if math.Abs(float64(result)-0.7) > epsilon {
		t.Fatalf("expected 0.7, got %f", float64(result))
	}
}

func TestComplement_zero(t *testing.T) {
	t.Parallel()

	result := Complement(0)
	if float64(result) != 1.0 {
		t.Fatalf("expected 1.0, got %f", float64(result))
	}
}

func TestComplement_one(t *testing.T) {
	t.Parallel()

	result := Complement(1)
	if float64(result) != 0.0 {
		t.Fatalf("expected 0.0, got %f", float64(result))
	}
}

func TestEntropy_uniform(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 0.5, "b": 0.5}

	h, err := Entropy(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(h-1.0) > epsilon {
		t.Fatalf("expected entropy 1.0 for fair coin, got %f", h)
	}
}

func TestEntropy_certain(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 1.0, "b": 0.0}

	h, err := Entropy(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(h) > epsilon {
		t.Fatalf("expected entropy 0.0 for certain outcome, got %f", h)
	}
}

func TestEntropy_empty(t *testing.T) {
	t.Parallel()

	_, err := Entropy(Distribution{})
	if !errors.Is(err, ErrEmptyDist) {
		t.Fatalf("expected ErrEmptyDist, got %v", err)
	}
}

func TestEntropy_invalidProb(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": -0.5}

	_, err := Entropy(d)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

func TestEntropy_invalidProbGreaterThanOne(t *testing.T) {
	t.Parallel()

	d := Distribution{"a": 1.5}

	_, err := Entropy(d)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}
