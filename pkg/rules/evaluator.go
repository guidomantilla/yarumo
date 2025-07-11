package rules

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

// EvaluateRules evaluates a set of rules against a given input using the provided predicates.
func EvaluateRules[T any](value *T, preds logic.Predicates[T], rules []Rule[T]) ([]Result[T], error) {
	if !pointer.IsStruct(value) {
		return nil, fmt.Errorf("value must be a pointer to a struct, got %T", value)
	}
	results := make([]Result[T], 0)
	for _, rule := range rules {
		result, _ := logic.EvaluateProposition(value, rule.Formula, preds)
		results = append(results, Result[T]{
			Rule:      rule,
			Input:     *value,
			Violated:  !result.Result,
			Satisfied: result.Result,
			Traces:    result.Traces,
		})
	}
	return results, nil
}
