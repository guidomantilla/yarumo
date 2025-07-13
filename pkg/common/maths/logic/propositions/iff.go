package propositions

import "fmt"

// IffF represents a logical biconditional (if and only if, ⇔) between two formulas.
type IffF struct {
	L, R Formula
}

// Iff creates a new IffF formula that represents the biconditional relationship between two formulas.
func (f IffF) String() string {
	return fmt.Sprintf("(%s ⇔ %s)", f.L.String(), f.R.String())
}

// Eval evaluates the IffF formula against a set of facts, returning true if both formulas evaluate to the same boolean value.
func (f IffF) Eval(facts Fact) bool {
	return f.L.Eval(facts) == f.R.Eval(facts)
}

// Vars returns a slice of variable names used in the IffF formula, combining the variables from both left and right formulas.
func (f IffF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

// And returns a new AndF formula that combines the IffF formula with another formula using logical conjunction.
func (f IffF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

// Or returns a new OrF formula that combines the IffF formula with another formula using logical disjunction.
func (f IffF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

// Not returns a NotF formula that negates the IffF formula, effectively flipping the truth value of the biconditional.
func (f IffF) Not() Formula {
	return NotF{F: f}
}

// Implies returns an ImplF formula that represents the implication of the IffF formula to another formula.
func (f IffF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the IffF formula with respect to another formula.
func (f IffF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the IffF formula and another formula.
func (f IffF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

// ToNNF converts the IffF formula to Negation Normal Form (NNF).
func (f IffF) ToNNF() Formula {
	return ToNNF(f)
}

// ToCNF converts the IffF formula to Conjunctive Normal Form (CNF).
func (f IffF) ToCNF() Formula {
	return ToCNF(f)
}

// ToDNF converts the IffF formula to Disjunctive Normal Form (DNF).
func (f IffF) ToDNF() Formula {
	return ToDNF(f)
}
