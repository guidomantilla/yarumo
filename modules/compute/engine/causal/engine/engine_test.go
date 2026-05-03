package engine

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/causal/explain"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

const tolerance = 1e-9

// buildLinearSCM creates: X→Z→Y where Z=2*X, Y=Z+3.
func buildLinearSCM(t *testing.T) model.SCM {
	t.Helper()

	s := model.NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Y", []string{"Z"}, func(parents map[string]float64) float64 {
		return parents["Z"] + 3
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return s
}

// buildMultiRootSCM creates: X,W→Y where Y=X+W.
func buildMultiRootSCM(t *testing.T) model.SCM {
	t.Helper()

	s := model.NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("W", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Y", []string{"X", "W"}, func(parents map[string]float64) float64 {
		return parents["X"] + parents["W"]
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return s
}

func assertFloat(t *testing.T, name string, got float64, want float64) {
	t.Helper()

	if math.Abs(got-want) > tolerance {
		t.Fatalf("expected %s=%.4f, got %.4f", name, want, got)
	}
}

func TestNewEngine(t *testing.T) {
	t.Parallel()

	e := NewEngine()
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestEngine_Propagate_basic(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	result := e.Propagate(scm, map[string]float64{"X": 5})

	assertFloat(t, "X", result.Values["X"], 5)
	assertFloat(t, "Z", result.Values["Z"], 10)
	assertFloat(t, "Y", result.Values["Y"], 13)
}

func TestEngine_Propagate_multipleRoots(t *testing.T) {
	t.Parallel()

	scm := buildMultiRootSCM(t)
	e := NewEngine()

	result := e.Propagate(scm, map[string]float64{"X": 2, "W": 3})

	assertFloat(t, "X", result.Values["X"], 2)
	assertFloat(t, "W", result.Values["W"], 3)
	assertFloat(t, "Y", result.Values["Y"], 5)
}

func TestEngine_Intervene_basic(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	// do(Z=7) cuts Z's dependence on X. Y = Z+3 = 10.
	result := e.Intervene(scm, map[string]float64{"Z": 7})

	assertFloat(t, "Z", result.Values["Z"], 7)
	assertFloat(t, "Y", result.Values["Y"], 10)
}

func TestEngine_Intervene_root(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	// do(X=10), Z=20, Y=23.
	result := e.Intervene(scm, map[string]float64{"X": 10})

	assertFloat(t, "X", result.Values["X"], 10)
	assertFloat(t, "Z", result.Values["Z"], 20)
	assertFloat(t, "Y", result.Values["Y"], 23)
}

func TestEngine_Counterfactual_basic(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	// Factual: X=5 → Z=10, Y=13.
	// Hypothetical: do(X=10) → Z=20, Y=23.
	result := e.Counterfactual(scm, map[string]float64{"X": 5}, map[string]float64{"X": 10})

	assertFloat(t, "X", result.Values["X"], 10)
	assertFloat(t, "Z", result.Values["Z"], 20)
	assertFloat(t, "Y", result.Values["Y"], 23)
}

func TestEngine_Counterfactual_midChain(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	// Factual: X=5 → Z=10, Y=13.
	// Hypothetical: do(Z=7) → Y=10, X stays 5.
	result := e.Counterfactual(scm, map[string]float64{"X": 5}, map[string]float64{"Z": 7})

	assertFloat(t, "X", result.Values["X"], 5)
	assertFloat(t, "Z", result.Values["Z"], 7)
	assertFloat(t, "Y", result.Values["Y"], 10)
}

func TestEngine_Propagate_trace(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	result := e.Propagate(scm, map[string]float64{"X": 5})

	if len(result.Trace.Steps) == 0 {
		t.Fatal("expected non-empty trace steps")
	}

	// Last step should be Complete.
	lastStep := result.Trace.Steps[len(result.Trace.Steps)-1]
	if lastStep.Phase != explain.Complete {
		t.Fatalf("expected last step phase Complete, got %s", lastStep.Phase.String())
	}

	// Outputs should be recorded.
	if len(result.Trace.Outputs) == 0 {
		t.Fatal("expected non-empty trace outputs")
	}

	assertFloat(t, "trace output Y", result.Trace.Outputs["Y"], 13)
}

func TestEngine_Intervene_trace(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	result := e.Intervene(scm, map[string]float64{"Z": 7})

	// First step should be Intervention.
	if len(result.Trace.Steps) == 0 {
		t.Fatal("expected non-empty trace steps")
	}

	firstStep := result.Trace.Steps[0]
	if firstStep.Phase != explain.Intervention {
		t.Fatalf("expected first step phase Intervention, got %s", firstStep.Phase.String())
	}

	// Last step should be Complete.
	lastStep := result.Trace.Steps[len(result.Trace.Steps)-1]
	if lastStep.Phase != explain.Complete {
		t.Fatalf("expected last step phase Complete, got %s", lastStep.Phase.String())
	}
}

func TestEngine_Counterfactual_trace(t *testing.T) {
	t.Parallel()

	scm := buildLinearSCM(t)
	e := NewEngine()

	result := e.Counterfactual(scm, map[string]float64{"X": 5}, map[string]float64{"X": 10})

	// First step should be Counterfactual (factual world computed).
	if len(result.Trace.Steps) == 0 {
		t.Fatal("expected non-empty trace steps")
	}

	firstStep := result.Trace.Steps[0]
	if firstStep.Phase != explain.Counterfactual {
		t.Fatalf("expected first step phase Counterfactual, got %s", firstStep.Phase.String())
	}

	// Last step should be Complete.
	lastStep := result.Trace.Steps[len(result.Trace.Steps)-1]
	if lastStep.Phase != explain.Complete {
		t.Fatalf("expected last step phase Complete, got %s", lastStep.Phase.String())
	}

	// Observations should record factual values.
	assertFloat(t, "trace observation X", result.Trace.Observations["X"], 5)
}
