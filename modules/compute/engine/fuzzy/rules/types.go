// Package rules provides fuzzy inference rule definitions.
package rules

// Operator defines how multiple conditions in a rule are combined.
type Operator int

const (
	// And combines conditions using t-norm (fuzzy AND).
	And Operator = iota
	// Or combines conditions using t-conorm (fuzzy OR).
	Or
)

// Condition represents a single fuzzy condition (e.g., "temperature IS hot").
type Condition struct {
	Variable string
	Term     string
}

// Consequent represents the rule output (e.g., "speed IS high").
type Consequent struct {
	Variable string
	Term     string
}

// Rule represents a fuzzy inference rule.
type Rule interface {
	// Name returns the rule identifier.
	Name() string
	// Conditions returns the antecedent conditions.
	Conditions() []Condition
	// Operator returns how conditions are combined.
	Operator() Operator
	// Consequent returns the rule output.
	Consequent() Consequent
	// Weight returns the rule weight in [0,1].
	Weight() float64
}

var _ Rule = (*rule)(nil)
