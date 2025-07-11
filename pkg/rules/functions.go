package rules

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Rule[T any] struct {
	Label     string
	Formula   propositions.Formula
	Predicate predicates.Predicate[T]
}

type Result[T any] struct {
	Rule      Rule[T]
	Input     T
	Violated  bool
	Satisfied bool
	Traces    []logic.Trace[T]
}

// EvaluateRules evaluates a set of rules against a given input using the provided predicates.
func EvaluateRules[T any](preds logic.Predicates[T], rules []Rule[T], input *T) []Result[T] {
	results := make([]Result[T], 0)
	for _, r := range rules {
		result := logic.EvaluateProposition(r.Formula, preds, input)
		results = append(results, Result[T]{
			Rule:      r,
			Input:     *input,
			Violated:  !result.Result,
			Satisfied: result.Result,
			Traces:    result.Traces,
		})
	}
	return results
}
