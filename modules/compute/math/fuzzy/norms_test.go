package fuzzy

import (
	"math"
	"testing"
)

func TestMin_aSmaller(t *testing.T) {
	t.Parallel()

	if Min(0.3, 0.7) != 0.3 {
		t.Fatalf("expected 0.3, got %f", float64(Min(0.3, 0.7)))
	}
}

func TestMin_bSmaller(t *testing.T) {
	t.Parallel()

	if Min(0.8, 0.2) != 0.2 {
		t.Fatalf("expected 0.2, got %f", float64(Min(0.8, 0.2)))
	}
}

func TestMin_equal(t *testing.T) {
	t.Parallel()

	if Min(0.5, 0.5) != 0.5 {
		t.Fatalf("expected 0.5, got %f", float64(Min(0.5, 0.5)))
	}
}

func TestProduct(t *testing.T) {
	t.Parallel()

	result := Product(0.6, 0.5)

	if math.Abs(float64(result)-0.3) > 1e-9 {
		t.Fatalf("expected 0.3, got %f", float64(result))
	}
}

func TestLukasiewicz_positive(t *testing.T) {
	t.Parallel()

	result := Lukasiewicz(0.8, 0.6)
	expected := 0.8 + 0.6 - 1.0

	if math.Abs(float64(result)-expected) > 1e-9 {
		t.Fatalf("expected %f, got %f", expected, float64(result))
	}
}

func TestLukasiewicz_negative(t *testing.T) {
	t.Parallel()

	result := Lukasiewicz(0.3, 0.4)

	if result != 0 {
		t.Fatalf("expected 0, got %f", float64(result))
	}
}

func TestMax_aLarger(t *testing.T) {
	t.Parallel()

	if Max(0.7, 0.3) != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(Max(0.7, 0.3)))
	}
}

func TestMax_bLarger(t *testing.T) {
	t.Parallel()

	if Max(0.2, 0.8) != 0.8 {
		t.Fatalf("expected 0.8, got %f", float64(Max(0.2, 0.8)))
	}
}

func TestMax_equal(t *testing.T) {
	t.Parallel()

	if Max(0.5, 0.5) != 0.5 {
		t.Fatalf("expected 0.5, got %f", float64(Max(0.5, 0.5)))
	}
}

func TestProbabilisticSum(t *testing.T) {
	t.Parallel()

	result := ProbabilisticSum(0.3, 0.4)
	expected := 0.3 + 0.4 - 0.3*0.4

	if math.Abs(float64(result)-expected) > 1e-9 {
		t.Fatalf("expected %f, got %f", expected, float64(result))
	}
}

func TestBoundedSum_underOne(t *testing.T) {
	t.Parallel()

	result := BoundedSum(0.3, 0.4)

	if math.Abs(float64(result)-0.7) > 1e-9 {
		t.Fatalf("expected 0.7, got %f", float64(result))
	}
}

func TestBoundedSum_overOne(t *testing.T) {
	t.Parallel()

	result := BoundedSum(0.7, 0.8)

	if result != 1.0 {
		t.Fatalf("expected 1.0, got %f", float64(result))
	}
}

func TestComplement(t *testing.T) {
	t.Parallel()

	result := Complement(0.3)

	if math.Abs(float64(result)-0.7) > 1e-9 {
		t.Fatalf("expected 0.7, got %f", float64(result))
	}
}

func TestComplement_zero(t *testing.T) {
	t.Parallel()

	if Complement(0) != 1.0 {
		t.Fatalf("expected 1.0, got %f", float64(Complement(0)))
	}
}

func TestComplement_one(t *testing.T) {
	t.Parallel()

	if Complement(1) != 0 {
		t.Fatalf("expected 0, got %f", float64(Complement(1)))
	}
}
