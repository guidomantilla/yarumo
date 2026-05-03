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

// TipInput represents restaurant service quality for tip calculation.
type TipInput struct {
	ServiceQuality float64
	FoodQuality    float64
}

// TipBinder implements decisions.Binder for TipInput.
type TipBinder struct{}

func (b TipBinder) BindDeductive(_ TipInput) logic.Fact {
	return nil
}

func (b TipBinder) BindBayesian(_ TipInput) evidence.EvidenceBase {
	return evidence.NewEvidenceBase()
}

func (b TipBinder) BindFuzzy(t TipInput) map[string]float64 {
	return map[string]float64{
		"service": t.ServiceQuality,
		"food":    t.FoodQuality,
	}
}

func (b TipBinder) BindExpression(_ TipInput) expressions.Context {
	return nil
}

// Verify interface compliance.
var _ evaluate.Binder[TipInput] = TipBinder{}

func ExampleService_fuzzy() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"tipping:v1": {
				Name:    "tipping",
				Version: "v1",
				Fuzzy: &schema.FuzzyConfig{
					InputVars: []schema.FuzzyVarDef{
						{
							Name: "service",
							Min:  0,
							Max:  10,
							Terms: []schema.FuzzyTermDef{
								{Name: "poor", Type: "triangular", Params: []float64{0, 0, 5}},
								{Name: "good", Type: "triangular", Params: []float64{2, 5, 8}},
								{Name: "excellent", Type: "triangular", Params: []float64{5, 10, 10}},
							},
						},
						{
							Name: "food",
							Min:  0,
							Max:  10,
							Terms: []schema.FuzzyTermDef{
								{Name: "bad", Type: "triangular", Params: []float64{0, 0, 5}},
								{Name: "good", Type: "triangular", Params: []float64{2, 5, 8}},
								{Name: "great", Type: "triangular", Params: []float64{5, 10, 10}},
							},
						},
					},
					OutputVars: []schema.FuzzyVarDef{
						{
							Name: "tip",
							Min:  0,
							Max:  30,
							Terms: []schema.FuzzyTermDef{
								{Name: "low", Type: "triangular", Params: []float64{0, 0, 15}},
								{Name: "medium", Type: "triangular", Params: []float64{5, 15, 25}},
								{Name: "high", Type: "triangular", Params: []float64{15, 30, 30}},
							},
						},
					},
					Rules: []schema.FuzzyRuleDef{
						{
							Name:       "poor-service",
							Conditions: []schema.FuzzyConditionDef{{Variable: "service", Term: "poor"}},
							Consequent: schema.FuzzyConsequentDef{Variable: "tip", Term: "low"},
						},
						{
							Name:       "good-service-food",
							Conditions: []schema.FuzzyConditionDef{{Variable: "service", Term: "good"}, {Variable: "food", Term: "good"}},
							Consequent: schema.FuzzyConsequentDef{Variable: "tip", Term: "medium"},
						},
						{
							Name:       "excellent-service",
							Conditions: []schema.FuzzyConditionDef{{Variable: "service", Term: "excellent"}},
							Consequent: schema.FuzzyConsequentDef{Variable: "tip", Term: "high"},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[TipInput](TipBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[TipInput]{
		Domain: TipInput{
			ServiceQuality: 8,
			FoodQuality:    7,
		},
		RuleSetName:    "tipping",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Fuzzy,
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Paradigm: %s\n", result.Paradigm)

	tip := result.Outcome.Outputs["tip"]

	switch {
	case tip > 20:
		fmt.Println("Tip: generous")
	case tip > 10:
		fmt.Println("Tip: moderate")
	default:
		fmt.Println("Tip: low")
	}

	// Output:
	// Paradigm: fuzzy
	// Tip: generous
}
