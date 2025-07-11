package logic

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Traces[T any] *[]Trace[T]

type Trace[T any] struct {
	Name  string
	Func  predicates.Predicate[T]
	Value bool
}

type Predicates[T any] map[propositions.Var]predicates.Predicate[T]

func NewPredicates[T any](predicates Predicates[T]) (Predicates[T], Traces[T]) {
	traces := make([]Trace[T], 0)
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

// CompileProposition translates a proposition formula into a predicate function using the provided predicates.
func CompileProposition[T any](formula propositions.Formula, preds Predicates[T]) predicates.Predicate[T] {
	switch x := formula.(type) {
	case propositions.AndF:
		return predicates.And(CompileProposition[T](x.L, preds), CompileProposition[T](x.R, preds))
	case propositions.FalseF:
		return predicates.False[T]()
	case propositions.GroupF:
		return CompileProposition[T](x.Inner, preds)
	case propositions.IffF:
		return predicates.Iff(CompileProposition[T](x.L, preds), CompileProposition[T](x.R, preds))
	case propositions.ImplF:
		return predicates.Implies(CompileProposition[T](x.L, preds), CompileProposition[T](x.R, preds))
	case propositions.NotF:
		return predicates.Not(CompileProposition[T](x.F, preds))
	case propositions.OrF:
		return predicates.Or(CompileProposition[T](x.L, preds), CompileProposition[T](x.R, preds))
	case propositions.TrueF:
		return predicates.True[T]()
	case propositions.Var:
		name := string(x)
		p, ok := preds[x]
		if ok {
			return p
		}
		panic(fmt.Sprintf("propositions.Var '%s' not found", name))
	default:
		panic(fmt.Sprintf("unsupported proposition type: %T", x))
	}
}
