package logic

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Result[T any] struct {
	Formula    propositions.Formula
	Predicates Predicates[T]
	Value      T
	Traces     []Trace[T]
	Result     bool
}

type Trace[T any] struct {
	Name  string
	Func  predicates.Predicate[T]
	Value bool
}

type Predicates[T any] map[propositions.Var]predicates.Predicate[T]

// EvaluateProposition translates a proposition formula into a predicate function using the provided predicates.
func EvaluateProposition[T any](formula propositions.Formula, preds Predicates[T], value *T) *Result[T] {
	tracedPredicates, traces := tracePredicates(preds)
	eval := compileProposition[T](formula, tracedPredicates)
	result := eval(*value)
	return &Result[T]{
		Formula:    formula,
		Predicates: tracedPredicates,
		Value:      *value,
		Traces:     *traces,
		Result:     result,
	}
}

func tracePredicates[T any](predicates Predicates[T]) (Predicates[T], *[]Trace[T]) {
	var traces = make([]Trace[T], 0)
	wrapped := make(Predicates[T])
	for variable, pred := range predicates {
		name := string(variable) // para cierre
		wrapped[variable] = func(t T) bool {
			val := pred(t)
			traces = append(traces, Trace[T]{Name: name, Func: pred, Value: val})
			return val
		}
	}

	return wrapped, &traces
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
