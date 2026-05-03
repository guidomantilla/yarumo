package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
)

const epsilon = 1e-9

func approx(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestExample_membership(t *testing.T) {
	t.Parallel()

	t.Run("triangular", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Peak at center.
		d := fuzzy.Fuzzify(fn, 5)

		if !approx(float64(d), 1.0) {
			t.Fatalf("expected 1.0 at peak, got %v", d)
		}

		// Zero at left edge.
		d = fuzzy.Fuzzify(fn, 0)

		if !approx(float64(d), 0.0) {
			t.Fatalf("expected 0.0 at left edge, got %v", d)
		}
	})

	t.Run("trapezoidal", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Trapezoidal(0, 2, 8, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Plateau between b and c.
		d := fuzzy.Fuzzify(fn, 5)

		if !approx(float64(d), 1.0) {
			t.Fatalf("expected 1.0 in plateau, got %v", d)
		}
	})

	t.Run("gaussian", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Gaussian(5, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		d := fuzzy.Fuzzify(fn, 5)

		if !approx(float64(d), 1.0) {
			t.Fatalf("expected 1.0 at center, got %v", d)
		}
	})

	t.Run("sigmoid", func(t *testing.T) {
		t.Parallel()

		fn := fuzzy.Sigmoid(5, 2)

		d := fuzzy.Fuzzify(fn, 5)

		if !approx(float64(d), 0.5) {
			t.Fatalf("expected 0.5 at center, got %v", d)
		}
	})

	t.Run("constant", func(t *testing.T) {
		t.Parallel()

		fn := fuzzy.Constant(0.7)

		d := fuzzy.Fuzzify(fn, 42)

		if !approx(float64(d), 0.7) {
			t.Fatalf("expected 0.7, got %v", d)
		}
	})
}

func TestExample_clipAndScale(t *testing.T) {
	t.Parallel()

	t.Run("clip", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		clipped := fuzzy.Clip(fn, 0.5)

		// At peak, clipped to 0.5.
		d := fuzzy.Fuzzify(clipped, 5)

		if !approx(float64(d), 0.5) {
			t.Fatalf("expected 0.5, got %v", d)
		}
	})

	t.Run("scale", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scaled := fuzzy.Scale(fn, 0.5)

		// At peak, scaled to 0.5.
		d := fuzzy.Fuzzify(scaled, 5)

		if !approx(float64(d), 0.5) {
			t.Fatalf("expected 0.5, got %v", d)
		}
	})
}

func TestExample_aggregateMax(t *testing.T) {
	t.Parallel()

	low, err := fuzzy.Triangular(0, 0, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	med, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	high, err := fuzzy.Triangular(5, 10, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	agg := fuzzy.AggregateMax(low, med, high)

	// At 5, med peaks at 1.0.
	d := fuzzy.Fuzzify(agg, 5)

	if !approx(float64(d), 1.0) {
		t.Fatalf("expected 1.0, got %v", d)
	}
}

func TestExample_sample(t *testing.T) {
	t.Parallel()

	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xs, ys, err := fuzzy.Sample(fn, 0, 10, 5)
	if err != nil {
		t.Fatalf("sample error: %v", err)
	}

	if len(xs) != 5 {
		t.Fatalf("expected 5 x-values, got %d", len(xs))
	}

	if len(ys) != 5 {
		t.Fatalf("expected 5 degrees, got %d", len(ys))
	}
}

func TestExample_norms(t *testing.T) {
	t.Parallel()

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.4)

	t.Run("min", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Min(a, b)

		if !approx(float64(result), 0.4) {
			t.Fatalf("expected 0.4, got %v", result)
		}
	})

	t.Run("product", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Product(a, b)

		if !approx(float64(result), 0.28) {
			t.Fatalf("expected 0.28, got %v", result)
		}
	})

	t.Run("lukasiewicz", func(t *testing.T) {
		t.Parallel()

		// max(0.7 + 0.4 - 1, 0) = 0.1.
		result := fuzzy.Lukasiewicz(a, b)

		if !approx(float64(result), 0.1) {
			t.Fatalf("expected 0.1, got %v", result)
		}
	})
}

func TestExample_conorms(t *testing.T) {
	t.Parallel()

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.4)

	t.Run("max", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Max(a, b)

		if !approx(float64(result), 0.7) {
			t.Fatalf("expected 0.7, got %v", result)
		}
	})

	t.Run("probabilistic sum", func(t *testing.T) {
		t.Parallel()

		// 0.7 + 0.4 - 0.7*0.4 = 0.82.
		result := fuzzy.ProbabilisticSum(a, b)

		if !approx(float64(result), 0.82) {
			t.Fatalf("expected 0.82, got %v", result)
		}
	})

	t.Run("bounded sum", func(t *testing.T) {
		t.Parallel()

		// min(0.7 + 0.4, 1) = 1.0.
		result := fuzzy.BoundedSum(a, b)

		if !approx(float64(result), 1.0) {
			t.Fatalf("expected 1.0, got %v", result)
		}
	})
}

func TestExample_complement(t *testing.T) {
	t.Parallel()

	d := fuzzy.Degree(0.7)
	c := fuzzy.Complement(d)

	if !approx(float64(c), 0.3) {
		t.Fatalf("expected 0.3, got %v", c)
	}
}

func TestExample_defuzzify(t *testing.T) {
	t.Parallel()

	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xs, ys, err := fuzzy.Sample(fn, 0, 10, 1000)
	if err != nil {
		t.Fatalf("sample error: %v", err)
	}

	t.Run("centroid", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Centroid(xs, ys)

		// Symmetric triangle: centroid at center.
		if math.Abs(result-5.0) > 0.01 {
			t.Fatalf("expected centroid ~5.0, got %v", result)
		}
	})

	t.Run("bisector", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.Bisector(xs, ys)

		// Symmetric triangle: bisector near center.
		if math.Abs(result-5.0) > 0.1 {
			t.Fatalf("expected bisector ~5.0, got %v", result)
		}
	})

	t.Run("mean of max", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.MeanOfMax(xs, ys)

		// Symmetric triangle: single max at center.
		if math.Abs(result-5.0) > 0.02 {
			t.Fatalf("expected mean of max ~5.0, got %v", result)
		}
	})

	t.Run("smallest of max", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.SmallestOfMax(xs, ys)

		if math.Abs(result-5.0) > 0.02 {
			t.Fatalf("expected smallest of max ~5.0, got %v", result)
		}
	})

	t.Run("largest of max", func(t *testing.T) {
		t.Parallel()

		result := fuzzy.LargestOfMax(xs, ys)

		if math.Abs(result-5.0) > 0.02 {
			t.Fatalf("expected largest of max ~5.0, got %v", result)
		}
	})
}

func TestExample_set(t *testing.T) {
	t.Parallel()

	fn, err := fuzzy.Trapezoidal(15, 20, 25, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := fuzzy.Set{
		Name: "warm",
		Fn:   fn,
	}

	str := s.String()

	if str != "Set(warm)" {
		t.Fatalf("expected Set(warm), got %s", str)
	}

	d := fuzzy.Fuzzify(s.Fn, 22)

	if !approx(float64(d), 1.0) {
		t.Fatalf("expected 1.0 in plateau, got %v", d)
	}
}
