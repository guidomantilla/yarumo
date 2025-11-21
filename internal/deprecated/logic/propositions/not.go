package propositions

import "fmt"

// NotF represents a logical negation (!) of a formula.
type NotF struct {
	F Formula
}

// Not creates a new NotF formula that negates the provided formula f.
func (f NotF) String() string {
	return fmt.Sprintf("!%s", f.F.String())
}

// Eval evaluates the NotF formula against a set of facts, returning the negation of the evaluation of the inner formula.
func (f NotF) Eval(facts Fact) bool {
	return !f.F.Eval(facts)
}

// Vars returns a slice of variable names used in the NotF formula, which are the same as those in the inner formula.
func (f NotF) Vars() []string {
	return f.F.Vars()
}

// And returns a new AndF formula that combines the NotF formula with another formula using logical conjunction.
func (f NotF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

// Or returns a new OrF formula that combines the NotF formula with another formula using logical disjunction.
func (f NotF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

// Not returns a NotF formula that negates the current NotF formula, effectively flipping the truth value of the negation.
func (f NotF) Not() Formula {
	return NotF{F: f}
}

// Implies returns an ImplF formula that represents the implication of the NotF formula to another formula.
func (f NotF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the NotF formula with respect to another formula.
func (f NotF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the NotF formula and another formula.
func (f NotF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

// ToNNF converts the NotF formula to Negation Normal Form (NNF).
func (f NotF) ToNNF() Formula {
	return ToNNF(f)
}

// ToCNF converts the NotF formula to Conjunctive Normal Form (CNF).
func (f NotF) ToCNF() Formula {
	return ToCNF(f)
}

// ToDNF converts the NotF formula to Disjunctive Normal Form (DNF).
func (f NotF) ToDNF() Formula {
	return ToDNF(f)
}
