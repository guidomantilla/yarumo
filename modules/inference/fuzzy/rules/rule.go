package rules

import cassert "github.com/guidomantilla/yarumo/common/assert"

type rule struct {
	name       string
	conditions []Condition
	operator   Operator
	consequent Consequent
	weight     float64
}

// NewRule creates a fuzzy rule with the given name, conditions, and consequent.
func NewRule(name string, conditions []Condition, consequent Consequent, opts ...Option) Rule {
	cassert.NotEmpty(name, "rule name is empty")
	cassert.NotEmpty(conditions, "rule conditions are empty")
	cassert.NotEmpty(consequent.Variable, "rule consequent variable is empty")
	cassert.NotEmpty(consequent.Term, "rule consequent term is empty")

	options := NewOptions(opts...)

	copied := make([]Condition, len(conditions))
	copy(copied, conditions)

	return &rule{
		name:       name,
		conditions: copied,
		operator:   options.operator,
		consequent: consequent,
		weight:     options.weight,
	}
}

// Name returns the rule identifier.
func (r *rule) Name() string {
	cassert.NotNil(r, "rule is nil")

	return r.name
}

// Conditions returns the antecedent conditions.
func (r *rule) Conditions() []Condition {
	cassert.NotNil(r, "rule is nil")

	copied := make([]Condition, len(r.conditions))
	copy(copied, r.conditions)

	return copied
}

// Operator returns how conditions are combined.
func (r *rule) Operator() Operator {
	cassert.NotNil(r, "rule is nil")

	return r.operator
}

// Consequent returns the rule output.
func (r *rule) Consequent() Consequent {
	cassert.NotNil(r, "rule is nil")

	return r.consequent
}

// Weight returns the rule weight in [0,1].
func (r *rule) Weight() float64 {
	cassert.NotNil(r, "rule is nil")

	return r.weight
}
