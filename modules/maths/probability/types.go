// Package probability provides discrete probability primitives.
package probability

// Var is a random variable name.
type Var string

// Outcome is a possible value of a random variable.
type Outcome string

// Prob is a probability value in [0,1].
type Prob float64

// Distribution maps outcomes to their probabilities.
type Distribution map[Outcome]Prob

// Assignment maps variables to observed outcomes.
type Assignment map[Var]Outcome

// CPT is a conditional probability table.
// Maps parent configurations to child outcome distributions.
type CPT struct {
	Variable Var
	Parents  []Var
	Entries  map[string]Distribution // serialized parent config -> Distribution.
}

// Factor is an intermediate table used in variable elimination.
type Factor struct {
	Variables []Var
	Table     map[string]Prob // serialized assignment -> probability.
}
