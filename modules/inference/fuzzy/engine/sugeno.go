package engine

import (
	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/explain"
)

// sugeno executes Sugeno-style fuzzy inference.
//
//  1. Fuzzify each input variable.
//  2. Evaluate each rule to get firing strengths.
//  3. Compute weighted average: output = sum(strength * singleton) / sum(strength).
func (e *engine) sugeno(inputs map[string]float64, trace explain.Trace) Result {
	stepNum := 0

	// 1. Fuzzify inputs.
	fuzzified, trace := e.fuzzifyInputs(inputs, trace, &stepNum)

	// 2. Evaluate rules.
	ruleResults, trace := e.evaluateRules(fuzzified, trace, &stepNum)

	// 3. Weighted average per output variable.
	outputs := make(map[string]float64, len(e.outputVars))

	for _, ov := range e.outputVars {
		var weightedSum float64

		var strengthSum fuzzym.Degree

		for _, rr := range ruleResults {
			if rr.rule.Consequent().Variable != ov.Name() {
				continue
			}

			if rr.strength <= 0 {
				continue
			}

			key := rr.rule.Consequent().Variable + "/" + rr.rule.Consequent().Term
			singleton, ok := e.sugenoOutputs[key]

			if !ok {
				continue
			}

			weightedSum += float64(rr.strength) * singleton
			strengthSum += rr.strength
		}

		crispValue := 0.0

		if strengthSum > 0 {
			crispValue = weightedSum / float64(strengthSum)
		}

		outputs[ov.Name()] = crispValue

		stepNum++
		trace = trace.AddStep(explain.Step{
			Number:  stepNum,
			Phase:   explain.Aggregation,
			Message: "sugeno aggregate " + ov.Name(),
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
