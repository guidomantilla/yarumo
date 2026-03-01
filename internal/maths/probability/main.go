package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func main() {
	distributions()
	bayesTheorem()
	conditionalProbabilityTables()
	factorOperations()
}

// distributions shows how to create and work with discrete probability distributions.
func distributions() {
	fmt.Println("=== Distributions ===")

	// A coin flip: 50/50
	coin := probability.Distribution{
		"heads": 0.5,
		"tails": 0.5,
	}
	fmt.Println("Coin:", coin)
	fmt.Println("Valid:", probability.IsValid(coin))

	// A loaded die
	die := probability.Distribution{
		"1": 0.1,
		"2": 0.1,
		"3": 0.1,
		"4": 0.1,
		"5": 0.1,
		"6": 0.5, // loaded!
	}
	fmt.Println("Die:", die)
	fmt.Println("Valid:", probability.IsValid(die))

	// Entropy: measures "surprise" or uncertainty
	// Coin has maximum entropy (most uncertain), loaded die has less
	coinEntropy, _ := probability.Entropy(coin)
	dieEntropy, _ := probability.Entropy(die)
	fmt.Printf("Coin entropy: %.3f bits\n", coinEntropy)
	fmt.Printf("Die entropy:  %.3f bits\n", dieEntropy)

	// Normalize an unnormalized distribution
	raw := probability.Distribution{"a": 2, "b": 3, "c": 5}
	normalized, _ := probability.Normalize(raw)
	fmt.Println("Normalized:", normalized)

	// Complement: P(not event) = 1 - P(event)
	fmt.Println("P(not heads):", probability.Complement(0.5))
	fmt.Println()
}

// bayesTheorem shows Bayes' theorem and related probability calculations.
func bayesTheorem() {
	fmt.Println("=== Bayes' Theorem ===")

	// Medical test example:
	// - 1% of people have disease X        P(D) = 0.01
	// - Test is 95% accurate for sick       P(+|D) = 0.95
	// - Test has 5% false positive rate     P(+|~D) = 0.05

	pDisease := probability.Prob(0.01)       // P(D) = prior
	pPosGivenDisease := probability.Prob(0.95) // P(+|D) = sensitivity
	pPosGivenHealthy := probability.Prob(0.05) // P(+|~D) = false positive rate

	// P(+) = P(+|D)*P(D) + P(+|~D)*P(~D)  (total probability)
	pPos := probability.TotalProbability(
		[]probability.Prob{pDisease, probability.Complement(pDisease)},
		[]probability.Prob{pPosGivenDisease, pPosGivenHealthy},
	)
	fmt.Printf("P(positive test): %.4f\n", pPos)

	// P(D|+) = P(+|D)*P(D) / P(+)  (Bayes)
	pDiseaseGivenPos := probability.Bayes(pDisease, pPosGivenDisease, pPos)
	fmt.Printf("P(disease | positive test): %.4f\n", pDiseaseGivenPos)
	fmt.Println("=> Even with a positive test, only ~16% chance of disease!")

	// Chain rule: P(A,B,C) = P(A) * P(B|A) * P(C|A,B)
	joint := probability.ChainRule(0.5, 0.3, 0.8)
	fmt.Printf("P(A,B,C) via chain rule: %.3f\n", joint)

	// Independent events: P(A and B) = P(A) * P(B)
	pBoth := probability.Independent(0.5, 0.3)
	fmt.Printf("P(A and B) independent: %.3f\n", pBoth)
	fmt.Println()
}

// conditionalProbabilityTables shows CPTs used in Bayesian networks.
func conditionalProbabilityTables() {
	fmt.Println("=== Conditional Probability Tables ===")

	// Model: P(Sprinkler | Season)
	// In summer, sprinkler is likely ON. In winter, likely OFF.
	cpt := probability.NewCPT("Sprinkler", []probability.Var{"Season"})

	cpt.Set(
		probability.Assignment{"Season": "summer"},
		probability.Distribution{"on": 0.8, "off": 0.2},
	)
	cpt.Set(
		probability.Assignment{"Season": "winter"},
		probability.Distribution{"on": 0.1, "off": 0.9},
	)

	// Lookup: what's the sprinkler distribution in summer?
	summer, _ := cpt.Lookup(probability.Assignment{"Season": "summer"})
	fmt.Println("P(Sprinkler | summer):", summer)

	winter, _ := cpt.Lookup(probability.Assignment{"Season": "winter"})
	fmt.Println("P(Sprinkler | winter):", winter)

	// Validate: all entries must be valid distributions
	err := cpt.Validate()
	fmt.Println("CPT valid:", err == nil)
	fmt.Println()
}

// factorOperations shows factor algebra used by variable elimination.
func factorOperations() {
	fmt.Println("=== Factor Operations ===")

	// Factor: a table mapping variable assignments to probabilities.
	// Used as intermediate representation during inference.

	// Factor for P(A)
	fA := probability.NewFactor(
		[]probability.Var{"A"},
		map[string]probability.Prob{
			"A=true":  0.6,
			"A=false": 0.4,
		},
	)
	fmt.Println("Factor P(A):", fA)

	// Factor for P(B|A)
	fBA := probability.NewFactor(
		[]probability.Var{"A", "B"},
		map[string]probability.Prob{
			"A=true,B=true":   0.9,
			"A=true,B=false":  0.1,
			"A=false,B=true":  0.2,
			"A=false,B=false": 0.8,
		},
	)
	fmt.Println("Factor P(B|A):", fBA)

	// Multiply: combine factors (join step in variable elimination)
	product := probability.Multiply(fA, fBA)
	fmt.Println("P(A) * P(B|A):", product)

	// SumOut: marginalize a variable (eliminate step)
	// Sum out A to get P(B)
	pB := probability.SumOut(product, "A")
	fmt.Println("Sum out A => P(B):", pB)

	// Normalize: make probabilities sum to 1
	pBNorm := probability.NormalizeFactor(pB)
	fmt.Println("Normalized P(B):", pBNorm)

	// Restrict: fix a variable to a specific value (evidence)
	// Given A=true, what is P(B)?
	restricted := probability.Restrict(fBA, "A", "true")
	fmt.Println("P(B | A=true):", restricted)
}
