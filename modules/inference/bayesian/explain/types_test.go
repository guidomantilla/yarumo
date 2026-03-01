package explain

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func TestPhase_constants(t *testing.T) {
	t.Parallel()

	if Initialize != 0 {
		t.Fatalf("expected Initialize=0, got %d", Initialize)
	}

	if Propagate != 1 {
		t.Fatalf("expected Propagate=1, got %d", Propagate)
	}

	if Marginalize != 2 {
		t.Fatalf("expected Marginalize=2, got %d", Marginalize)
	}

	if Complete != 3 {
		t.Fatalf("expected Complete=3, got %d", Complete)
	}
}

func TestFactor_struct(t *testing.T) {
	t.Parallel()

	f := Factor{
		Variables: []probability.Var{"A", "B"},
		Size:      4,
	}

	if len(f.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(f.Variables))
	}

	if f.Size != 4 {
		t.Fatalf("expected size 4, got %d", f.Size)
	}
}

func TestStep_struct(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  1,
		Phase:   Propagate,
		Message: "test message",
		Factor:  Factor{Variables: []probability.Var{"X"}, Size: 2},
	}

	if s.Number != 1 {
		t.Fatalf("expected step 1, got %d", s.Number)
	}

	if s.Phase != Propagate {
		t.Fatalf("expected Propagate, got %d", s.Phase)
	}
}

func TestPosterior_struct(t *testing.T) {
	t.Parallel()

	p := Posterior{
		Variable:     "X",
		Distribution: probability.Distribution{"true": 0.7, "false": 0.3},
	}

	if p.Variable != "X" {
		t.Fatalf("expected X, got %s", string(p.Variable))
	}
}

func TestTrace_struct(t *testing.T) {
	t.Parallel()

	tr := Trace{
		Query:    "X",
		Evidence: probability.Assignment{"Y": "true"},
	}

	if tr.Query != "X" {
		t.Fatalf("expected X, got %s", string(tr.Query))
	}

	if tr.Evidence["Y"] != "true" {
		t.Fatalf("expected Y=true, got %s", string(tr.Evidence["Y"]))
	}
}
