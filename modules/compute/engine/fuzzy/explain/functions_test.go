package explain

import "testing"

func TestNewTrace(t *testing.T) {
	t.Parallel()

	inputs := map[string]float64{"temp": 75.0, "humidity": 0.5}
	tr := NewTrace(inputs)

	if len(tr.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(tr.Inputs))
	}

	if tr.Inputs["temp"] != 75.0 {
		t.Fatalf("expected 75.0, got %f", tr.Inputs["temp"])
	}
}

func TestNewTrace_nil(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)

	if tr.Inputs == nil {
		t.Fatal("expected non-nil inputs map")
	}

	if len(tr.Inputs) != 0 {
		t.Fatalf("expected 0 inputs, got %d", len(tr.Inputs))
	}
}

func TestNewTrace_defensiveCopy(t *testing.T) {
	t.Parallel()

	inputs := map[string]float64{"temp": 75.0}
	tr := NewTrace(inputs)

	inputs["temp"] = 999.0

	if tr.Inputs["temp"] != 75.0 {
		t.Fatal("expected defensive copy")
	}
}

func TestTrace_AddStep(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddStep(Step{Number: 1, Phase: Fuzzification, Message: "test"})

	if len(tr.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tr.Steps))
	}

	if tr.Steps[0].Number != 1 {
		t.Fatalf("expected step 1, got %d", tr.Steps[0].Number)
	}
}

func TestTrace_AddOutput(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddOutput(Output{Variable: "speed", CrispValue: 42.0})

	if len(tr.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(tr.Outputs))
	}

	if tr.Outputs[0].Variable != "speed" {
		t.Fatalf("expected speed, got %s", tr.Outputs[0].Variable)
	}
}

func TestTrace_AddStep_immutable(t *testing.T) {
	t.Parallel()

	tr1 := NewTrace(nil)
	tr2 := tr1.AddStep(Step{Number: 1, Phase: Fuzzification, Message: "test"})

	if len(tr1.Steps) != 0 {
		t.Fatal("expected original trace unchanged")
	}

	if len(tr2.Steps) != 1 {
		t.Fatal("expected new trace with step")
	}
}
