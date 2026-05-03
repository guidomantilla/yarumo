package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// LoanDecision represents a loan decision domain object.
type LoanDecision struct {
	Income      float64
	CreditScore int
	Age         int
}

// LoanDecisionBinder implements models.ExpressionBinder for LoanDecision.
type LoanDecisionBinder struct{}

func (b LoanDecisionBinder) BindExpression(d LoanDecision) expressions.Context {
	return expressions.Context{
		"income":       d.Income,
		"credit_score": d.CreditScore,
		"age":          d.Age,
	}
}

// Verify interface compliance.
var _ evaluate.ExpressionBinder[LoanDecision] = LoanDecisionBinder{}

func ExampleService_table() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"loan-table:v1": {
				Name:    "loan-table",
				Version: "v1",
				Table: &schema.TableConfig{
					HitPolicy: "first",
					Rules: []schema.TableRuleDef{
						{
							Name:       "approve-high-income",
							Conditions: []string{"income > 75000", "credit_score > 700"},
							Outputs:    map[string]any{"decision": "approved", "tier": "premium"},
						},
						{
							Name:       "approve-standard",
							Conditions: []string{"income > 40000", "credit_score > 600"},
							Outputs:    map[string]any{"decision": "approved", "tier": "standard"},
						},
						{
							Name:       "reject",
							Conditions: []string{"credit_score <= 600"},
							Outputs:    map[string]any{"decision": "rejected"},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[LoanDecision](LoanDecisionBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[LoanDecision]{
		Domain: LoanDecision{
			Income:      80000,
			CreditScore: 750,
			Age:         35,
		},
		RuleSetName:    "loan-table",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Table,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Decision: %s\n", result.Outcome.Table.Outputs["decision"])
	fmt.Printf("Paradigm: %s\n", result.Paradigm)

	// Output:
	// Decision: approved
	// Paradigm: table
}
