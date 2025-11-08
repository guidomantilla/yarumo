package logic

import (
	"fmt"
	"sync"

	"github.com/guidomantilla/yarumo/deprecated/logic/predicates"
	"github.com/guidomantilla/yarumo/deprecated/logic/propositions"
)

type Rule[T any] struct {
	Label       string
	Formula     propositions.Formula
	Consequence *propositions.Var
	tree        *EvalNode
}

type RuleSet[T any] struct {
	mu       sync.Mutex
	registry PredicatesRegistry[T]
	rules    []Rule[T]
}

func NewRuleSet[T any](registry PredicatesRegistry[T], rules []Rule[T]) *RuleSet[T] {
	return &RuleSet[T]{
		registry: registry,
		rules:    rules,
	}
}

// Evaluate evaluates a set of rules against a given input using the provided predicates.
func (e *RuleSet[T]) Evaluate(input *T) (*EvalNode, error) {
	if !IsStruct(input) {
		return nil, fmt.Errorf("input must be a pointer to a struct, got %T", input)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	result := &EvalNode{
		Label: "rules set evaluation",
		Value: ToPtr(true),
		Nodes: []EvalNode{
			{
				Expr:  "consequences",
				Nodes: make([]EvalNode, 0),
			},
		},
	}

	for _, rule := range e.rules {
		tree, err := e.registry.Evaluate(rule.Formula, input)
		if err != nil {
			return nil, fmt.Errorf("error evaluating rule '%s': %w", rule.Label, err)
		}

		tree.Label = rule.Label
		result.Value = ToPtr(*result.Value && *tree.Value)
		if rule.Consequence != nil {
			predicate := predicates.False[T]()
			if *tree.Value {
				predicate = predicates.True[T]()
			}
			e.registry[*rule.Consequence] = predicate

			fact := NewEvalNode(rule.Consequence, *result.Value)
			tree.Nodes = append(tree.Nodes, *fact)
			result.Nodes[0].Nodes = append(result.Nodes[0].Nodes, *fact)
		}
		result.Nodes = append(result.Nodes, *tree)

		rule.tree = tree
	}

	return result, nil
}
