package rules

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

// EvaluateRules evaluates a set of rules against a given input using the provided predicates.
func EvaluateRules[T any](value *T, registry logic.PredicatesRegistry[T], rules []Rule[T]) ([]Result[T], error) {
	if !pointer.IsStruct(value) {
		return nil, fmt.Errorf("value must be a pointer to a struct, got %T", value)
	}
	results := make([]Result[T], 0)
	for _, rule := range rules {
		_, err := registry.Evaluate(rule.Formula, *value)
		if err != nil {
			return nil, fmt.Errorf("error evaluating rule '%s': %w", rule.Label, err)
		}
		/*
				registry[rule.Consequence] = predicates.False[T]()
				if result.Result {
					registry[rule.Consequence] = predicates.True[T]()
				}

				consequence := &logic.Fact[T]{
					Variable: rule.Consequence,
					Value:    result.Result,
					Func:     registry[rule.Consequence],
				}

			result.Facts = append(result.Facts, *consequence)
		*/
		evalTree := BuildEvalTree[T](rule.Formula, *value, registry)
		results = append(results, Result[T]{
			Rule:  rule,
			Input: *value,
			//Violated:  !result.Result,
			//Satisfied: result.Result,
			//Facts:     result.Facts,
			//Consequence: consequence,
			EvalTree: *evalTree,
		})
	}
	return results, nil
}

func BuildEvalTree[T any](f propositions.Formula, input T, registry logic.PredicatesRegistry[T]) *EvalNode {
	switch x := f.(type) {
	case propositions.Var:
		val := registry[x](input)
		return &EvalNode{Expr: x.String(), Value: val}

	case propositions.TrueF:
		return &EvalNode{Expr: x.String(), Value: true}

	case propositions.FalseF:
		return &EvalNode{Expr: x.String(), Value: false}

	case propositions.NotF:
		child := BuildEvalTree[T](x.F, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    !child.Value,
			Children: []EvalNode{*child},
		}

	case propositions.AndF:
		left := BuildEvalTree[T](x.L, input, registry)
		right := BuildEvalTree[T](x.R, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    left.Value && right.Value,
			Children: []EvalNode{*left, *right},
		}

	case propositions.OrF:
		left := BuildEvalTree[T](x.L, input, registry)
		right := BuildEvalTree[T](x.R, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    left.Value || right.Value,
			Children: []EvalNode{*left, *right},
		}

	case propositions.ImplF:
		left := BuildEvalTree[T](x.L, input, registry)
		right := BuildEvalTree[T](x.R, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    !left.Value || right.Value,
			Children: []EvalNode{*left, *right},
		}

	case propositions.IffF:
		left := BuildEvalTree[T](x.L, input, registry)
		right := BuildEvalTree[T](x.R, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    left.Value == right.Value,
			Children: []EvalNode{*left, *right},
		}

	case propositions.GroupF:
		child := BuildEvalTree[T](x.Inner, input, registry)
		return &EvalNode{
			Expr:     x.String(),
			Value:    child.Value,
			Children: []EvalNode{*child},
		}

	default:
		return &EvalNode{Expr: "UNKNOWN", Value: false}
	}
}
