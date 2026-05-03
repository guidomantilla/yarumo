package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
)

// TestProperties_tnorms verifies t-norm axioms: commutativity, identity, and annihilation.
func TestProperties_tnorms(t *testing.T) {
	t.Parallel()

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.3)

	t.Run("min commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Min(a, b)), float64(fuzzy.Min(b, a))) {
			t.Fatal("Min is not commutative")
		}
	})

	t.Run("min identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Min(a, 1.0)), float64(a)) {
			t.Fatal("Min(a, 1) should equal a")
		}
	})

	t.Run("min annihilation", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Min(a, 0.0)), 0.0) {
			t.Fatal("Min(a, 0) should equal 0")
		}
	})

	t.Run("product commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Product(a, b)), float64(fuzzy.Product(b, a))) {
			t.Fatal("Product is not commutative")
		}
	})

	t.Run("product identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Product(a, 1.0)), float64(a)) {
			t.Fatal("Product(a, 1) should equal a")
		}
	})

	t.Run("product annihilation", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Product(a, 0.0)), 0.0) {
			t.Fatal("Product(a, 0) should equal 0")
		}
	})

	t.Run("lukasiewicz commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Lukasiewicz(a, b)), float64(fuzzy.Lukasiewicz(b, a))) {
			t.Fatal("Lukasiewicz is not commutative")
		}
	})

	t.Run("lukasiewicz identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Lukasiewicz(a, 1.0)), float64(a)) {
			t.Fatal("Lukasiewicz(a, 1) should equal a")
		}
	})

	t.Run("lukasiewicz annihilation", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Lukasiewicz(a, 0.0)), 0.0) {
			t.Fatal("Lukasiewicz(a, 0) should equal 0")
		}
	})
}

// TestProperties_tconorms verifies t-conorm axioms: commutativity, identity, and annihilation.
func TestProperties_tconorms(t *testing.T) {
	t.Parallel()

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.3)

	t.Run("max commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Max(a, b)), float64(fuzzy.Max(b, a))) {
			t.Fatal("Max is not commutative")
		}
	})

	t.Run("max identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Max(a, 0.0)), float64(a)) {
			t.Fatal("Max(a, 0) should equal a")
		}
	})

	t.Run("max annihilation", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.Max(a, 1.0)), 1.0) {
			t.Fatal("Max(a, 1) should equal 1")
		}
	})

	t.Run("probabilistic sum commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.ProbabilisticSum(a, b)), float64(fuzzy.ProbabilisticSum(b, a))) {
			t.Fatal("ProbabilisticSum is not commutative")
		}
	})

	t.Run("probabilistic sum identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.ProbabilisticSum(a, 0.0)), float64(a)) {
			t.Fatal("ProbabilisticSum(a, 0) should equal a")
		}
	})

	t.Run("bounded sum commutativity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.BoundedSum(a, b)), float64(fuzzy.BoundedSum(b, a))) {
			t.Fatal("BoundedSum is not commutative")
		}
	})

	t.Run("bounded sum identity", func(t *testing.T) {
		t.Parallel()

		if !approx(float64(fuzzy.BoundedSum(a, 0.0)), float64(a)) {
			t.Fatal("BoundedSum(a, 0) should equal a")
		}
	})
}

// TestProperties_complement verifies fuzzy complement properties.
func TestProperties_complement(t *testing.T) {
	t.Parallel()

	t.Run("involution", func(t *testing.T) {
		t.Parallel()

		a := fuzzy.Degree(0.7)
		result := fuzzy.Complement(fuzzy.Complement(a))

		if !approx(float64(result), float64(a)) {
			t.Fatalf("C(C(a)) should equal a: got %v", result)
		}
	})

	t.Run("boundary zero", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Complement(0.0)

		if !approx(float64(result), 1.0) {
			t.Fatalf("C(0) should equal 1: got %v", result)
		}
	})

	t.Run("boundary one", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Complement(1.0)

		if !approx(float64(result), 0.0) {
			t.Fatalf("C(1) should equal 0: got %v", result)
		}
	})
}

// TestProperties_deMorgan verifies De Morgan laws for dual t-norm/t-conorm pairs.
func TestProperties_deMorgan(t *testing.T) {
	t.Parallel()

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.4)

	t.Run("min/max pair", func(t *testing.T) {
		t.Parallel()

		// C(Min(a,b)) = Max(C(a), C(b)).
		lhs := fuzzy.Complement(fuzzy.Min(a, b))
		rhs := fuzzy.Max(fuzzy.Complement(a), fuzzy.Complement(b))

		if !approx(float64(lhs), float64(rhs)) {
			t.Fatalf("De Morgan failed: %v != %v", lhs, rhs)
		}
	})

	t.Run("product/probabilistic sum pair", func(t *testing.T) {
		t.Parallel()

		// C(Product(a,b)) = ProbabilisticSum(C(a), C(b)).
		lhs := fuzzy.Complement(fuzzy.Product(a, b))
		rhs := fuzzy.ProbabilisticSum(fuzzy.Complement(a), fuzzy.Complement(b))

		if !approx(float64(lhs), float64(rhs)) {
			t.Fatalf("De Morgan failed: %v != %v", lhs, rhs)
		}
	})

	t.Run("lukasiewicz/bounded sum pair", func(t *testing.T) {
		t.Parallel()

		// C(Lukasiewicz(a,b)) = BoundedSum(C(a), C(b)).
		lhs := fuzzy.Complement(fuzzy.Lukasiewicz(a, b))
		rhs := fuzzy.BoundedSum(fuzzy.Complement(a), fuzzy.Complement(b))

		if !approx(float64(lhs), float64(rhs)) {
			t.Fatalf("De Morgan failed: %v != %v", lhs, rhs)
		}
	})
}

// TestProperties_membership verifies membership function mathematical properties.
func TestProperties_membership(t *testing.T) {
	t.Parallel()

	t.Run("triangular peaks at center", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		peak := fuzzy.Fuzzify(fn, 5)
		left := fuzzy.Fuzzify(fn, 3)
		right := fuzzy.Fuzzify(fn, 7)

		if float64(peak) <= float64(left) || float64(peak) <= float64(right) {
			t.Fatal("peak should be greater than surrounding points")
		}
	})

	t.Run("gaussian is symmetric", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Gaussian(5, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		left := fuzzy.Fuzzify(fn, 3)
		right := fuzzy.Fuzzify(fn, 7)

		if !approx(float64(left), float64(right)) {
			t.Fatalf("expected symmetric: %v != %v", left, right)
		}
	})

	t.Run("clip limits maximum degree", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		clipped := fuzzy.Clip(fn, 0.3)

		peak := fuzzy.Fuzzify(clipped, 5)

		if float64(peak) > 0.3+epsilon {
			t.Fatalf("clipped peak should not exceed 0.3: got %v", peak)
		}
	})

	t.Run("scale preserves shape proportionally", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scaled := fuzzy.Scale(fn, 0.5)

		original := fuzzy.Fuzzify(fn, 3)
		result := fuzzy.Fuzzify(scaled, 3)

		expected := float64(original) * 0.5

		if !approx(float64(result), expected) {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	})
}
