package logic

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/deprecated/logic/predicates"
	"github.com/guidomantilla/yarumo/internal/deprecated/logic/propositions"
)

// PredicatesRegistry is a type alias for a registry of predicates that maps proposition variables to their corresponding predicate functions.
type PredicatesRegistry[T any] map[propositions.Var]predicates.Predicate[T]

func (registry PredicatesRegistry[T]) sanitize() { // Ensure all predicates are valid functions
	for key, predicate := range registry {
		if predicate == nil {
			delete(registry, key)
		}
	}
}

// Compile translates a proposition formula into a predicate function using the provided predicates' registry.
func (registry PredicatesRegistry[T]) Compile(formula propositions.Formula) (predicates.Predicate[T], error) {
	registry.sanitize()
	switch x := formula.(type) {
	case propositions.Var:
		predicate, ok := registry[x]
		if !ok {
			return nil, fmt.Errorf("missing predicate for variable '%s'", x)
		}
		return predicate, nil

	case propositions.TrueF:
		predicate := predicates.True[T]()
		return predicate, nil

	case propositions.FalseF:
		predicate := predicates.False[T]()
		return predicate, nil

	case propositions.NotF:
		predicate, err := registry.Compile(x.F)
		if err != nil {
			return nil, fmt.Errorf("error compiling NOT formula: %w", err)
		}
		return predicates.Not(predicate), nil

	case propositions.AndF:
		left, err := registry.Compile(x.L)
		if err != nil {
			return nil, fmt.Errorf("error compiling AND left formula: %w", err)
		}
		right, err := registry.Compile(x.R)
		if err != nil {
			return nil, fmt.Errorf("error compiling AND right formula: %w", err)
		}
		return predicates.And(left, right), nil

	case propositions.OrF:
		left, err := registry.Compile(x.L)
		if err != nil {
			return nil, fmt.Errorf("error compiling OR left formula: %w", err)
		}
		right, err := registry.Compile(x.R)
		if err != nil {
			return nil, fmt.Errorf("error compiling OR right formula: %w", err)
		}
		return predicates.Or(left, right), nil

	case propositions.ImplF:
		left, err := registry.Compile(x.L)
		if err != nil {
			return nil, fmt.Errorf("error compiling IMPL left formula: %w", err)
		}
		right, err := registry.Compile(x.R)
		if err != nil {
			return nil, fmt.Errorf("error compiling IMPL right formula: %w", err)
		}
		return predicates.Implies(left, right), nil

	case propositions.IffF:
		left, err := registry.Compile(x.L)
		if err != nil {
			return nil, fmt.Errorf("error compiling IFF left formula: %w", err)
		}
		right, err := registry.Compile(x.R)
		if err != nil {
			return nil, fmt.Errorf("error compiling IFF right formula: %w", err)
		}
		return predicates.Iff(left, right), nil

	case propositions.GroupF:
		predicate, err := registry.Compile(x.Inner)
		if err != nil {
			return nil, fmt.Errorf("error compiling GROUP formula: %w", err)
		}
		return predicate, nil

	default:
		return nil, fmt.Errorf("unknown formula type: %T", x)
	}
}

// Evaluate evaluates a proposition formula against an input of type T using the predicate registry.
func (registry PredicatesRegistry[T]) Evaluate(f propositions.Formula, input *T) (*EvalNode, error) {
	if !IsStruct(input) {
		return nil, fmt.Errorf("input must be a pointer to a struct, got %T", input)
	}
	registry.sanitize()

	switch x := f.(type) {
	case propositions.Var:
		predicate, ok := registry[x]
		if !ok {
			return nil, fmt.Errorf("missing predicate for variable '%s'", x)
		}
		return NewEvalNode(x, predicate(*input)), nil

	case propositions.TrueF:
		return NewEvalNode(x, true), nil

	case propositions.FalseF:
		return NewEvalNode(x, false), nil

	case propositions.NotF:
		child, err := registry.Evaluate(x.F, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating NOT formula: %w", err)
		}
		return NewEvalNode(x, !*child.Value, *child), nil

	case propositions.AndF:
		left, err := registry.Evaluate(x.L, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating AND left formula: %w", err)
		}
		right, err := registry.Evaluate(x.R, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating AND right formula: %w", err)
		}
		return NewEvalNode(x, *left.Value && *right.Value, *left, *right), nil

	case propositions.OrF:
		left, err := registry.Evaluate(x.L, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating OR left formula: %w", err)
		}
		right, err := registry.Evaluate(x.R, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating OR right formula: %w", err)
		}
		return NewEvalNode(x, *left.Value || *right.Value, *left, *right), nil

	case propositions.ImplF:
		left, err := registry.Evaluate(x.L, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating IMPL left formula: %w", err)
		}
		right, err := registry.Evaluate(x.R, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating IMPL right formula: %w", err)
		}
		return NewEvalNode(x, !*left.Value || *right.Value, *left, *right), nil

	case propositions.IffF:
		left, err := registry.Evaluate(x.L, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating IFF left formula: %w", err)
		}
		right, err := registry.Evaluate(x.R, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating IFF right formula: %w", err)
		}
		return NewEvalNode(x, left.Value == right.Value, *left, *right), nil

	case propositions.GroupF:
		child, err := registry.Evaluate(x.Inner, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating GROUP formula: %w", err)
		}
		return NewEvalNode(x, *child.Value, *child), nil

	default:
		return nil, fmt.Errorf("unknown formula type: %T", x)
	}
}
