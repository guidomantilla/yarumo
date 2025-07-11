package rules

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
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
func EvaluateRules[T any](value *T, preds logic.Predicates[T], rules []Rule[T]) ([]Result[T], error) {
	if !pointer.IsStruct(value) {
		return nil, fmt.Errorf("value must be a pointer to a struct, got %T", value)
	}
	results := make([]Result[T], 0)
	for _, r := range rules {
		result, _ := logic.EvaluateProposition(value, r.Formula, preds)
		results = append(results, Result[T]{
			Rule:      r,
			Input:     *value,
			Violated:  !result.Result,
			Satisfied: result.Result,
			Traces:    result.Traces,
		})
	}
	return results, nil
}

func Unwrap[T any](rules []Rule[T]) (propositions.Formula, predicates.Predicate[T], error) {
	if len(rules) == 0 {
		return nil, nil, fmt.Errorf("no rules provided")
	}
	formula, predicate := rules[0].Formula, rules[0].Predicate
	for _, rule := range rules[1:] {
		formula = formula.And(rule.Formula)
		predicate = predicate.And(rule.Predicate)
	}
	return formula, predicate, nil
}
