package explain

import (
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

func TestPhase_constants(t *testing.T) {
	t.Parallel()

	if Fuzzification != 0 {
		t.Fatalf("expected Fuzzification=0, got %d", Fuzzification)
	}

	if RuleEvaluation != 1 {
		t.Fatalf("expected RuleEvaluation=1, got %d", RuleEvaluation)
	}

	if Aggregation != 2 {
		t.Fatalf("expected Aggregation=2, got %d", Aggregation)
	}

	if Defuzzification != 3 {
		t.Fatalf("expected Defuzzification=3, got %d", Defuzzification)
	}

	if Complete != 4 {
		t.Fatalf("expected Complete=4, got %d", Complete)
	}
}

func TestMembership_struct(t *testing.T) {
	t.Parallel()

	m := Membership{Variable: "temp", Term: "hot", Degree: 0.8}

	if m.Variable != "temp" {
		t.Fatalf("expected temp, got %s", m.Variable)
	}

	if m.Term != "hot" {
		t.Fatalf("expected hot, got %s", m.Term)
	}

	if m.Degree != 0.8 {
		t.Fatalf("expected 0.8, got %f", float64(m.Degree))
	}
}

func TestActivation_struct(t *testing.T) {
	t.Parallel()

	output := "velocity"
	a := Activation{RuleName: "r1", Strength: 0.7, Output: output, Term: "fast"}

	if a.RuleName != "r1" {
		t.Fatalf("expected r1, got %s", a.RuleName)
	}

	if a.Strength != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(a.Strength))
	}

	if a.Output != output {
		t.Fatalf("expected %s, got %s", output, a.Output)
	}

	if a.Term != "fast" {
		t.Fatalf("expected fast, got %s", a.Term)
	}
}

func TestStep_struct(t *testing.T) {
	t.Parallel()

	s := Step{
		Number:  1,
		Phase:   Fuzzification,
		Message: "fuzzify temp",
		Memberships: []Membership{
			{Variable: "temp", Term: "hot", Degree: fuzzym.Degree(0.8)},
		},
	}

	if s.Number != 1 {
		t.Fatalf("expected 1, got %d", s.Number)
	}

	if s.Phase != Fuzzification {
		t.Fatalf("expected Fuzzification, got %d", s.Phase)
	}

	if len(s.Memberships) != 1 {
		t.Fatalf("expected 1 membership, got %d", len(s.Memberships))
	}
}

func TestOutput_struct(t *testing.T) {
	t.Parallel()

	varName := "rpm"
	o := Output{Variable: varName, CrispValue: 42.5}

	if o.Variable != varName {
		t.Fatalf("expected %s, got %s", varName, o.Variable)
	}

	if o.CrispValue != 42.5 {
		t.Fatalf("expected 42.5, got %f", o.CrispValue)
	}
}

func TestTrace_struct(t *testing.T) {
	t.Parallel()

	outVar := "fanSpeed"
	tr := Trace{
		Inputs:  map[string]float64{"temp": 75.0},
		Steps:   []Step{{Number: 1, Phase: Fuzzification, Message: "test"}},
		Outputs: []Output{{Variable: outVar, CrispValue: 50.0}},
	}

	if len(tr.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tr.Steps))
	}

	if len(tr.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(tr.Outputs))
	}

	if tr.Inputs["temp"] != 75.0 {
		t.Fatalf("expected 75.0, got %f", tr.Inputs["temp"])
	}
}
