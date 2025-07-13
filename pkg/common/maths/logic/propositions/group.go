package propositions

import "fmt"

// GroupF represents a logical grouping of a formula, allowing for clearer expression of complex logical statements.
type GroupF struct {
	Inner Formula
}

// Group creates a new GroupF formula that encapsulates the provided formula f.
func Group(f Formula) Formula {
	return GroupF{Inner: f}
}

// String returns a string representation of the GroupF formula, enclosing the inner formula in parentheses.
func (g GroupF) String() string {
	return fmt.Sprintf("(%s)", g.Inner.String())
}

// Eval evaluates the GroupF formula against a set of facts, delegating the evaluation to the inner formula.
func (g GroupF) Eval(facts Fact) bool {
	return g.Inner.Eval(facts)
}

// Vars returns a slice of variable names used in the inner formula of the GroupF.
func (g GroupF) Vars() []string {
	return g.Inner.Vars()
}

// And returns a new AndF formula that combines the GroupF with another formula using logical conjunction.
func (g GroupF) And(f Formula) Formula {
	return AndF{L: g, R: f}
}

// Or returns a new OrF formula that combines the GroupF with another formula using logical disjunction.
func (g GroupF) Or(f Formula) Formula {
	return OrF{L: g, R: f}
}

// Not returns a NotF formula that negates the GroupF formula.
func (g GroupF) Not() Formula {
	return NotF{F: g}
}

// Implies returns an ImplF formula that represents the implication of the GroupF formula to another formula.
func (g GroupF) Implies(f Formula) Formula {
	return ImplF{L: g, R: f}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the GroupF formula with respect to another formula.
func (g GroupF) Contrapositive(f Formula) Formula {
	return ImplF{L: f.Not(), R: g.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the GroupF formula and another formula.
func (g GroupF) Iff(f Formula) Formula {
	return IffF{L: g, R: f}
}

// ToNNF converts the GroupF formula to Negation Normal Form (NNF).
func (g GroupF) ToNNF() Formula {
	return ToNNF(g)
}

// ToCNF converts the GroupF formula to Conjunctive Normal Form (CNF).
func (g GroupF) ToCNF() Formula {
	return ToCNF(g)
}

// ToDNF converts the GroupF formula to Disjunctive Normal Form (DNF).
func (g GroupF) ToDNF() Formula {
	return ToDNF(g)
}
