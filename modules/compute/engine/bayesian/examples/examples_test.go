package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/explain"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

// makeRainNetwork builds the classic sprinkler Bayesian network:
//
//	Rain → Sprinkler → WetGrass ← Rain
func makeRainNetwork() network.Network {
	bn := network.NewNetwork()

	rainCPT := bayesian.NewCPT("Rain", nil)
	rainCPT.Set(stats.Assignment{}, stats.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(network.Node{
		Variable: "Rain",
		CPT:      rainCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	sprinklerCPT := bayesian.NewCPT("Sprinkler", []stats.Var{"Rain"})
	sprinklerCPT.Set(stats.Assignment{"Rain": "true"}, stats.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(stats.Assignment{"Rain": "false"}, stats.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(network.Node{
		Variable: "Sprinkler",
		Parents:  []stats.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

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

// makeMedicalNetwork builds a simple diagnostic network:
//
//	Disease → TestResult
//	Disease → Symptom
func makeMedicalNetwork() network.Network {
	bn := network.NewNetwork()

	diseaseCPT := bayesian.NewCPT("Disease", nil)
	diseaseCPT.Set(stats.Assignment{}, stats.Distribution{"present": 0.01, "absent": 0.99})

	bn.AddNode(network.Node{
		Variable: "Disease",
		CPT:      diseaseCPT,
		Outcomes: []stats.Outcome{"present", "absent"},
	})

	testCPT := bayesian.NewCPT("TestResult", []stats.Var{"Disease"})
	testCPT.Set(stats.Assignment{"Disease": "present"}, stats.Distribution{"positive": 0.95, "negative": 0.05})
	testCPT.Set(stats.Assignment{"Disease": "absent"}, stats.Distribution{"positive": 0.10, "negative": 0.90})

	bn.AddNode(network.Node{
		Variable: "TestResult",
		Parents:  []stats.Var{"Disease"},
		CPT:      testCPT,
		Outcomes: []stats.Outcome{"positive", "negative"},
	})

	symptomCPT := bayesian.NewCPT("Symptom", []stats.Var{"Disease"})
	symptomCPT.Set(stats.Assignment{"Disease": "present"}, stats.Distribution{"yes": 0.80, "no": 0.20})
	symptomCPT.Set(stats.Assignment{"Disease": "absent"}, stats.Distribution{"yes": 0.05, "no": 0.95})

	bn.AddNode(network.Node{
		Variable: "Symptom",
		Parents:  []stats.Var{"Disease"},
		CPT:      symptomCPT,
		Outcomes: []stats.Outcome{"yes", "no"},
	})

	return bn
}

func TestEnumerationPriorQuery(t *testing.T) {
	t.Parallel()

	t.Run("prior probability without evidence", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		eng := engine.NewEngine()

		result := eng.Query(bn, ev, "Rain")

		got := float64(result.Posterior["true"])
		if math.Abs(got-0.2) > 0.01 {
			t.Fatalf("expected P(Rain=true) ≈ 0.2, got %f", got)
		}
	})
}

func TestEnumerationWithEvidence(t *testing.T) {
	t.Parallel()

	t.Run("wet grass increases rain probability", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		got := float64(result.Posterior["true"])
		if got <= 0.2 {
			t.Fatalf("expected P(Rain|WetGrass=true) > 0.2, got %f", got)
		}
	})

	t.Run("sprinkler observed reduces rain probability", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")
		ev.Observe("Sprinkler", "true")

		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		noSprinklerEv := evidence.NewEvidenceBase()
		noSprinklerEv.Observe("WetGrass", "true")

		resultNoSprinkler := eng.Query(bn, noSprinklerEv, "Rain")

		withSprinkler := float64(result.Posterior["true"])

		withoutSprinkler := float64(resultNoSprinkler.Posterior["true"])
		if withSprinkler >= withoutSprinkler {
			t.Fatalf("expected explaining away: P(Rain|Wet,Sprinkler) < P(Rain|Wet), got %f >= %f",
				withSprinkler, withoutSprinkler)
		}
	})
}

func TestVariableEliminationMatchesEnumeration(t *testing.T) {
	t.Parallel()

	t.Run("both algorithms produce same posterior", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		enumEng := engine.NewEngine()
		veEng := engine.NewEngine(engine.WithAlgorithm(engine.VariableElimination))

		enumResult := enumEng.Query(bn, ev, "Rain")
		veResult := veEng.Query(bn, ev, "Rain")

		enumTrue := float64(enumResult.Posterior["true"])

		veTrue := float64(veResult.Posterior["true"])
		if math.Abs(enumTrue-veTrue) > 0.01 {
			t.Fatalf("algorithms disagree: enumeration=%f, VE=%f", enumTrue, veTrue)
		}
	})
}

func TestVariableEliminationCustomOrder(t *testing.T) {
	t.Parallel()

	t.Run("custom elimination order produces valid result", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		eng := engine.NewEngine(
			engine.WithAlgorithm(engine.VariableElimination),
			engine.WithEliminationOrder([]stats.Var{"Sprinkler", "WetGrass"}),
		)

		result := eng.Query(bn, ev, "Rain")

		sum := result.Posterior["true"] + result.Posterior["false"]
		if sum < 0.99 {
			t.Fatalf("expected normalized posterior, sum = %f", float64(sum))
		}
	})
}

func TestMedicalDiagnosis(t *testing.T) {
	t.Parallel()

	t.Run("positive test increases disease probability", func(t *testing.T) {
		t.Parallel()

		bn := makeMedicalNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("TestResult", "positive")

		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Disease")

		got := float64(result.Posterior["present"])
		if got <= 0.01 {
			t.Fatalf("expected P(Disease|positive test) > prior 0.01, got %f", got)
		}
	})

	t.Run("multiple evidence strengthens diagnosis", func(t *testing.T) {
		t.Parallel()

		bn := makeMedicalNetwork()
		eng := engine.NewEngine()

		testOnly := evidence.NewEvidenceBase()
		testOnly.Observe("TestResult", "positive")
		resultTestOnly := eng.Query(bn, testOnly, "Disease")

		testAndSymptom := evidence.NewEvidenceBase()
		testAndSymptom.Observe("TestResult", "positive")
		testAndSymptom.Observe("Symptom", "yes")
		resultBoth := eng.Query(bn, testAndSymptom, "Disease")

		probTestOnly := float64(resultTestOnly.Posterior["present"])

		probBoth := float64(resultBoth.Posterior["present"])
		if probBoth <= probTestOnly {
			t.Fatalf("expected more evidence to increase P(Disease), got %f <= %f", probBoth, probTestOnly)
		}
	})
}

func TestTraceInspection(t *testing.T) {
	t.Parallel()

	t.Run("trace contains query and posteriors", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		if result.Trace.Query != "Rain" {
			t.Fatalf("expected query=Rain, got %s", string(result.Trace.Query))
		}

		if len(result.Trace.Steps) == 0 {
			t.Fatal("expected non-empty trace steps")
		}

		if len(result.Trace.Posteriors) == 0 {
			t.Fatal("expected posteriors in trace")
		}
	})

	t.Run("trace string is non-empty", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		traceStr := result.Trace.String()
		if traceStr == "" {
			t.Fatal("expected non-empty trace string")
		}
	})

	t.Run("trace phases are ordered", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		ev := evidence.NewEvidenceBase()
		eng := engine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		hasInit := false
		hasComplete := false

		for _, step := range result.Trace.Steps {
			if step.Phase == explain.Initialize {
				hasInit = true
			}

			if step.Phase == explain.Complete {
				hasComplete = true
			}
		}

		if !hasInit {
			t.Fatal("expected initialize phase in trace")
		}

		if !hasComplete {
			t.Fatal("expected complete phase in trace")
		}
	})
}

func TestEvidenceBaseOperations(t *testing.T) {
	t.Parallel()

	t.Run("observe and retract", func(t *testing.T) {
		t.Parallel()

		ev := evidence.NewEvidenceBase()
		ev.Observe("Rain", "true")

		outcome, ok := ev.Get("Rain")
		if !ok {
			t.Fatal("expected Rain observed")
		}

		if outcome != "true" {
			t.Fatalf("expected true, got %s", string(outcome))
		}

		ev.Retract("Rain")

		_, ok = ev.Get("Rain")
		if ok {
			t.Fatal("expected Rain retracted")
		}
	})

	t.Run("clone independence", func(t *testing.T) {
		t.Parallel()

		ev := evidence.NewEvidenceBase()
		ev.Observe("Rain", "true")

		clone := ev.Clone()
		clone.Observe("Sprinkler", "false")

		if ev.Len() != 1 {
			t.Fatalf("expected original len=1, got %d", ev.Len())
		}

		if clone.Len() != 2 {
			t.Fatalf("expected clone len=2, got %d", clone.Len())
		}
	})
}

func TestNetworkValidation(t *testing.T) {
	t.Parallel()

	t.Run("valid network passes validation", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()

		err := bn.Validate()
		if err != nil {
			t.Fatalf("expected valid network, got: %v", err)
		}
	})

	t.Run("topological order is consistent", func(t *testing.T) {
		t.Parallel()

		bn := makeRainNetwork()
		order := bn.TopologicalOrder()

		if len(order) != 3 {
			t.Fatalf("expected 3 variables in order, got %d", len(order))
		}

		rainIdx := -1
		wetIdx := -1

		for i, v := range order {
			if v == "Rain" {
				rainIdx = i
			}

			if v == "WetGrass" {
				wetIdx = i
			}
		}

		if rainIdx >= wetIdx {
			t.Fatalf("expected Rain before WetGrass, got indices %d, %d", rainIdx, wetIdx)
		}
	})
}
