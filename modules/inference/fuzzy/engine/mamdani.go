package engine

import (
	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/explain"
)

// mamdani executes Mamdani-style fuzzy inference.
//
//  1. Fuzzify each input variable.
//  2. Evaluate each rule to get firing strengths.
//  3. Clip output membership functions by rule strengths.
//  4. Aggregate clipped outputs per output variable (max).
//  5. Defuzzify aggregated output to get crisp values.
func (e *engine) mamdani(inputs map[string]float64, trace explain.Trace) Result {
	stepNum := 0

	// 1. Fuzzify inputs.
	fuzzified, trace := e.fuzzifyInputs(inputs, trace, &stepNum)

	// 2. Evaluate rules.
	ruleResults, trace := e.evaluateRules(fuzzified, trace, &stepNum)

	// 3+4. Aggregate clipped outputs per output variable.
	outputs := make(map[string]float64, len(e.outputVars))

	for _, ov := range e.outputVars {
		clipped := make([]fuzzym.MembershipFn, 0)

		for _, rr := range ruleResults {
			if rr.rule.Consequent().Variable != ov.Name() {
				continue
			}

			if rr.strength <= 0 {
				continue
			}

			term, ok := ov.Term(rr.rule.Consequent().Term)
			if !ok {
				continue
			}

			clipped = append(clipped, fuzzym.Clip(term.Fn, rr.strength))
		}

		stepNum++
		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Aggregation,
			Message: "aggregate " + ov.Name(),
		})

		// 5. Defuzzify.
		crispValue := 0.0

		if len(clipped) > 0 {
			aggregated := fuzzym.AggregateMax(clipped...)

			xs, ys, err := fuzzym.Sample(aggregated, ov.Min(), ov.Max(), e.options.resolution)
			if err == nil {
				crispValue = e.options.defuzzify(xs, ys)
			}
		}

		outputs[ov.Name()] = crispValue

		stepNum++
		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Defuzzification,
			Message: "defuzzify " + ov.Name(),
		})

		trace = trace.AddOutput(explain.Output{
			Variable:   ov.Name(),
			CrispValue: crispValue,
		})
	}

	stepNum++
	trace = trace.AddStep(explain.Step{
		Number:  stepNum,
		Phase:   explain.Complete,
		Message: "inference complete",
	})

	return Result{
		Outputs: outputs,
		Trace:   trace,
	}
}
