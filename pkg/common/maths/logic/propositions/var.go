package propositions

// Var represents a propositional variable in logic.
type Var string

// V creates a new Var instance representing a propositional variable with the given name.
func V(name string) Formula {
	return Var(name)
}

// String returns a string representation of the Var, which is simply its name.
func (v Var) String() string {
	return string(v)
}

// Eval evaluates the Var against a set of facts, returning true if the fact corresponding to the variable's name is true.
func (v Var) Eval(facts Fact) bool {
	return facts[v]
}

// Vars returns a slice containing the name of the variable, as it is the only variable in this formula.
func (v Var) Vars() []string {
	return []string{string(v)}
}

// And returns a new AndF formula that combines the Var with another formula using logical conjunction.
func (v Var) And(f Formula) Formula {
	return AndF{L: v, R: f}
}

// Or returns a new OrF formula that combines the Var with another formula using logical disjunction.
func (v Var) Or(f Formula) Formula {
	return OrF{L: v, R: f}
}

// Not returns a NotF formula that negates the Var, effectively flipping its truth value.
func (v Var) Not() Formula {
	return NotF{F: v}
}

// Implies returns an ImplF formula that represents the implication of the Var to another formula.
func (v Var) Implies(f Formula) Formula {
	return ImplF{L: v, R: f}
}

// Contrapositive returns an ImplF formula that represents the contrapositive of the Var with respect to another formula.
func (v Var) Contrapositive(f Formula) Formula {
	return ImplF{L: f.Not(), R: v.Not()}
}

// Iff returns an IffF formula that represents the biconditional relationship between the Var and another formula.
func (v Var) Iff(f Formula) Formula {
	return IffF{L: v, R: f}
}

// ToNNF converts the Var to Negation Normal Form (NNF).
func (v Var) ToNNF() Formula {
	return ToNNF(v)
}

// ToCNF converts the Var to Conjunctive Normal Form (CNF).
func (v Var) ToCNF() Formula {
	return ToCNF(v)
}

// ToDNF converts the Var to Disjunctive Normal Form (DNF).
func (v Var) ToDNF() Formula {
	return ToDNF(v)
}
