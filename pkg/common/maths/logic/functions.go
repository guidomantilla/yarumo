package logic

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

type Result[T any] struct {
	Formula    propositions.Formula
	Predicates Predicates[T]
	Value      T
	Facts      []Fact[T]
	Result     bool
}

type Fact[T any] struct {
	Variable propositions.Var
	Value    bool
	Func     predicates.Predicate[T]
}

type Predicates[T any] map[propositions.Var]predicates.Predicate[T]

// EvaluateProposition translates a proposition formula into a predicate function using the provided predicates.
func EvaluateProposition[T any](value *T, formula propositions.Formula, preds Predicates[T]) (*Result[T], error) {
	if !pointer.IsStruct(value) {
		return nil, fmt.Errorf("value must be a pointer to a struct, got %T", value)
	}
	checkedPredicates, facts := checkPredicates(preds)
	eval := compileProposition[T](formula, checkedPredicates)
	result := eval(*value)
	return &Result[T]{
		Formula:    formula,
		Predicates: checkedPredicates,
		Value:      *value,
		Facts:      *facts,
		Result:     result,
	}, nil
}

func checkPredicates[T any](predicates Predicates[T]) (Predicates[T], *[]Fact[T]) {
	var facts = make([]Fact[T], 0)
	wrapped := make(Predicates[T])
	for variable, pred := range predicates {
		wrapped[variable] = func(t T) bool {
			val := pred(t)
			facts = append(facts, Fact[T]{Variable: variable, Value: val, Func: pred})
			return val
		}
	}

	return wrapped, &facts
}

func compileProposition[T any](formula propositions.Formula, preds Predicates[T]) predicates.Predicate[T] {
	switch x := formula.(type) {
	case propositions.AndF:
		return predicates.And(compileProposition[T](x.L, preds), compileProposition[T](x.R, preds))
	case propositions.FalseF:
		return predicates.False[T]()
	case propositions.GroupF:
		return compileProposition[T](x.Inner, preds)
	case propositions.IffF:
		return predicates.Iff(compileProposition[T](x.L, preds), compileProposition[T](x.R, preds))
	case propositions.ImplF:
		return predicates.Implies(compileProposition[T](x.L, preds), compileProposition[T](x.R, preds))
	case propositions.NotF:
		return predicates.Not(compileProposition[T](x.F, preds))
	case propositions.OrF:
		return predicates.Or(compileProposition[T](x.L, preds), compileProposition[T](x.R, preds))
	case propositions.TrueF:
		return predicates.True[T]()
	case propositions.Var:
		p, ok := preds[x]
		if ok {
			return p
		}
		panic(fmt.Sprintf("propositions.Var '%s' not found", string(x)))
	default:
		panic(fmt.Sprintf("unsupported proposition type: %T", x))
	}
}
