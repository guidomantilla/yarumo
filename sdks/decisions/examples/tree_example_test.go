package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// RiskAssessment represents a risk assessment domain object.
type RiskAssessment struct {
	Income      float64
	CreditScore int
	DebtRatio   float64
}

// RiskAssessmentBinder implements models.ExpressionBinder for RiskAssessment.
type RiskAssessmentBinder struct{}

func (b RiskAssessmentBinder) BindExpression(d RiskAssessment) expressions.Context {
	return expressions.Context{
		"income":       d.Income,
		"credit_score": d.CreditScore,
		"debt_ratio":   d.DebtRatio,
	}
}

// Verify interface compliance.
var _ evaluate.ExpressionBinder[RiskAssessment] = RiskAssessmentBinder{}

func ExampleService_tree() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"risk-tree:v1": {
				Name:    "risk-tree",
				Version: "v1",
				Tree: &schema.TreeConfig{
					Root: schema.TreeNodeDef{
						Condition: "credit_score > 700",
						True: &schema.TreeNodeDef{
							Condition: "debt_ratio < 0.3",
							True:      &schema.TreeNodeDef{Output: map[string]any{"risk": "low", "limit": 50000}},
							False:     &schema.TreeNodeDef{Output: map[string]any{"risk": "medium", "limit": 25000}},
						},
						False: &schema.TreeNodeDef{
							Condition: "income > 50000",
							True:      &schema.TreeNodeDef{Output: map[string]any{"risk": "medium", "limit": 15000}},
							False:     &schema.TreeNodeDef{Output: map[string]any{"risk": "high", "limit": 5000}},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[RiskAssessment](RiskAssessmentBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[RiskAssessment]{
		Domain: RiskAssessment{
			Income:      75000,
			CreditScore: 750,
			DebtRatio:   0.2,
		},
		RuleSetName:    "risk-tree",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Tree,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Risk: %s\n", result.Outcome.Tree.Outputs["risk"])
	fmt.Printf("Paradigm: %s\n", result.Paradigm)

	// Output:
	// Risk: low
	// Paradigm: tree
}
