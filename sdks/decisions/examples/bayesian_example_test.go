package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/common/expressions"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/decisions/core/evaluate"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// PatientSymptoms represents a patient's symptoms for diagnosis.
type PatientSymptoms struct {
	HasFever   bool
	HasCough   bool
	HasFatigue bool
}

// PatientBinder implements decisions.Binder for PatientSymptoms.
type PatientBinder struct{}

func (b PatientBinder) BindDeductive(_ PatientSymptoms) logic.Fact {
	return nil
}

func (b PatientBinder) BindBayesian(s PatientSymptoms) evidence.EvidenceBase {
	eb := evidence.NewEvidenceBase()

	if s.HasFever {
		eb.Observe(stats.Var("fever"), stats.Outcome("yes"))
	} else {
		eb.Observe(stats.Var("fever"), stats.Outcome("no"))
	}

	if s.HasCough {
		eb.Observe(stats.Var("cough"), stats.Outcome("yes"))
	} else {
		eb.Observe(stats.Var("cough"), stats.Outcome("no"))
	}

	return eb
}

func (b PatientBinder) BindFuzzy(_ PatientSymptoms) map[string]float64 {
	return nil
}

func (b PatientBinder) BindExpression(_ PatientSymptoms) expressions.Context {
	return nil
}

// Verify interface compliance.
var _ evaluate.Binder[PatientSymptoms] = PatientBinder{}

func ExampleService_bayesian() {
	repo := &memoryRepo{
		rulesets: map[string]*schema.RuleSet{
			"diagnosis:v1": {
				Name:    "diagnosis",
				Version: "v1",
				Bayesian: &schema.BayesianConfig{
					Nodes: []schema.BayesianNodeDef{
						{
							Variable: "fever",
							Outcomes: []string{"yes", "no"},
							CPT: []schema.CPTRow{
								{Probabilities: map[string]float64{"yes": 0.2, "no": 0.8}},
							},
						},
						{
							Variable: "cough",
							Outcomes: []string{"yes", "no"},
							CPT: []schema.CPTRow{
								{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
							},
						},
						{
							Variable: "flu",
							Parents:  []string{"fever", "cough"},
							Outcomes: []string{"yes", "no"},
							CPT: []schema.CPTRow{
								{Given: map[string]string{"fever": "yes", "cough": "yes"}, Probabilities: map[string]float64{"yes": 0.9, "no": 0.1}},
								{Given: map[string]string{"fever": "yes", "cough": "no"}, Probabilities: map[string]float64{"yes": 0.6, "no": 0.4}},
								{Given: map[string]string{"fever": "no", "cough": "yes"}, Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
								{Given: map[string]string{"fever": "no", "cough": "no"}, Probabilities: map[string]float64{"yes": 0.05, "no": 0.95}},
							},
						},
					},
				},
			},
		},
	}

	svc := evaluate.NewService[PatientSymptoms](PatientBinder{}, repo)

	result, err := svc.Execute(context.Background(), evaluate.Request[PatientSymptoms]{
		Domain: PatientSymptoms{
			HasFever: true,
			HasCough: true,
		},
		RuleSetName:    "diagnosis",
		RuleSetVersion: "v1",
		Paradigm:       evaluate.Bayesian,
		Query:          "flu",
	})

	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Printf("Paradigm: %s\n", result.Paradigm)
	fmt.Printf("P(flu=yes) = %.2f\n", float64(result.Outcome.Distribution["yes"]))

	// Output:
	// Paradigm: bayesian
	// P(flu=yes) = 0.90
}
