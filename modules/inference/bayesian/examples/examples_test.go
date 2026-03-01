package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/engine"
	"github.com/guidomantilla/yarumo/inference/bayesian/evidence"
	"github.com/guidomantilla/yarumo/inference/bayesian/explain"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

// makeRainNetwork builds the classic sprinkler Bayesian network:
//
//	Rain → Sprinkler → WetGrass ← Rain
func makeRainNetwork() network.Network {
	bn := network.NewNetwork()

	rainCPT := probability.NewCPT("Rain", nil)
	rainCPT.Set(probability.Assignment{}, probability.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(network.Node{
		Variable: "Rain",
		CPT:      rainCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	sprinklerCPT := probability.NewCPT("Sprinkler", []probability.Var{"Rain"})
	sprinklerCPT.Set(probability.Assignment{"Rain": "true"}, probability.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(probability.Assignment{"Rain": "false"}, probability.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(network.Node{
		Variable: "Sprinkler",
		Parents:  []probability.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	wetCPT := probability.NewCPT("WetGrass", []probability.Var{"Rain", "Sprinkler"})
	wetCPT.Set(probability.Assignment{"Rain": "true", "Sprinkler": "true"}, probability.Distribution{"true": 0.99, "false": 0.01})
	wetCPT.Set(probability.Assignment{"Rain": "true", "Sprinkler": "false"}, probability.Distribution{"true": 0.8, "false": 0.2})
	wetCPT.Set(probability.Assignment{"Rain": "false", "Sprinkler": "true"}, probability.Distribution{"true": 0.9, "false": 0.1})
	wetCPT.Set(probability.Assignment{"Rain": "false", "Sprinkler": "false"}, probability.Distribution{"true": 0.0, "false": 1.0})

	bn.AddNode(network.Node{
		Variable: "WetGrass",
		Parents:  []probability.Var{"Rain", "Sprinkler"},
		CPT:      wetCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	return bn
}

// makeMedicalNetwork builds a simple diagnostic network:
//
//	Disease → TestResult
//	Disease → Symptom
func makeMedicalNetwork() network.Network {
	bn := network.NewNetwork()

	diseaseCPT := probability.NewCPT("Disease", nil)
	diseaseCPT.Set(probability.Assignment{}, probability.Distribution{"present": 0.01, "absent": 0.99})

	bn.AddNode(network.Node{
		Variable: "Disease",
		CPT:      diseaseCPT,
		Outcomes: []probability.Outcome{"present", "absent"},
	})

	testCPT := probability.NewCPT("TestResult", []probability.Var{"Disease"})
	testCPT.Set(probability.Assignment{"Disease": "present"}, probability.Distribution{"positive": 0.95, "negative": 0.05})
	testCPT.Set(probability.Assignment{"Disease": "absent"}, probability.Distribution{"positive": 0.10, "negative": 0.90})

	bn.AddNode(network.Node{
		Variable: "TestResult",
		Parents:  []probability.Var{"Disease"},
		CPT:      testCPT,
		Outcomes: []probability.Outcome{"positive", "negative"},
	})

	symptomCPT := probability.NewCPT("Symptom", []probability.Var{"Disease"})
	symptomCPT.Set(probability.Assignment{"Disease": "present"}, probability.Distribution{"yes": 0.80, "no": 0.20})
	symptomCPT.Set(probability.Assignment{"Disease": "absent"}, probability.Distribution{"yes": 0.05, "no": 0.95})

	bn.AddNode(network.Node{
		Variable: "Symptom",
		Parents:  []probability.Var{"Disease"},
		CPT:      symptomCPT,
		Outcomes: []probability.Outcome{"yes", "no"},
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
			engine.WithEliminationOrder([]probability.Var{"Sprinkler", "WetGrass"}),
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
