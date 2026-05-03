package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/common/expressions"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// LoanApplication represents a loan application domain object.
type LoanApplication struct {
	Income        float64
	CreditScore   int
	HasCollateral bool
}

// LoanBinder implements decisions.Binder for LoanApplication.
type LoanBinder struct{}

func (b LoanBinder) BindDeductive(app LoanApplication) logic.Fact {
	facts := logic.Fact{
		"has_collateral": app.HasCollateral,
	}

	if app.Income > 50000 {
		facts["high_income"] = true
	}

	if app.CreditScore > 700 {
		facts["good_credit"] = true
	}

	return facts
}

func (b LoanBinder) BindBayesian(_ LoanApplication) evidence.EvidenceBase {
	return evidence.NewEvidenceBase()
}

func (b LoanBinder) BindFuzzy(_ LoanApplication) map[string]float64 {
	return nil
}

func (b LoanBinder) BindExpression(_ LoanApplication) expressions.Context {
	return nil
}

// Verify interface compliance.
var _ evaluate.Binder[LoanApplication] = LoanBinder{}

func ExampleService_deductive() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"loan-rules:v1": {
				Name:    "loan-rules",
				Version: "v1",
				Deductive: &schema.DeductiveConfig{
					Rules: []schema.DeductiveRuleDef{
						{
							Name:       "approve-high-income",
							Condition:  "high_income and good_credit",
							Conclusion: map[string]bool{"approved": true},
						},
						{
							Name:       "approve-collateral",
							Condition:  "good_credit and has_collateral",
							Conclusion: map[string]bool{"approved": true},
						},
						{
							Name:       "reject-low-credit",
							Condition:  "not good_credit",
							Conclusion: map[string]bool{"approved": false},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[LoanApplication](LoanBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[LoanApplication]{
		Domain: LoanApplication{
			Income:        75000,
			CreditScore:   750,
			HasCollateral: true,
		},
		RuleSetName:    "loan-rules",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Deductive,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	approved, ok := result.Outcome.Facts["approved"]
	if ok && approved {
		fmt.Println("Loan: APPROVED")
	} else {
		fmt.Println("Loan: REJECTED")
	}

	fmt.Printf("Paradigm: %s\n", result.Paradigm)

	// Output:
	// Loan: APPROVED
	// Paradigm: deductive
}
