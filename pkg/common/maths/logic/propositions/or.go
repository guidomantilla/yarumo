package propositions

import "fmt"

// OrF represents a logical disjunction (∨) of two formulas.
type OrF struct {
	L, R Formula
}

// Or creates a new OrF formula that represents the disjunction of two formulas L and R.
func (f OrF) String() string {
	return fmt.Sprintf("(%s ∨ %s)", f.L.String(), f.R.String())
}

// Eval evaluates the OrF formula against a set of facts, returning true if either L or R evaluates to true.
func (f OrF) Eval(facts Fact) bool {
	return f.L.Eval(facts) || f.R.Eval(facts)
}

// Vars returns a slice of variable names used in the OrF formula, combining the variables from both left and right formulas.
func (f OrF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

// And returns a new AndF formula that combines the OrF formula with another formula using logical conjunction.
func (f OrF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

// Or returns a new OrF formula that combines the current OrF formula with another formula using logical disjunction.
func (f OrF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

// Not returns a NotF formula that negates the current OrF formula, effectively flipping the truth value of the disjunction.
func (f OrF) Not() Formula {
	return NotF{F: f}
}

// Implies returns an ImplF formula that represents the implication of the current OrF formula to another formula.
func (f OrF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the current OrF formula with respect to another formula.
func (f OrF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the current OrF formula and another formula.
func (f OrF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

// ToNNF converts the OrF formula to Negation Normal Form (NNF).
func (f OrF) ToNNF() Formula {
	return ToNNF(f)
}

// ToCNF converts the OrF formula to Conjunctive Normal Form (CNF).
func (f OrF) ToCNF() Formula {
	return ToCNF(f)
}

// ToDNF converts the OrF formula to Disjunctive Normal Form (DNF).
func (f OrF) ToDNF() Formula {
	return ToDNF(f)
}
