package explain

import (
	"strings"
	"testing"
)

func TestPhase_String_propagation(t *testing.T) {
	t.Parallel()

	if Propagation.String() != "propagation" {
		t.Fatalf("expected propagation, got %q", Propagation.String())
	}
}

func TestPhase_String_intervention(t *testing.T) {
	t.Parallel()

	if Intervention.String() != "intervention" {
		t.Fatalf("expected intervention, got %q", Intervention.String())
	}
}

func TestPhase_String_counterfactual(t *testing.T) {
	t.Parallel()

	if Counterfactual.String() != "counterfactual" {
		t.Fatalf("expected counterfactual, got %q", Counterfactual.String())
	}
}

func TestPhase_String_attribution(t *testing.T) {
	t.Parallel()

	if Attribution.String() != "attribution" {
		t.Fatalf("expected attribution, got %q", Attribution.String())
	}
}

func TestPhase_String_complete(t *testing.T) {
	t.Parallel()

	if Complete.String() != "complete" {
		t.Fatalf("expected complete, got %q", Complete.String())
	}
}

func TestPhase_String_outOfRange(t *testing.T) {
	t.Parallel()

	p := Phase(99)
	if p.String() != "unknown" {
		t.Fatalf("expected unknown for out-of-range, got %q", p.String())
	}
}

func TestStep_String_withValues(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  1,
		Phase:   Propagation,
		Message: "compute Z",
		Values:  map[string]float64{"X": 5.0, "Z": 10.0},
	}

	result := s.String()

	if !strings.Contains(result, "step 1") {
		t.Fatalf("expected step number, got %q", result)
	}

	if !strings.Contains(result, "[propagation]") {
		t.Fatalf("expected phase, got %q", result)
	}

	if !strings.Contains(result, "compute Z") {
		t.Fatalf("expected message, got %q", result)
	}

	if !strings.Contains(result, "X=5.0000") {
		t.Fatalf("expected X value, got %q", result)
	}
}

func TestStep_String_noValues(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  2,
		Phase:   Complete,
		Message: "done",
	}

	result := s.String()

	if strings.Contains(result, "{") {
		t.Fatalf("expected no values, got %q", result)
	}
}

func TestCausalAttribution_String(t *testing.T) {
	t.Parallel()

	a := CausalAttribution{
		Target:       "Y",
		Attributions: map[string]float64{"X": 0.8},
	}

	result := a.String()

	if !strings.Contains(result, "attribution(Y)") {
		t.Fatalf("expected attribution target, got %q", result)
	}

	if !strings.Contains(result, "X=0.8000") {
		t.Fatalf("expected attribution value, got %q", result)
	}
}

func TestTrace_String_basic(t *testing.T) {
	t.Parallel()

	obs := map[string]float64{"X": 5.0}
	tr := NewTrace(obs)
	tr = tr.AddStep(Step{Number: 1, Phase: Propagation, Message: "compute Z"})
	tr = tr.AddOutput("Z", 10.0)

	result := tr.String()

	if !strings.Contains(result, "observations:") {
		t.Fatalf("expected observations, got %q", result)
	}

	if !strings.Contains(result, "X=5.0000") {
		t.Fatalf("expected X observation, got %q", result)
	}

	if !strings.Contains(result, "outputs:") {
		t.Fatalf("expected outputs, got %q", result)
	}

	if !strings.Contains(result, "Z=10.0000") {
		t.Fatalf("expected Z output, got %q", result)
	}
}

func TestTrace_String_noOutputs(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	result := tr.String()

	if strings.Contains(result, "outputs:") {
		t.Fatalf("expected no outputs, got %q", result)
	}
}

func TestTrace_String_withAttribution(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddAttribution(CausalAttribution{
		Target:       "Y",
		Attributions: map[string]float64{"X": 0.8},
	})

	result := tr.String()

	if !strings.Contains(result, "attribution(Y)") {
		t.Fatalf("expected attribution, got %q", result)
	}
}

func TestFormatValues_sorted(t *testing.T) {
	t.Parallel()

	vals := map[string]float64{"C": 3.0, "A": 1.0, "B": 2.0}
	result := formatValues(vals)

	if result != "A=1.0000, B=2.0000, C=3.0000" {
		t.Fatalf("expected sorted values, got %q", result)
	}
}

func TestFormatValues_empty(t *testing.T) {
	t.Parallel()

	result := formatValues(nil)

	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}
