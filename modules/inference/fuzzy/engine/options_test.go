package engine

import (
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

func TestNewOptions_defaults(t *testing.T) {
	t.Parallel()

	o := NewOptions()

	if o.method != Mamdani {
		t.Fatalf("expected Mamdani, got %d", o.method)
	}

	if o.tnorm == nil {
		t.Fatal("expected non-nil tnorm")
	}

	if o.tconorm == nil {
		t.Fatal("expected non-nil tconorm")
	}

	if o.defuzzify == nil {
		t.Fatal("expected non-nil defuzzify")
	}

	if o.resolution != defaultResolution {
		t.Fatalf("expected %d, got %d", defaultResolution, o.resolution)
	}
}

func TestWithMethod_mamdani(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithMethod(Mamdani))

	if o.method != Mamdani {
		t.Fatalf("expected Mamdani, got %d", o.method)
	}
}

func TestWithMethod_sugeno(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithMethod(Sugeno))

	if o.method != Sugeno {
		t.Fatalf("expected Sugeno, got %d", o.method)
	}
}

func TestWithMethod_outOfRange(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithMethod(Method(99)))

	if o.method != Mamdani {
		t.Fatalf("expected default Mamdani, got %d", o.method)
	}
}

func TestWithTNorm(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithTNorm(fuzzym.Product))

	result := o.tnorm(0.5, 0.5)
	expected := fuzzym.Product(0.5, 0.5)

	if result != expected {
		t.Fatalf("expected %f, got %f", float64(expected), float64(result))
	}
}

func TestWithTNorm_nil(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithTNorm(nil))

	if o.tnorm == nil {
		t.Fatal("expected default tnorm, not nil")
	}
}

func TestWithTConorm(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithTConorm(fuzzym.ProbabilisticSum))

	result := o.tconorm(0.5, 0.5)
	expected := fuzzym.ProbabilisticSum(0.5, 0.5)

	if result != expected {
		t.Fatalf("expected %f, got %f", float64(expected), float64(result))
	}
}

func TestWithTConorm_nil(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithTConorm(nil))

	if o.tconorm == nil {
		t.Fatal("expected default tconorm, not nil")
	}
}

func TestWithDefuzzify(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithDefuzzify(fuzzym.MeanOfMax))

	if o.defuzzify == nil {
		t.Fatal("expected non-nil defuzzify")
	}
}

func TestWithDefuzzify_nil(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithDefuzzify(nil))

	if o.defuzzify == nil {
		t.Fatal("expected default defuzzify, not nil")
	}
}

func TestWithResolution(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(200))

	if o.resolution != 200 {
		t.Fatalf("expected 200, got %d", o.resolution)
	}
}

func TestWithResolution_zero(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(0))

	if o.resolution != defaultResolution {
		t.Fatalf("expected default, got %d", o.resolution)
	}
}

func TestWithResolution_negative(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithResolution(-5))

	if o.resolution != defaultResolution {
		t.Fatalf("expected default, got %d", o.resolution)
	}
}

func TestWithSugenoOutputs(t *testing.T) {
	t.Parallel()

	outputs := map[string]float64{"speed/fast": 80.0}
	o := NewOptions(WithSugenoOutputs(outputs))

	if len(o.sugenoOutputs) != 1 {
		t.Fatalf("expected 1 sugeno output, got %d", len(o.sugenoOutputs))
	}
}

func TestWithSugenoOutputs_empty(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithSugenoOutputs(nil))

	if o.sugenoOutputs != nil {
		t.Fatal("expected nil sugeno outputs")
	}
}
