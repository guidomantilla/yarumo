package explain

import (
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func TestPhase_String_initialize(t *testing.T) {
	t.Parallel()

	if Initialize.String() != "initialize" {
		t.Fatalf("expected initialize, got %q", Initialize.String())
	}
}

func TestPhase_String_propagate(t *testing.T) {
	t.Parallel()

	if Propagate.String() != "propagate" {
		t.Fatalf("expected propagate, got %q", Propagate.String())
	}
}

func TestPhase_String_marginalize(t *testing.T) {
	t.Parallel()

	if Marginalize.String() != "marginalize" {
		t.Fatalf("expected marginalize, got %q", Marginalize.String())
	}
}

func TestPhase_String_complete(t *testing.T) {
	t.Parallel()

	if Complete.String() != "complete" {
		t.Fatalf("expected complete, got %q", Complete.String())
	}
}

func TestFactor_String(t *testing.T) {
	t.Parallel()

	f := Factor{Variables: []probability.Var{"A", "B"}, Size: 4}
	result := f.String()

	if result != "Factor(A, B)[4]" {
		t.Fatalf("expected Factor(A, B)[4], got %q", result)
	}
}

func TestFactor_String_empty(t *testing.T) {
	t.Parallel()

	f := Factor{}
	result := f.String()

	if result != "Factor()[0]" {
		t.Fatalf("expected Factor()[0], got %q", result)
	}
}

func TestStep_String_withFactor(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  1,
		Phase:   Propagate,
		Message: "multiplying factors",
		Factor:  Factor{Variables: []probability.Var{"X"}, Size: 2},
	}

	result := s.String()

	if !strings.Contains(result, "step 1") {
		t.Fatalf("expected step number, got %q", result)
	}

	if !strings.Contains(result, "[propagate]") {
		t.Fatalf("expected phase, got %q", result)
	}

	if !strings.Contains(result, "Factor(X)[2]") {
		t.Fatalf("expected factor, got %q", result)
	}
}

func TestStep_String_noFactor(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  2,
		Phase:   Complete,
		Message: "inference complete",
	}

	result := s.String()

	if strings.Contains(result, "Factor") {
		t.Fatalf("expected no factor, got %q", result)
	}
}

func TestPosterior_String(t *testing.T) {
	t.Parallel()

	p := Posterior{
		Variable:     "X",
		Distribution: probability.Distribution{"true": 0.7, "false": 0.3},
	}

	result := p.String()

	if !strings.HasPrefix(result, "P(X) = ") {
		t.Fatalf("expected P(X) prefix, got %q", result)
	}
}

func TestTrace_String_basic(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", probability.Assignment{"Y": "true"})
	tr = tr.AddStep(Step{Number: 1, Phase: Initialize, Message: "init"})

	result := tr.String()

	if !strings.Contains(result, "query: X") {
		t.Fatalf("expected query, got %q", result)
	}

	if !strings.Contains(result, "evidence: Y=true") {
		t.Fatalf("expected evidence, got %q", result)
	}
}

func TestTrace_String_noEvidence(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", nil)
	result := tr.String()

	if strings.Contains(result, "evidence") {
		t.Fatalf("expected no evidence, got %q", result)
	}
}

func TestTrace_String_withPosterior(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", nil)
	tr = tr.AddPosterior(Posterior{
		Variable:     "X",
		Distribution: probability.Distribution{"true": 0.7, "false": 0.3},
	})

	result := tr.String()

	if !strings.Contains(result, "P(X)") {
		t.Fatalf("expected posterior, got %q", result)
	}
}

func TestPhase_String_outOfRange(t *testing.T) {
	t.Parallel()

	p := Phase(99)
	if p.String() != "initialize" {
		t.Fatalf("expected initialize for out-of-range, got %q", p.String())
	}
}

func TestFormatEvidence_sorted(t *testing.T) {
	t.Parallel()

	ev := probability.Assignment{"C": "c1", "A": "a1", "B": "b1"}
	result := formatEvidence(ev)

	if result != "A=a1, B=b1, C=c1" {
		t.Fatalf("expected sorted evidence, got %q", result)
	}
}
