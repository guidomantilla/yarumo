package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// CreditApplicant represents a credit risk applicant.
type CreditApplicant struct {
	Age            int
	Income         float64
	YearsEmployed  int
	HasBankAccount bool
}

// CreditApplicantBinder implements models.ExpressionBinder for CreditApplicant.
type CreditApplicantBinder struct{}

func (b CreditApplicantBinder) BindExpression(d CreditApplicant) expressions.Context {
	return expressions.Context{
		"age":              d.Age,
		"income":           d.Income,
		"years_employed":   d.YearsEmployed,
		"has_bank_account": d.HasBankAccount,
	}
}

// Verify interface compliance.
var _ evaluate.ExpressionBinder[CreditApplicant] = CreditApplicantBinder{}

func ExampleService_scorecard() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"credit-score:v1": {
				Name:    "credit-score",
				Version: "v1",
				Scorecard: &schema.ScorecardConfig{
					BaseScore: 300,
					Attributes: []schema.ScorecardAttributeDef{
						{
							Name:   "age",
							Weight: 1.0,
							Bins: []schema.ScorecardBinDef{
								{Condition: "age > 40", Points: 50},
								{Condition: "age > 25", Points: 30},
								{Condition: "age > 18", Points: 10},
							},
						},
						{
							Name:   "income",
							Weight: 1.5,
							Bins: []schema.ScorecardBinDef{
								{Condition: "income > 80000", Points: 100},
								{Condition: "income > 50000", Points: 60},
								{Condition: "income > 30000", Points: 30},
							},
						},
						{
							Name:   "employment",
							Weight: 1.0,
							Bins: []schema.ScorecardBinDef{
								{Condition: "years_employed > 5", Points: 40},
								{Condition: "years_employed > 2", Points: 20},
							},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[CreditApplicant](CreditApplicantBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[CreditApplicant]{
		Domain: CreditApplicant{
			Age:            35,
			Income:         65000,
			YearsEmployed:  8,
			HasBankAccount: true,
		},
		RuleSetName:    "credit-score",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Scorecard,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Score: %.0f\n", result.Outcome.Score.TotalScore)
	fmt.Printf("Paradigm: %s\n", result.Paradigm)

	// Output:
	// Score: 460
	// Paradigm: scorecard
}
