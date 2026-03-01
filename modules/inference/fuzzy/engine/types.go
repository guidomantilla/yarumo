// Package engine provides fuzzy inference algorithms (Mamdani and Sugeno).
package engine

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/explain"
	"github.com/guidomantilla/yarumo/inference/fuzzy/rules"
	"github.com/guidomantilla/yarumo/inference/fuzzy/variable"
)

// Method identifies the fuzzy inference method.
type Method int

const (
	// Mamdani uses fuzzy sets as outputs and defuzzifies via centroid or similar.
	Mamdani Method = iota
	// Sugeno uses singleton outputs and computes a weighted average.
	Sugeno
)

// Result holds the outcome of a fuzzy inference run.
type Result struct {
	Outputs map[string]float64
	Trace   explain.Trace
}

// Engine defines the interface for a fuzzy inference engine.
type Engine interface {
	// Infer executes fuzzy inference given crisp input values.
	Infer(inputs map[string]float64) Result
}

var _ Engine = (*engine)(nil)

// engine is the private implementation of Engine.
type engine struct {
	inputVars  []variable.Variable
	outputVars []variable.Variable
	ruleSet    []rules.Rule
	options    Options
	// sugenoOutputs maps "variable/term" to singleton crisp values for Sugeno method.
	sugenoOutputs map[string]float64
}

// NewEngine creates a new fuzzy inference engine.
func NewEngine(inputVars, outputVars []variable.Variable, ruleSet []rules.Rule, opts ...Option) Engine {
	cassert.NotEmpty(inputVars, "input variables are empty")
	cassert.NotEmpty(outputVars, "output variables are empty")
	cassert.NotEmpty(ruleSet, "rule set is empty")

	options := NewOptions(opts...)

	return &engine{
		inputVars:     inputVars,
		outputVars:    outputVars,
		ruleSet:       ruleSet,
		options:       options,
		sugenoOutputs: options.sugenoOutputs,
	}
}

// Infer executes fuzzy inference given crisp input values.
func (e *engine) Infer(inputs map[string]float64) Result {
	cassert.NotNil(e, "engine is nil")

	trace := explain.NewTrace(inputs)

	if e.options.method == Sugeno {
		return e.sugeno(inputs, trace)
	}

	return e.mamdani(inputs, trace)
}

// fuzzifyInputs evaluates all input variables against the given inputs and records fuzzification steps.
func (e *engine) fuzzifyInputs(inputs map[string]float64, trace explain.Trace, stepNum *int) (map[string]map[string]fuzzym.Degree, explain.Trace) {
	cassert.NotNil(e, "engine is nil")

	fuzzified := make(map[string]map[string]fuzzym.Degree, len(e.inputVars))

	for _, iv := range e.inputVars {
		val, ok := inputs[iv.Name()]
		if !ok {
			continue
		}

		degrees := iv.Fuzzify(val)
		fuzzified[iv.Name()] = degrees

		memberships := make([]explain.Membership, 0, len(degrees))

		for term, deg := range degrees {
			memberships = append(memberships, explain.Membership{
				Variable: iv.Name(),
				Term:     term,
				Degree:   deg,
			})
		}

		*stepNum++
		trace = trace.AddStep(explain.Step{
			Number:      *stepNum,
			Phase:       explain.Fuzzification,
			Message:     "fuzzify " + iv.Name(),
			Memberships: memberships,
		})
	}

	return fuzzified, trace
}

// evaluateRules computes rule strengths from fuzzified inputs and records rule evaluation steps.
func (e *engine) evaluateRules(fuzzified map[string]map[string]fuzzym.Degree, trace explain.Trace, stepNum *int) ([]ruleResult, explain.Trace) {
	cassert.NotNil(e, "engine is nil")

	results := make([]ruleResult, 0, len(e.ruleSet))

	for _, r := range e.ruleSet {
		strength := e.evaluateRule(r, fuzzified)
		strength = fuzzym.Degree(float64(strength) * r.Weight())

		results = append(results, ruleResult{
			rule:     r,
			strength: strength,
		})

		*stepNum++
		trace = trace.AddStep(explain.Step{
			Number:  *stepNum,
			Phase:   explain.RuleEvaluation,
			Message: "evaluate " + r.Name(),
			Activations: []explain.Activation{{
				RuleName: r.Name(),
				Strength: strength,
				Output:   r.Consequent().Variable,
				Term:     r.Consequent().Term,
			}},
		})
	}

	return results, trace
}

// evaluateRule computes the firing strength of a single rule.
func (e *engine) evaluateRule(r rules.Rule, fuzzified map[string]map[string]fuzzym.Degree) fuzzym.Degree {
	cassert.NotNil(e, "engine is nil")

	conditions := r.Conditions()
	if len(conditions) == 0 {
		return 0
	}

	degrees := make([]fuzzym.Degree, 0, len(conditions))

	for _, c := range conditions {
		varDegrees, ok := fuzzified[c.Variable]
		if !ok {
			degrees = append(degrees, 0)

			continue
		}

		deg, ok := varDegrees[c.Term]
		if !ok {
			degrees = append(degrees, 0)

			continue
		}

		degrees = append(degrees, deg)
	}

	result := degrees[0]

	for i := 1; i < len(degrees); i++ {
		if r.Operator() == rules.Or {
			result = e.options.tconorm(result, degrees[i])
		} else {
			result = e.options.tnorm(result, degrees[i])
		}
	}

	return result
}

// ruleResult pairs a rule with its computed firing strength.
type ruleResult struct {
	rule     rules.Rule
	strength fuzzym.Degree
}
