package engine

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

func makeRainNetwork() network.Network {
	bn := network.NewNetwork()

	// Rain node (no parents).
	rainCPT := bayesian.NewCPT("Rain", nil)
	rainCPT.Set(stats.Assignment{}, stats.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(network.Node{
		Variable: "Rain",
		CPT:      rainCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	// Sprinkler | Rain.
	sprinklerCPT := bayesian.NewCPT("Sprinkler", []stats.Var{"Rain"})
	sprinklerCPT.Set(stats.Assignment{"Rain": "true"}, stats.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(stats.Assignment{"Rain": "false"}, stats.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(network.Node{
		Variable: "Sprinkler",
		Parents:  []stats.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	// WetGrass | Rain, Sprinkler.
	wetCPT := bayesian.NewCPT("WetGrass", []stats.Var{"Rain", "Sprinkler"})
	wetCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "true"}, stats.Distribution{"true": 0.99, "false": 0.01})
	wetCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "false"}, stats.Distribution{"true": 0.8, "false": 0.2})
	wetCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "true"}, stats.Distribution{"true": 0.9, "false": 0.1})
	wetCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "false"}, stats.Distribution{"true": 0.0, "false": 1.0})

	bn.AddNode(network.Node{
		Variable: "WetGrass",
		Parents:  []stats.Var{"Rain", "Sprinkler"},
		CPT:      wetCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	return bn
}

func TestNewEngine(t *testing.T) {
	t.Parallel()

	eng := NewEngine()
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestEngine_Query_enumeration_priorRain(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine()

	result := eng.Query(bn, ev, "Rain")

	// With no evidence, P(Rain=true) should be 0.2.
	if math.Abs(float64(result.Posterior["true"])-0.2) > 0.01 {
		t.Fatalf("expected P(Rain=true) ≈ 0.2, got %f", float64(result.Posterior["true"]))
	}
}

func TestEngine_Query_enumeration_withEvidence(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	ev.Observe("WetGrass", "true")

	eng := NewEngine()
	result := eng.Query(bn, ev, "Rain")

	// P(Rain=true | WetGrass=true) should be higher than prior 0.2.
	if float64(result.Posterior["true"]) <= 0.2 {
		t.Fatalf("expected P(Rain|WetGrass=true) > 0.2, got %f", float64(result.Posterior["true"]))
	}

	if len(result.Trace.Steps) == 0 {
		t.Fatal("expected non-empty trace")
	}
}

func TestEngine_Query_variableElimination_priorRain(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine(WithAlgorithm(VariableElimination))

	result := eng.Query(bn, ev, "Rain")

	// Should produce a valid distribution.
	if result.Posterior["true"]+result.Posterior["false"] < 0.99 {
		t.Fatalf("expected normalized posterior, sum = %f",
			float64(result.Posterior["true"]+result.Posterior["false"]))
	}
}

func TestEngine_Query_variableElimination_withEvidence(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	ev.Observe("WetGrass", "true")

	eng := NewEngine(WithAlgorithm(VariableElimination))
	result := eng.Query(bn, ev, "Rain")

	// P(Rain=true | WetGrass=true) should be higher than prior 0.2.
	if float64(result.Posterior["true"]) <= 0.2 {
		t.Fatalf("expected P(Rain|WetGrass=true) > 0.2, got %f", float64(result.Posterior["true"]))
	}
}

func TestEngine_Query_variableElimination_customOrder(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine(
		WithAlgorithm(VariableElimination),
		WithEliminationOrder([]stats.Var{"Sprinkler", "WetGrass"}),
	)

	result := eng.Query(bn, ev, "Rain")

	if result.Posterior["true"]+result.Posterior["false"] < 0.99 {
		t.Fatalf("expected normalized posterior, sum = %f",
			float64(result.Posterior["true"]+result.Posterior["false"]))
	}
}

func TestEngine_Query_trace(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine()

	result := eng.Query(bn, ev, "Rain")

	if result.Trace.Query != "Rain" {
		t.Fatalf("expected query Rain, got %s", string(result.Trace.Query))
	}

	if len(result.Trace.Posteriors) == 0 {
		t.Fatal("expected posteriors in trace")
	}
}

func TestAlgorithm_constants(t *testing.T) {
	t.Parallel()

	if Enumeration != 0 {
		t.Fatalf("expected Enumeration=0, got %d", Enumeration)
	}

	if VariableElimination != 1 {
		t.Fatalf("expected VariableElimination=1, got %d", VariableElimination)
	}
}

// makeImpossibleEvidenceNetwork builds a network where observing B=b1 makes
// all joint probabilities zero, causing Normalize to fail.
//
//	A: P(a1)=1.0, P(a2)=0.0
//	B | A: P(b1|a1)=0.0, P(b2|a1)=1.0; P(b1|a2)=0.0, P(b2|a2)=1.0
//
// Querying A given B=b1 yields zero probability for every outcome.
func makeImpossibleEvidenceNetwork() network.Network {
	bn := network.NewNetwork()

	aCPT := bayesian.NewCPT("A", nil)
	aCPT.Set(stats.Assignment{}, stats.Distribution{"a1": 1.0, "a2": 0.0})

	bn.AddNode(network.Node{
		Variable: "A",
		CPT:      aCPT,
		Outcomes: []stats.Outcome{"a1", "a2"},
	})

	bCPT := bayesian.NewCPT("B", []stats.Var{"A"})
	bCPT.Set(stats.Assignment{"A": "a1"}, stats.Distribution{"b1": 0.0, "b2": 1.0})
	bCPT.Set(stats.Assignment{"A": "a2"}, stats.Distribution{"b1": 0.0, "b2": 1.0})

	bn.AddNode(network.Node{
		Variable: "B",
		Parents:  []stats.Var{"A"},
		CPT:      bCPT,
		Outcomes: []stats.Outcome{"b1", "b2"},
	})

	return bn
}

// makeMismatchedOutcomeNetwork builds a network where a parent variable uses
// outcomes ("high","low") that differ from the child CPT distribution keys
// ("yes","no"). outcomesForVar for the parent returns the child outcomes,
// causing CPT Lookup to fail for those assignments.
func makeMismatchedOutcomeNetwork() network.Network {
	bn := network.NewNetwork()

	// Sensor: outcomes are "high" and "low".
	sensorCPT := bayesian.NewCPT("Sensor", nil)
	sensorCPT.Set(stats.Assignment{}, stats.Distribution{"high": 0.6, "low": 0.4})

	bn.AddNode(network.Node{
		Variable: "Sensor",
		CPT:      sensorCPT,
		Outcomes: []stats.Outcome{"high", "low"},
	})

	// Alarm | Sensor: outcomes are "yes" and "no".
	// CPT keyed by Sensor="high" and Sensor="low".
	alarmCPT := bayesian.NewCPT("Alarm", []stats.Var{"Sensor"})
	alarmCPT.Set(stats.Assignment{"Sensor": "high"}, stats.Distribution{"yes": 0.9, "no": 0.1})
	alarmCPT.Set(stats.Assignment{"Sensor": "low"}, stats.Distribution{"yes": 0.2, "no": 0.8})

	bn.AddNode(network.Node{
		Variable: "Alarm",
		Parents:  []stats.Var{"Sensor"},
		CPT:      alarmCPT,
		Outcomes: []stats.Outcome{"yes", "no"},
	})

	return bn
}

func TestEngine_Query_enumeration_impossibleEvidence(t *testing.T) {
	t.Parallel()

	bn := makeImpossibleEvidenceNetwork()
	ev := evidence.NewEvidenceBase()
	ev.Observe("B", "b1")

	eng := NewEngine()

	result := eng.Query(bn, ev, "A")

	// Normalize fails (all-zero), so the engine returns the raw zero distribution.
	sumProb := float64(result.Posterior["a1"]) + float64(result.Posterior["a2"])
	if sumProb != 0.0 {
		t.Fatalf("expected zero sum for impossible evidence, got %f", sumProb)
	}
}

func TestEngine_Query_enumeration_cptLookupFailure(t *testing.T) {
	t.Parallel()

	bn := makeMismatchedOutcomeNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine()

	// Query Alarm with no evidence. jointProbability will encounter CPT lookup
	// failures for the Alarm node because outcomesForVar returns child outcomes
	// ("yes","no") as parent values instead of ("high","low").
	result := eng.Query(bn, ev, "Alarm")

	// The query should still complete without panic.
	if result.Posterior == nil {
		t.Fatal("expected non-nil posterior")
	}
}

func TestEngine_Query_variableElimination_cptLookupFailure(t *testing.T) {
	t.Parallel()

	bn := makeMismatchedOutcomeNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine(WithAlgorithm(VariableElimination))

	// cptToFactor will encounter CPT lookup failures for the Alarm node
	// because generateEntries uses child outcomes as parent values.
	result := eng.Query(bn, ev, "Alarm")

	// The query should still complete without panic.
	if result.Posterior == nil {
		t.Fatal("expected non-nil posterior")
	}
}

func TestEngine_Query_variableElimination_noRelevantFactors(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	// Include "Phantom" in the elimination order. No factor mentions this variable,
	// so len(relevant) == 0 and the continue branch is taken.
	eng := NewEngine(
		WithAlgorithm(VariableElimination),
		WithEliminationOrder([]stats.Var{"Phantom", "Sprinkler", "WetGrass"}),
	)

	result := eng.Query(bn, ev, "Rain")

	// Despite the phantom variable, the result should still be valid.
	sumProb := result.Posterior["true"] + result.Posterior["false"]
	if sumProb < 0.99 {
		t.Fatalf("expected normalized posterior, sum = %f", float64(sumProb))
	}
}

func TestEngine_Query_variableElimination_evidenceOnParent(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	// Observe Rain so that generateEntries hits the evidence branch for a
	// variable that is first in allVars with non-empty sub-entries after it.
	// For the WetGrass node, allVars=[Rain, Sprinkler, WetGrass]. Rain is
	// first and in evidence; the recursive sub-entries for [Sprinkler, WetGrass]
	// are non-empty, so the inner copy loop executes.
	ev.Observe("Rain", "true")

	eng := NewEngine(WithAlgorithm(VariableElimination))

	result := eng.Query(bn, ev, "Sprinkler")

	// P(Sprinkler=true | Rain=true) should be about 0.01.
	if result.Posterior == nil {
		t.Fatal("expected non-nil posterior")
	}

	if result.Posterior["true"]+result.Posterior["false"] < 0.99 {
		t.Fatalf("expected normalized posterior, sum = %f",
			float64(result.Posterior["true"]+result.Posterior["false"]))
	}
}

// makeIncompleteCPTNetwork builds a network where a parent has 3 outcomes
// but the child CPT only defines entries for 2 of them, so Lookup fails
// for the third parent configuration.
func makeIncompleteCPTNetwork() network.Network {
	bn := network.NewNetwork()

	// Level: outcomes "low", "medium", "high".
	levelCPT := bayesian.NewCPT("Level", nil)
	levelCPT.Set(stats.Assignment{}, stats.Distribution{
		"low": 0.3, "medium": 0.4, "high": 0.3,
	})

	bn.AddNode(network.Node{
		Variable: "Level",
		CPT:      levelCPT,
		Outcomes: []stats.Outcome{"low", "medium", "high"},
	})

	// Alert | Level: only defines CPT entries for "low" and "high".
	// Missing entry for Level="medium" causes Lookup to fail.
	alertCPT := bayesian.NewCPT("Alert", []stats.Var{"Level"})
	alertCPT.Set(stats.Assignment{"Level": "low"}, stats.Distribution{"on": 0.1, "off": 0.9})
	alertCPT.Set(stats.Assignment{"Level": "high"}, stats.Distribution{"on": 0.9, "off": 0.1})

	bn.AddNode(network.Node{
		Variable: "Alert",
		Parents:  []stats.Var{"Level"},
		CPT:      alertCPT,
		Outcomes: []stats.Outcome{"on", "off"},
	})

	return bn
}

func TestEngine_Query_enumeration_incompleteCPT(t *testing.T) {
	t.Parallel()

	bn := makeIncompleteCPTNetwork()
	ev := evidence.NewEvidenceBase()
	eng := NewEngine()

	// jointProbability will encounter a CPT Lookup error for Level="medium"
	// in the Alert node, triggering the err != nil continue branch.
	result := eng.Query(bn, ev, "Alert")

	if result.Posterior == nil {
		t.Fatal("expected non-nil posterior")
	}

	// The result should still produce a distribution (albeit partial).
	sumProb := float64(result.Posterior["on"]) + float64(result.Posterior["off"])
	if sumProb == 0.0 {
		t.Fatal("expected non-zero posterior sum")
	}
}
