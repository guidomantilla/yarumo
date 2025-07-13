package propositions

// FalseF represents a logical falsehood in propositional logic.
type FalseF struct {
}

// F returns a new instance of the FalseF formula, which represents a logical falsehood.
func F() Formula {
	return FalseF{}
}

// String returns a string representation of the FalseF formula.
func (FalseF) String() string {
	return "F"
}

// Eval evaluates the FalseF formula against a Fact, always returning false.
func (FalseF) Eval(_ Fact) bool {
	return false
}

// Vars return an empty slice, as FalseF does not contain any variables.
func (FalseF) Vars() []string {
	return nil
}

// And returns the FalseF formula itself, as FalseF AND any formula is always false.
func (f FalseF) And(g Formula) Formula {
	return f
}

// Or returns the other formula, as FalseF or any formula is equivalent to that formula.
func (f FalseF) Or(g Formula) Formula {
	return g
}

// Not returns a TrueF formula, as the negation of FalseF is always true.
func (f FalseF) Not() Formula {
	return TrueF{}
}

// Implies returns a TrueF formula, as FalseF implies any formula.
func (f FalseF) Implies(g Formula) Formula {
	return TrueF{}
}

// Contrapositive returns a TrueF formula, as the contrapositive of FalseF is always true.
func (f FalseF) Contrapositive(g Formula) Formula {
	return g.Not()
}

// Iff returns an IffF formula that represents the biconditional relationship between the FalseF formula and another formula.
func (f FalseF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

// ToNNF converts the FalseF formula to Negation Normal Form (NNF).
func (f FalseF) ToNNF() Formula {
	return ToNNF(f)
}

// ToCNF converts the FalseF formula to Conjunctive Normal Form (CNF).
func (f FalseF) ToCNF() Formula {
	return ToCNF(f)
}

// ToDNF converts the FalseF formula to Disjunctive Normal Form (DNF).
func (f FalseF) ToDNF() Formula {
	return ToDNF(f)
}
