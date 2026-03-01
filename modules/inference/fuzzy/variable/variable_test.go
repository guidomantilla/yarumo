package variable

import (
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

func makeTemperatureVariable(opts ...Option) Variable {
	terms := []Term{
		{Name: "cold", Fn: fuzzym.Triangular(0, 0, 50)},
		{Name: "warm", Fn: fuzzym.Triangular(20, 50, 80)},
		{Name: "hot", Fn: fuzzym.Triangular(50, 100, 100)},
	}

	return NewVariable("temperature", 0, 100, terms, opts...)
}

func TestNewVariable(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()

	if v.Name() != "temperature" {
		t.Fatalf("expected temperature, got %s", v.Name())
	}

	if v.Min() != 0 {
		t.Fatalf("expected 0, got %f", v.Min())
	}

	if v.Max() != 100 {
		t.Fatalf("expected 100, got %f", v.Max())
	}

	if len(v.Terms()) != 3 {
		t.Fatalf("expected 3 terms, got %d", len(v.Terms()))
	}

	if v.Resolution() != defaultResolution {
		t.Fatalf("expected %d, got %d", defaultResolution, v.Resolution())
	}
}

func TestNewVariable_withResolution(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable(WithResolution(50))

	if v.Resolution() != 50 {
		t.Fatalf("expected 50, got %d", v.Resolution())
	}
}

func TestVariable_Term_found(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()
	term, ok := v.Term("cold")

	if !ok {
		t.Fatal("expected cold term found")
	}

	if term.Name != "cold" {
		t.Fatalf("expected cold, got %s", term.Name)
	}
}

func TestVariable_Term_notFound(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()
	_, ok := v.Term("freezing")

	if ok {
		t.Fatal("expected not found")
	}
}

func TestVariable_Fuzzify(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()
	degrees := v.Fuzzify(50)

	if len(degrees) != 3 {
		t.Fatalf("expected 3 degrees, got %d", len(degrees))
	}

	// At x=50: cold=0.0, warm=1.0 (peak), hot=0.0.
	if degrees["warm"] != 1.0 {
		t.Fatalf("expected warm=1.0, got %f", float64(degrees["warm"]))
	}
}

func TestVariable_Fuzzify_boundary(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()
	degrees := v.Fuzzify(25)

	// At x=25: cold=Triangular(0,0,50)(25)=0.5, warm=Triangular(20,50,80)(25)=0.166.
	if degrees["cold"] != 0.5 {
		t.Fatalf("expected cold=0.5 at x=25, got %f", float64(degrees["cold"]))
	}
}

func TestVariable_Terms_defensiveCopy(t *testing.T) {
	t.Parallel()

	v := makeTemperatureVariable()
	terms1 := v.Terms()
	terms2 := v.Terms()

	terms1[0].Name = "modified"

	if terms2[0].Name == "modified" {
		t.Fatal("expected defensive copy")
	}
}
