package examples

import (
	"fmt"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
	"github.com/guidomantilla/yarumo/decisions/core/validate"
	"github.com/guidomantilla/yarumo/compute/math/logic/sat"
)

func ExampleValidator_deductive() {
	v := validate.NewValidator(sat.Solver())

	config := &schema.DeductiveConfig{
		Rules: []schema.DeductiveRuleDef{
			{
				Name:       "approve",
				Condition:  "eligible and good_credit",
				Conclusion: map[string]bool{"approved": true},
			},
			{
				Name:       "reject",
				Condition:  "not eligible",
				Conclusion: map[string]bool{"approved": false},
			},
			{
				Name:       "double-neg",
				Condition:  "not not eligible",
				Conclusion: map[string]bool{"reviewed": true},
			},
		},
	}

	report := v.ValidateDeductive(config)

	fmt.Printf("Parsed: %d\n", report.Parsed)
	fmt.Printf("Valid: %v\n", report.Valid)
	fmt.Printf("Contradictions: %d\n", len(report.Contradictions))

	if len(report.Simplified) > 0 {
		fmt.Printf("Simplification: %s -> %s\n", report.Simplified[0].Original, report.Simplified[0].Simplified)
	}

	// Output:
	// Parsed: 3
	// Valid: true
	// Contradictions: 0
	// Simplification: ¬¬eligible -> eligible
}

func ExampleValidator_bayesian() {
	v := validate.NewValidator(sat.Solver())

	config := &schema.BayesianConfig{
		Nodes: []schema.BayesianNodeDef{
			{
				Variable: "rain",
				Outcomes: []string{"yes", "no"},
				CPT: []schema.CPTRow{
					{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
				},
			},
			{
				Variable: "wet_grass",
				Parents:  []string{"rain"},
				Outcomes: []string{"yes", "no"},
				CPT: []schema.CPTRow{
					{Given: map[string]string{"rain": "yes"}, Probabilities: map[string]float64{"yes": 0.9, "no": 0.1}},
					{Given: map[string]string{"rain": "no"}, Probabilities: map[string]float64{"yes": 0.2, "no": 0.8}},
				},
			},
		},
	}

	report := v.ValidateBayesian(config)

	fmt.Printf("Parsed: %d\n", report.Parsed)
	fmt.Printf("Valid: %v\n", report.Valid)
	fmt.Printf("Errors: %d\n", len(report.Errors))

	// Output:
	// Parsed: 2
	// Valid: true
	// Errors: 0
}
