package explain

import (
	"testing"
)

func TestNewTrace(t *testing.T) {
	t.Parallel()

	obs := map[string]float64{"X": 5.0}
	tr := NewTrace(obs)

	if tr.Observations == nil {
		t.Fatal("expected non-nil observations")
	}

	if tr.Observations["X"] != 5.0 {
		t.Fatalf("expected X=5.0, got %f", tr.Observations["X"])
	}
}

func TestNewTrace_nil(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)

	if tr.Observations != nil {
		t.Fatal("expected nil observations")
	}
}

func TestTrace_AddStep(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)

	tr = tr.AddStep(Step{Number: 1, Phase: Propagation, Message: "test"})

	if len(tr.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(tr.Steps))
	}

	if tr.Steps[0].Number != 1 {
		t.Fatalf("expected step number 1, got %d", tr.Steps[0].Number)
	}
}

func TestTrace_AddStep_multiple(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddStep(Step{Number: 1, Phase: Propagation, Message: "first"})
	tr = tr.AddStep(Step{Number: 2, Phase: Complete, Message: "second"})

	if len(tr.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(tr.Steps))
	}
}

func TestTrace_AddOutput(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddOutput("Y", 13.0)

	if tr.Outputs == nil {
		t.Fatal("expected non-nil outputs")
	}

	if tr.Outputs["Y"] != 13.0 {
		t.Fatalf("expected Y=13.0, got %f", tr.Outputs["Y"])
	}
}

func TestTrace_AddOutput_multiple(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)
	tr = tr.AddOutput("X", 5.0)
	tr = tr.AddOutput("Y", 13.0)

	if len(tr.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(tr.Outputs))
	}
}

func TestTrace_AddAttribution(t *testing.T) {
	t.Parallel()

	tr := NewTrace(nil)

	attr := CausalAttribution{
		Target:       "Y",
		Attributions: map[string]float64{"X": 0.8, "Z": 0.2},
	}

	tr = tr.AddAttribution(attr)

	if len(tr.Attributions) != 1 {
		t.Fatalf("expected 1 attribution, got %d", len(tr.Attributions))
	}

	if tr.Attributions[0].Target != "Y" {
		t.Fatalf("expected target Y, got %s", tr.Attributions[0].Target)
	}
}
