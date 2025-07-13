package rules

import (
	"fmt"
	"sync"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

type Evaluator[T any] struct {
	mu       sync.Mutex
	registry logic.PredicatesRegistry[T]
	rules    []Rule[T]
}

func NewEvaluator[T any](registry logic.PredicatesRegistry[T], rules []Rule[T]) *Evaluator[T] {
	return &Evaluator[T]{
		registry: registry,
		rules:    rules,
	}
}

// Evaluate evaluates a set of rules against a given input using the provided predicates.
func (e *Evaluator[T]) Evaluate(input *T) (*logic.EvalNode, error) {
	if !pointer.IsStruct(input) {
		return nil, fmt.Errorf("input must be a pointer to a struct, got %T", input)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	result := &logic.EvalNode{
		Expr:  "rules set evaluation",
		Value: pointer.ToPtr(true),
		Nodes: []logic.EvalNode{
			{
				Expr:  "consequences",
				Nodes: make([]logic.EvalNode, 0),
			},
		},
	}

	for _, rule := range e.rules {
		tree, err := e.registry.Evaluate(rule.Formula, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating rule '%s': %w", rule.Label, err)
		}
		result.Value = pointer.ToPtr(*result.Value && *tree.Value)
		result.Nodes = append(result.Nodes, *tree)

		if rule.Consequence == nil {
			continue
		}

		derived := result.Facts()
		if derived[*rule.Consequence] {
			continue
		}

		predicate := predicates.False[T]()
		if *tree.Value {
			predicate = predicates.True[T]()
		}
		e.registry[*rule.Consequence] = predicate

		fact := logic.NewEvalNode(rule.Consequence, *result.Value)
		result.Nodes[0].Nodes = append(result.Nodes[0].Nodes, *fact)
	}

	return result, nil
}
