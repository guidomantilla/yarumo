package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/fuzzy"
)

func main() {
	membershipFunctions()
	fuzzyOperators()
	fuzzifyAndDefuzzify()
	temperatureControlExample()
}

// membershipFunctions shows how to define and evaluate membership functions.
func membershipFunctions() {
	fmt.Println("=== Membership Functions ===")

	// Triangular: peaks at center, linearly falls to zero at edges
	// "medium temperature" peaks at 20°C, zero below 10 and above 30
	medium := fuzzy.Triangular(10, 20, 30)
	fmt.Printf("Triangular(10,20,30):\n")
	fmt.Printf("  15°C = %.2f (somewhat medium)\n", medium(15))
	fmt.Printf("  20°C = %.2f (fully medium)\n", medium(20))
	fmt.Printf("  25°C = %.2f (somewhat medium)\n", medium(25))
	fmt.Printf("   5°C = %.2f (not medium)\n", medium(5))

	// Trapezoidal: flat top between b and c
	// "comfortable temperature" fully comfortable between 18-24, tapers at edges
	comfortable := fuzzy.Trapezoidal(14, 18, 24, 28)
	fmt.Printf("Trapezoidal(14,18,24,28):\n")
	fmt.Printf("  16°C = %.2f\n", comfortable(16))
	fmt.Printf("  21°C = %.2f (fully comfortable)\n", comfortable(21))
	fmt.Printf("  26°C = %.2f\n", comfortable(26))

	// Gaussian: bell curve
	warm := fuzzy.Gaussian(30, 5)
	fmt.Printf("Gaussian(center=30, sigma=5):\n")
	fmt.Printf("  25°C = %.2f\n", warm(25))
	fmt.Printf("  30°C = %.2f\n", warm(30))
	fmt.Printf("  35°C = %.2f\n", warm(35))

	// Sigmoid: S-curve, useful for "high" or "low" categories
	hot := fuzzy.Sigmoid(35, 0.5)
	fmt.Printf("Sigmoid(center=35, slope=0.5):\n")
	fmt.Printf("  25°C = %.2f\n", hot(25))
	fmt.Printf("  35°C = %.2f\n", hot(35))
	fmt.Printf("  45°C = %.2f\n", hot(45))
	fmt.Println()
}

// fuzzyOperators shows t-norms (AND), t-conorms (OR), and complement (NOT).
func fuzzyOperators() {
	fmt.Println("=== Fuzzy Operators ===")

	a := fuzzy.Degree(0.7)
	b := fuzzy.Degree(0.4)
	fmt.Printf("a = %.1f, b = %.1f\n", a, b)

	// AND (t-norms): different ways to combine "both"
	fmt.Printf("Min AND:         %.1f (standard)\n", fuzzy.Min(a, b))
	fmt.Printf("Product AND:     %.2f (softer)\n", fuzzy.Product(a, b))
	fmt.Printf("Lukasiewicz AND: %.1f (strictest)\n", fuzzy.Lukasiewicz(a, b))

	// OR (t-conorms): different ways to combine "either"
	fmt.Printf("Max OR:          %.1f (standard)\n", fuzzy.Max(a, b))
	fmt.Printf("Probabilistic OR:%.2f (softer)\n", fuzzy.ProbabilisticSum(a, b))
	fmt.Printf("Bounded OR:      %.1f (most generous)\n", fuzzy.BoundedSum(a, b))

	// NOT (complement)
	fmt.Printf("NOT a:           %.1f\n", fuzzy.Complement(a))
	fmt.Println()
}

// fuzzifyAndDefuzzify shows the full pipeline: crisp -> fuzzy -> crisp.
func fuzzifyAndDefuzzify() {
	fmt.Println("=== Fuzzify and Defuzzify ===")

	// Define a fuzzy set
	cold := fuzzy.Triangular(0, 0, 20)

	// Fuzzify: convert a crisp value to a fuzzy degree
	temp := 12.0
	degree := fuzzy.Fuzzify(cold, temp)
	fmt.Printf("%.0f°C is 'cold' with degree %.2f\n", temp, degree)

	// Clip: limit the output membership function by the firing degree
	clipped := fuzzy.Clip(cold, degree)
	fmt.Printf("Clipped cold at degree %.2f:\n", degree)
	fmt.Printf("  0°C  = %.2f\n", clipped(0))
	fmt.Printf("  10°C = %.2f\n", clipped(10))
	fmt.Printf("  15°C = %.2f\n", clipped(15))

	// Sample: discretize a membership function into points
	xs, ys, _ := fuzzy.Sample(cold, 0, 25, 6)
	fmt.Println("Sampled cold:")
	for i := range xs {
		fmt.Printf("  %.1f°C = %.2f\n", xs[i], ys[i])
	}

	// Defuzzify: convert fuzzy output back to a crisp value
	centroid := fuzzy.Centroid(xs, ys)
	fmt.Printf("Centroid: %.2f°C\n", centroid)

	bisector := fuzzy.Bisector(xs, ys)
	fmt.Printf("Bisector: %.2f°C\n", bisector)

	mom := fuzzy.MeanOfMax(xs, ys)
	fmt.Printf("Mean of Max: %.2f°C\n", mom)
	fmt.Println()
}

// temperatureControlExample shows a mini fuzzy control system.
// Input: temperature (crisp) -> Output: fan speed (crisp)
func temperatureControlExample() {
	fmt.Println("=== Temperature Control Example ===")

	// Define input membership functions for temperature
	cold := fuzzy.Triangular(0, 0, 20)
	warm := fuzzy.Triangular(15, 25, 35)
	hot := fuzzy.Triangular(30, 40, 40)

	// Define output membership functions for fan speed
	low := fuzzy.Triangular(0, 0, 50)
	medium := fuzzy.Triangular(25, 50, 75)
	high := fuzzy.Triangular(50, 100, 100)

	// Input: current temperature
	temp := 28.0
	fmt.Printf("Temperature: %.0f°C\n", temp)

	// Fuzzify input
	coldDeg := fuzzy.Fuzzify(cold, temp)
	warmDeg := fuzzy.Fuzzify(warm, temp)
	hotDeg := fuzzy.Fuzzify(hot, temp)
	fmt.Printf("  cold=%.2f  warm=%.2f  hot=%.2f\n", coldDeg, warmDeg, hotDeg)

	// Apply rules (Mamdani-style: clip output by rule firing degree)
	// Rule 1: IF cold THEN fan=low
	// Rule 2: IF warm THEN fan=medium
	// Rule 3: IF hot THEN fan=high
	rule1 := fuzzy.Clip(low, coldDeg)
	rule2 := fuzzy.Clip(medium, warmDeg)
	rule3 := fuzzy.Clip(high, hotDeg)

	// Aggregate outputs (max of all rule outputs)
	aggregated := fuzzy.AggregateMax(rule1, rule2, rule3)

	// Sample and defuzzify
	xs, ys, _ := fuzzy.Sample(aggregated, 0, 100, 101)
	fanSpeed := fuzzy.Centroid(xs, ys)
	fmt.Printf("  Fan speed: %.1f%%\n", fanSpeed)
}
