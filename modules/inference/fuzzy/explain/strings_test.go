package explain

import (
	"strings"
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

func TestPhase_String_fuzzification(t *testing.T) {
	t.Parallel()

	if Fuzzification.String() != "fuzzification" {
		t.Fatalf("expected fuzzification, got %q", Fuzzification.String())
	}
}

func TestPhase_String_ruleEvaluation(t *testing.T) {
	t.Parallel()

	if RuleEvaluation.String() != "rule-evaluation" {
		t.Fatalf("expected rule-evaluation, got %q", RuleEvaluation.String())
	}
}

func TestPhase_String_aggregation(t *testing.T) {
	t.Parallel()

	if Aggregation.String() != "aggregation" {
		t.Fatalf("expected aggregation, got %q", Aggregation.String())
	}
}

func TestPhase_String_defuzzification(t *testing.T) {
	t.Parallel()

	if Defuzzification.String() != "defuzzification" {
		t.Fatalf("expected defuzzification, got %q", Defuzzification.String())
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
	if p.String() != "fuzzification" {
		t.Fatalf("expected fuzzification for out-of-range, got %q", p.String())
	}
}

func TestMembership_String(t *testing.T) {
	t.Parallel()

	m := Membership{Variable: "temp", Term: "hot", Degree: fuzzym.Degree(0.8)}
	result := m.String()

	if !strings.Contains(result, "temp/hot") {
		t.Fatalf("expected temp/hot, got %q", result)
	}
}

func TestActivation_String(t *testing.T) {
	t.Parallel()

	a := Activation{RuleName: "r1", Strength: 0.7, Output: "speed", Term: "fast"}
	result := a.String()

	if !strings.Contains(result, "r1") {
		t.Fatalf("expected r1, got %q", result)
	}

	if !strings.Contains(result, "speed/fast") {
		t.Fatalf("expected speed/fast, got %q", result)
	}
}

func TestStep_String_withMemberships(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  1,
		Phase:   Fuzzification,
		Message: "fuzzify temp",
		Memberships: []Membership{
			{Variable: "temp", Term: "hot", Degree: 0.8},
		},
	}

	result := s.String()

	if !strings.Contains(result, "step 1") {
		t.Fatalf("expected step 1, got %q", result)
	}

	if !strings.Contains(result, "[fuzzification]") {
		t.Fatalf("expected phase, got %q", result)
	}

	if !strings.Contains(result, "temp/hot") {
		t.Fatalf("expected membership, got %q", result)
	}
}

func TestStep_String_withActivations(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  2,
		Phase:   RuleEvaluation,
		Message: "evaluate r1",
		Activations: []Activation{
			{RuleName: "r1", Strength: 0.7, Output: "speed", Term: "fast"},
		},
	}

	result := s.String()

	if !strings.Contains(result, "r1") {
		t.Fatalf("expected rule name, got %q", result)
	}
}

func TestStep_String_noDetails(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  3,
		Phase:   Complete,
		Message: "done",
	}

	result := s.String()

	if !strings.Contains(result, "step 3") {
		t.Fatalf("expected step 3, got %q", result)
	}

	if !strings.Contains(result, "[complete]") {
		t.Fatalf("expected phase, got %q", result)
	}
}

func TestOutput_String(t *testing.T) {
	t.Parallel()

	o := Output{Variable: "speed", CrispValue: 42.5}
	result := o.String()

	if !strings.Contains(result, "speed") {
		t.Fatalf("expected speed, got %q", result)
	}

	if !strings.Contains(result, "42.5") {
		t.Fatalf("expected 42.5, got %q", result)
	}
}

func TestTrace_String_basic(t *testing.T) {
	t.Parallel()

	tr := NewTrace(map[string]float64{"temp": 75.0})
	tr = tr.AddStep(Step{Number: 1, Phase: Fuzzification, Message: "fuzzify"})

	result := tr.String()

	if !strings.Contains(result, "inputs:") {
		t.Fatalf("expected inputs header, got %q", result)
	}

	if !strings.Contains(result, "temp=75.00") {
		t.Fatalf("expected temp value, got %q", result)
	}
}

func TestTrace_String_withOutputs(t *testing.T) {
	t.Parallel()

	tr := NewTrace(map[string]float64{"temp": 75.0})
	tr = tr.AddOutput(Output{Variable: "speed", CrispValue: 50.0})

	result := tr.String()

	if !strings.Contains(result, "speed") {
		t.Fatalf("expected speed output, got %q", result)
	}
}

func TestTrace_String_emptyInputs(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	result := tr.String()

	if !strings.HasPrefix(result, "inputs: ") {
		t.Fatalf("expected inputs prefix, got %q", result)
	}
}

func TestFormatInputs_sorted(t *testing.T) {
	t.Parallel()

	inputs := map[string]float64{"C": 3.0, "A": 1.0, "B": 2.0}
	result := formatInputs(inputs)

	if result != "A=1.00, B=2.00, C=3.00" {
		t.Fatalf("expected sorted inputs, got %q", result)
	}
}
