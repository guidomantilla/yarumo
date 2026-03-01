package rules

import (
	"maps"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/logic"
)

type rule struct {
	name       string
	priority   int
	condition  logic.Formula
	conclusion map[logic.Var]bool
}

// NewRule creates a rule with the given name, condition, and conclusion.
func NewRule(name string, condition logic.Formula, conclusion map[logic.Var]bool, opts ...Option) Rule {
	cassert.NotEmpty(name, "rule name is empty")
	cassert.NotNil(condition, "rule condition is nil")
	cassert.NotEmpty(conclusion, "rule conclusion is empty")

	options := NewOptions(opts...)

	return &rule{
		name:       name,
		priority:   options.priority,
		condition:  condition,
		conclusion: copyConclusion(conclusion),
	}
}

// Name returns the rule identifier.
func (r *rule) Name() string {
	cassert.NotNil(r, "rule is nil")

	return r.name
}

// Priority returns the rule priority.
func (r *rule) Priority() int {
	cassert.NotNil(r, "rule is nil")

	return r.priority
}

// Condition returns the propositional formula that must be satisfied.
func (r *rule) Condition() logic.Formula {
	cassert.NotNil(r, "rule is nil")

	return r.condition
}

// Conclusion returns a copy of the variable assignments produced when the rule fires.
func (r *rule) Conclusion() map[logic.Var]bool {
	cassert.NotNil(r, "rule is nil")

	return copyConclusion(r.conclusion)
}

// Fires reports whether the rule condition is satisfied by the given facts.
func (r *rule) Fires(facts logic.Fact) bool {
	cassert.NotNil(r, "rule is nil")

	return r.condition.Eval(facts)
}

// Produces reports whether the rule would derive new information from the given facts.
func (r *rule) Produces(facts logic.Fact) bool {
	cassert.NotNil(r, "rule is nil")

	if !r.Fires(facts) {
		return false
	}

	for v, val := range r.conclusion {
		current, known := facts[v]
		if !known || current != val {
			return true
		}
	}

	return false
}

func copyConclusion(src map[logic.Var]bool) map[logic.Var]bool {
	dst := make(map[logic.Var]bool, len(src))
	maps.Copy(dst, src)

	return dst
}
