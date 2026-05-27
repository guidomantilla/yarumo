package engine

import (
	"maps"
	"strconv"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"

	"github.com/guidomantilla/yarumo/compute/engine/causal/explain"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

type engine struct{}

// NewEngine creates a new causal inference engine.
func NewEngine() Engine {
	return &engine{}
}

func (e *engine) Propagate(scm model.SCM, observations map[string]float64) Result {
	cassert.NotNil(e, "engine is nil")

	trace := explain.NewTrace(observations)
	values := make(map[string]float64)
	stepNum := 0

	maps.Copy(values, observations)

	for _, varName := range scm.Variables() {
		_, observed := observations[varName]
		if observed {
			continue
		}

		v, ok := scm.Variable(varName)
		if !ok || v.Equation == nil {
			continue
		}

		parentVals := gatherParents(v.Parents, values)
		values[varName] = v.Equation(parentVals)

		stepNum++

		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Propagation,
			Message: "compute " + varName + " = " + formatFloat(values[varName]),
			Values:  maps.Clone(values),
		})
	}

	for k, v := range values {
		trace = trace.AddOutput(k, v)
	}

	stepNum++

	trace = trace.AddStep(explain.Step{
		Number:  stepNum,
		Phase:   explain.Complete,
		Message: "propagation complete",
	})

	return Result{Values: values, Trace: trace}
}

func (e *engine) Intervene(scm model.SCM, interventions map[string]float64) Result {
	cassert.NotNil(e, "engine is nil")

	trace := explain.NewTrace(interventions)
	values := make(map[string]float64)
	stepNum := 0

	for k, v := range interventions {
		values[k] = v

		stepNum++

		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Intervention,
			Message: "do(" + k + " = " + formatFloat(v) + ")",
			Values:  maps.Clone(values),
		})
	}

	for _, varName := range scm.Variables() {
		_, intervened := interventions[varName]
		if intervened {
			continue
		}

		v, ok := scm.Variable(varName)
		if !ok || v.Equation == nil {
			continue
		}

		parentVals := gatherParents(v.Parents, values)
		values[varName] = v.Equation(parentVals)

		stepNum++

		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Propagation,
			Message: "compute " + varName + " = " + formatFloat(values[varName]),
			Values:  maps.Clone(values),
		})
	}

	for k, v := range values {
		trace = trace.AddOutput(k, v)
	}

	stepNum++

	trace = trace.AddStep(explain.Step{
		Number:  stepNum,
		Phase:   explain.Complete,
		Message: "intervention complete",
	})

	return Result{Values: values, Trace: trace}
}

func (e *engine) Counterfactual(scm model.SCM, factual map[string]float64, hypothetical map[string]float64) Result {
	cassert.NotNil(e, "engine is nil")

	trace := explain.NewTrace(factual)
	stepNum := 0

	// Step 1: Compute factual world.
	factualResult := e.Propagate(scm, factual)
	factualValues := factualResult.Values

	stepNum++

	trace = trace.AddStep(explain.Step{
		Number:  stepNum,
		Phase:   explain.Counterfactual,
		Message: "factual world computed",
		Values:  maps.Clone(factualValues),
	})

	// Step 2: Start with factual values, apply hypothetical interventions.
	cfValues := maps.Clone(factualValues)

	for k, v := range hypothetical {
		cfValues[k] = v

		stepNum++

		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Counterfactual,
			Message: "hypothetical do(" + k + " = " + formatFloat(v) + ")",
			Values:  maps.Clone(cfValues),
		})
	}

	// Step 3: Re-propagate downstream of intervened variables.
	stepNum, trace = repropagate(scm, factual, hypothetical, cfValues, stepNum, trace)

	for k, v := range cfValues {
		trace = trace.AddOutput(k, v)
	}

	stepNum++

	trace = trace.AddStep(explain.Step{
		Number:  stepNum,
		Phase:   explain.Complete,
		Message: "counterfactual complete",
	})

	return Result{Values: cfValues, Trace: trace}
}

// repropagate recomputes downstream variables affected by hypothetical interventions.
func repropagate(scm model.SCM, factual map[string]float64, hypothetical map[string]float64,
	cfValues map[string]float64, stepNum int, trace explain.Trace,
) (int, explain.Trace) {
	for _, varName := range scm.Variables() {
		if shouldSkipCounterfactual(scm, varName, factual, hypothetical) {
			continue
		}

		v, ok := scm.Variable(varName)
		if !ok || v.Equation == nil {
			continue
		}

		parentVals := gatherParents(v.Parents, cfValues)
		cfValues[varName] = v.Equation(parentVals)

		stepNum++

		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Counterfactual,
			Message: "recompute " + varName + " = " + formatFloat(cfValues[varName]),
			Values:  maps.Clone(cfValues),
		})
	}

	return stepNum, trace
}

// shouldSkipCounterfactual determines whether a variable should be skipped during re-propagation.
func shouldSkipCounterfactual(scm model.SCM, varName string, factual map[string]float64, hypothetical map[string]float64) bool {
	_, intervened := hypothetical[varName]
	if intervened {
		return true
	}

	// Root observations stay if they are truly exogenous (no parents).
	_, observed := factual[varName]
	if observed {
		v, ok := scm.Variable(varName)
		if ok && len(v.Parents) == 0 {
			return true
		}
	}

	return !hasIntervenedAncestor(scm, varName, hypothetical)
}

// hasIntervenedAncestor checks if any ancestor of varName was intervened on.
func hasIntervenedAncestor(scm model.SCM, varName string, interventions map[string]float64) bool {
	for _, parent := range scm.Parents(varName) {
		_, ok := interventions[parent]
		if ok {
			return true
		}

		if hasIntervenedAncestor(scm, parent, interventions) {
			return true
		}
	}

	return false
}

func gatherParents(parents []string, values map[string]float64) map[string]float64 {
	parentVals := make(map[string]float64, len(parents))

	for _, p := range parents {
		parentVals[p] = values[p]
	}

	return parentVals
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}
