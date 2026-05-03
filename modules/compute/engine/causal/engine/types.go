// Package engine provides causal inference algorithms based on structural causal models.
package engine

import (
	"github.com/guidomantilla/yarumo/compute/engine/causal/explain"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

// Result of a causal inference query.
type Result struct {
	Values map[string]float64
	Trace  explain.Trace
}

// Engine performs causal inference queries on structural causal models.
type Engine interface {
	// Propagate computes all variable values given root observations (Level 1: Association).
	Propagate(scm model.SCM, observations map[string]float64) Result
	// Intervene applies the do-operator and propagates (Level 2: Intervention).
	Intervene(scm model.SCM, interventions map[string]float64) Result
	// Counterfactual computes what would happen under hypothetical intervention given factual observations.
	Counterfactual(scm model.SCM, factual map[string]float64, hypothetical map[string]float64) Result
}

// Type compliance.
var _ Engine = (*engine)(nil)
