package propositions

import "fmt"

// AndF represents a logical conjunction (∧) of two formulas.
type AndF struct {
	L, R Formula
}

// String returns a string representation of the AndF formula.
func (f AndF) String() string {
	return fmt.Sprintf("(%s ∧ %s)", f.L.String(), f.R.String())
}

// Eval evaluates the AndF formula against a set of facts.
func (f AndF) Eval(facts Fact) bool {
	return f.L.Eval(facts) && f.R.Eval(facts)
}

// Vars returns a slice of variable names used in the AndF formula.
func (f AndF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

// And returns a new AndF formula that combines the current formula with another formula using logical conjunction.
func (f AndF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

// Or returns a new OrF formula that combines the current formula with another formula using logical disjunction.
func (f AndF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

// Not returns a NotF formula that negates the current AndF formula.
func (f AndF) Not() Formula {
	return NotF{F: f}
}

// Implies returns an ImplF formula that represents the implication of the current AndF formula to another formula.
func (f AndF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the current AndF formula with respect to another formula.
func (f AndF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the current AndF formula and another formula.
func (f AndF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

// ToNNF converts the AndF formula to Negation Normal Form (NNF).
func (f AndF) ToNNF() Formula {
	return ToNNF(f)
}

// ToCNF converts the AndF formula to Conjunctive Normal Form (CNF).
func (f AndF) ToCNF() Formula {
	return ToCNF(f)
}

// ToDNF converts the AndF formula to Disjunctive Normal Form (DNF).
func (f AndF) ToDNF() Formula {
	return ToDNF(f)
}
