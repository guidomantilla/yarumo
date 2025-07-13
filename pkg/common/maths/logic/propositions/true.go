package propositions

// TrueF represents a logical truth in propositional logic.
type TrueF struct {
}

// T returns a new instance of the TrueF formula, which represents a logical truth.
func T() Formula {
	return TrueF{}
}

// String returns a string representation of the TrueF formula.
func (t TrueF) String() string {
	return "T"
}

// Eval evaluates the TrueF formula against a Fact, always returning true.
func (t TrueF) Eval(_ Fact) bool {
	return true
}

// Vars returns an empty slice, as TrueF does not contain any variables.
func (t TrueF) Vars() []string {
	return nil
}

// And returns the other formula, as TrueF AND any formula is equivalent to that formula.
func (t TrueF) And(f Formula) Formula {
	return f
}

// Or returns the TrueF formula itself, as TrueF OR any formula is always true.
func (t TrueF) Or(f Formula) Formula {
	return t
}

// Not returns a FalseF formula, as the negation of TrueF is always false.
func (t TrueF) Not() Formula {
	return FalseF{}
}

// Implies returns the TrueF formula itself, as TrueF implies any formula.
func (t TrueF) Implies(f Formula) Formula {
	return f
}

// Contrapositive returns a FalseF formula, as the contrapositive of TrueF is always false.
func (t TrueF) Contrapositive(f Formula) Formula {
	return f.Not()
}

// Iff returns an IffF formula that represents the biconditional relationship between the TrueF formula and another formula.
func (t TrueF) Iff(f Formula) Formula {
	return IffF{L: t, R: f}
}

// ToNNF converts the TrueF formula to Negation Normal Form (NNF).
func (t TrueF) ToNNF() Formula {
	return ToNNF(t)
}

// ToCNF converts the TrueF formula to Conjunctive Normal Form (CNF).
func (t TrueF) ToCNF() Formula {
	return ToCNF(t)
}

// ToDNF converts the TrueF formula to Disjunctive Normal Form (DNF).
func (t TrueF) ToDNF() Formula {
	return ToDNF(t)
}
