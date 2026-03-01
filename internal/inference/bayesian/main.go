package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/engine"
	"github.com/guidomantilla/yarumo/inference/bayesian/evidence"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

func main() {
	buildNetwork()
	queryWithEvidence()
	comparingAlgorithms()
	medicalDiagnosis()
	explainingAway()
}

// buildNetwork shows how to construct a Bayesian network from scratch.
func buildNetwork() {
	fmt.Println("=== Building a Bayesian Network ===")

	bn := network.NewNetwork()

	// Root node: Rain (no parents)
	// P(Rain=true) = 0.2, P(Rain=false) = 0.8
	rainCPT := probability.NewCPT("Rain", nil)
	rainCPT.Set(probability.Assignment{}, probability.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(network.Node{
		Variable: "Rain",
		CPT:      rainCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	// Child node: Sprinkler (depends on Rain)
	// If raining, sprinkler is usually OFF. If not raining, might be ON.
	sprinklerCPT := probability.NewCPT("Sprinkler", []probability.Var{"Rain"})
	sprinklerCPT.Set(probability.Assignment{"Rain": "true"}, probability.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(probability.Assignment{"Rain": "false"}, probability.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(network.Node{
		Variable: "Sprinkler",
		Parents:  []probability.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	// Child node: WetGrass (depends on Rain AND Sprinkler)
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

	// Validate the network (checks for cycles, CPT consistency)
	err := bn.Validate()
	fmt.Println("Network valid:", err == nil)

	// Topological order: parents always come before children
	order := bn.TopologicalOrder()
	fmt.Println("Topological order:", order)

	// Inspect structure
	fmt.Println("Children of Rain:", bn.Children("Rain"))
	fmt.Println("Parents of WetGrass:", bn.Parents("WetGrass"))
	fmt.Println()
}

// queryWithEvidence shows how to compute posterior probabilities.
func queryWithEvidence() {
	fmt.Println("=== Query with Evidence ===")

	bn := makeRainNetwork()
	eng := engine.NewEngine()

	// Prior probability: P(Rain) with no evidence
	ev := evidence.NewEvidenceBase()
	prior := eng.Query(bn, ev, "Rain")
	fmt.Printf("P(Rain=true) prior:              %.4f\n", prior.Posterior["true"])

	// Observe that the grass is wet: how does this change P(Rain)?
	ev.Observe("WetGrass", "true")
	posterior := eng.Query(bn, ev, "Rain")
	fmt.Printf("P(Rain=true | WetGrass=true):    %.4f\n", posterior.Posterior["true"])
	fmt.Println("=> Wet grass makes rain more likely!")

	// Evidence operations
	fmt.Println()
	fmt.Println("Evidence base operations:")
	ev2 := evidence.NewEvidenceBase()
	ev2.Observe("Rain", "true")
	ev2.Observe("Sprinkler", "false")
	fmt.Printf("  Observations: %d\n", ev2.Len())

	outcome, ok := ev2.Get("Rain")
	fmt.Printf("  Rain observed: %v = %s\n", ok, outcome)

	// Retract evidence
	ev2.Retract("Sprinkler")
	fmt.Printf("  After retract: %d observations\n", ev2.Len())

	// Clone
	clone := ev2.Clone()
	clone.Observe("WetGrass", "true")
	fmt.Printf("  Original: %d, Clone: %d\n", ev2.Len(), clone.Len())
	fmt.Println()
}

// comparingAlgorithms shows that enumeration and variable elimination give the same results.
func comparingAlgorithms() {
	fmt.Println("=== Comparing Algorithms ===")

	bn := makeRainNetwork()
	ev := evidence.NewEvidenceBase()
	ev.Observe("WetGrass", "true")

	// Enumeration: exact inference by summing over all combinations
	enumEng := engine.NewEngine() // default: Enumeration
	enumResult := enumEng.Query(bn, ev, "Rain")

	// Variable elimination: exact inference using factor operations (faster)
	veEng := engine.NewEngine(engine.WithAlgorithm(engine.VariableElimination))
	veResult := veEng.Query(bn, ev, "Rain")

	fmt.Println("P(Rain=true | WetGrass=true):")
	fmt.Printf("  Enumeration:          %.4f\n", enumResult.Posterior["true"])
	fmt.Printf("  Variable Elimination: %.4f\n", veResult.Posterior["true"])
	fmt.Println("  Both algorithms agree!")

	// Custom elimination order (can affect performance, not results)
	customEng := engine.NewEngine(
		engine.WithAlgorithm(engine.VariableElimination),
		engine.WithEliminationOrder([]probability.Var{"Sprinkler", "WetGrass"}),
	)
	customResult := customEng.Query(bn, ev, "Rain")
	fmt.Printf("  Custom order:         %.4f\n", customResult.Posterior["true"])

	// Inspect the trace
	fmt.Println()
	fmt.Println("Inference trace:")
	fmt.Println(veResult.Trace.String())
}

// medicalDiagnosis shows a practical diagnostic Bayesian network.
func medicalDiagnosis() {
	fmt.Println("=== Medical Diagnosis ===")

	bn := network.NewNetwork()

	// Disease: rare (1% prevalence)
	diseaseCPT := probability.NewCPT("Disease", nil)
	diseaseCPT.Set(probability.Assignment{}, probability.Distribution{"present": 0.01, "absent": 0.99})

	bn.AddNode(network.Node{
		Variable: "Disease",
		CPT:      diseaseCPT,
		Outcomes: []probability.Outcome{"present", "absent"},
	})

	// Test: 95% sensitivity, 10% false positive rate
	testCPT := probability.NewCPT("Test", []probability.Var{"Disease"})
	testCPT.Set(probability.Assignment{"Disease": "present"}, probability.Distribution{"positive": 0.95, "negative": 0.05})
	testCPT.Set(probability.Assignment{"Disease": "absent"}, probability.Distribution{"positive": 0.10, "negative": 0.90})

	bn.AddNode(network.Node{
		Variable: "Test",
		Parents:  []probability.Var{"Disease"},
		CPT:      testCPT,
		Outcomes: []probability.Outcome{"positive", "negative"},
	})

	// Symptom: 80% if disease, 5% if no disease
	symptomCPT := probability.NewCPT("Symptom", []probability.Var{"Disease"})
	symptomCPT.Set(probability.Assignment{"Disease": "present"}, probability.Distribution{"yes": 0.80, "no": 0.20})
	symptomCPT.Set(probability.Assignment{"Disease": "absent"}, probability.Distribution{"yes": 0.05, "no": 0.95})

	bn.AddNode(network.Node{
		Variable: "Symptom",
		Parents:  []probability.Var{"Disease"},
		CPT:      symptomCPT,
		Outcomes: []probability.Outcome{"yes", "no"},
	})

	eng := engine.NewEngine()

	// Just a positive test
	ev1 := evidence.NewEvidenceBase()
	ev1.Observe("Test", "positive")
	result1 := eng.Query(bn, ev1, "Disease")
	fmt.Printf("P(Disease | positive test):                    %.4f\n", result1.Posterior["present"])

	// Positive test AND symptom present
	ev2 := evidence.NewEvidenceBase()
	ev2.Observe("Test", "positive")
	ev2.Observe("Symptom", "yes")
	result2 := eng.Query(bn, ev2, "Disease")
	fmt.Printf("P(Disease | positive test + symptom):          %.4f\n", result2.Posterior["present"])
	fmt.Println("=> More evidence increases diagnostic confidence!")

	// Negative test despite symptom
	ev3 := evidence.NewEvidenceBase()
	ev3.Observe("Test", "negative")
	ev3.Observe("Symptom", "yes")
	result3 := eng.Query(bn, ev3, "Disease")
	fmt.Printf("P(Disease | negative test + symptom):          %.4f\n", result3.Posterior["present"])
	fmt.Println("=> Negative test strongly reduces probability even with symptom")
	fmt.Println()
}

// explainingAway shows a classic Bayesian reasoning pattern:
// observing one cause of an effect reduces the probability of another cause.
func explainingAway() {
	fmt.Println("=== Explaining Away ===")
	fmt.Println("(Observing one cause reduces the need for alternative causes)")

	bn := makeRainNetwork()
	eng := engine.NewEngine()

	// Grass is wet. How likely is rain?
	evWet := evidence.NewEvidenceBase()
	evWet.Observe("WetGrass", "true")
	rainGivenWet := eng.Query(bn, evWet, "Rain")

	// Grass is wet AND sprinkler is on. How likely is rain now?
	evWetAndSprinkler := evidence.NewEvidenceBase()
	evWetAndSprinkler.Observe("WetGrass", "true")
	evWetAndSprinkler.Observe("Sprinkler", "true")
	rainGivenBoth := eng.Query(bn, evWetAndSprinkler, "Rain")

	fmt.Printf("P(Rain | WetGrass):                            %.4f\n", rainGivenWet.Posterior["true"])
	fmt.Printf("P(Rain | WetGrass + Sprinkler):                %.4f\n", rainGivenBoth.Posterior["true"])
	fmt.Println("=> Knowing the sprinkler caused wet grass 'explains away' rain")
	fmt.Println("   — the alternative cause is less needed!")
}

// makeRainNetwork builds the classic sprinkler Bayesian network.
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
