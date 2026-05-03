package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func ExampleCascadePipeline() {
	complianceRuleSet := &schema.RuleSet{
		Name: "compliance",
		Deductive: &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "has-invoicing",
					Condition:  "facturacion",
					Conclusion: map[string]bool{"cumple_facturacion": true},
				},
				{
					Name:       "has-retention",
					Condition:  "retencion",
					Conclusion: map[string]bool{"cumple_retencion": true},
				},
			},
		},
	}

	riskRuleSet := &schema.RuleSet{
		Name: "risk",
		Bayesian: &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "compliance",
					Outcomes: []string{"compliant", "non_compliant"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"compliant": 0.7, "non_compliant": 0.3}},
					},
				},
				{
					Variable: "audit_risk",
					Parents:  []string{"compliance"},
					Outcomes: []string{"high", "low"},
					CPT: []schema.CPTRow{
						{
							Given:         map[string]string{"compliance": "compliant"},
							Probabilities: map[string]float64{"high": 0.1, "low": 0.9},
						},
						{
							Given:         map[string]string{"compliance": "non_compliant"},
							Probabilities: map[string]float64{"high": 0.8, "low": 0.2},
						},
					},
				},
			},
		},
	}

	stages := []evaluate.CascadeStage{
		{Name: "compliance-check", Paradigm: evaluate.Deductive, RuleSet: complianceRuleSet},
		{Name: "risk-assessment", Paradigm: evaluate.Bayesian, RuleSet: riskRuleSet, Query: "audit_risk"},
	}

	converter := func(result evaluate.Result) (any, error) {
		eb := evidence.NewEvidenceBase()

		compliant, ok := result.Outcome.Facts[logic.Var("cumple_facturacion")]
		if ok && compliant {
			eb.Observe(stats.Var("compliance"), stats.Outcome("compliant"))
		} else {
			eb.Observe(stats.Var("compliance"), stats.Outcome("non_compliant"))
		}

		return eb, nil
	}

	pipeline := evaluate.NewCascadePipeline(stages, []evaluate.StageConverter{converter})

	result, err := pipeline.Execute(context.Background(), logic.Fact{
		"facturacion": true,
		"retencion":   true,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Stages: %d\n", len(result.Stages))
	fmt.Printf("Final paradigm: %s\n", result.Final.Paradigm)

	highRisk := float64(result.Final.Outcome.Distribution["high"])
	if highRisk < 0.5 {
		fmt.Println("Risk: LOW")
	} else {
		fmt.Println("Risk: HIGH")
	}

	// Output:
	// Stages: 2
	// Final paradigm: bayesian
	// Risk: LOW
}
