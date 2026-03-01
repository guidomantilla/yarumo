package logic

// Eval returns the truth value of the variable in the given assignment.
// Returns false if the variable is not present in facts.
func (v Var) Eval(facts Fact) bool {
	return facts[v]
}

// Eval always returns true.
func (t TrueF) Eval(_ Fact) bool {
	return true
}

// Eval always returns false.
func (f FalseF) Eval(_ Fact) bool {
	return false
}

// Eval returns the negation of the inner formula.
func (n NotF) Eval(facts Fact) bool {
	return !n.F.Eval(facts)
}

// Eval returns true if both operands are true.
func (a AndF) Eval(facts Fact) bool {
	return a.L.Eval(facts) && a.R.Eval(facts)
}

// Eval returns true if at least one operand is true.
func (o OrF) Eval(facts Fact) bool {
	return o.L.Eval(facts) || o.R.Eval(facts)
}

// Eval returns true unless the antecedent is true and the consequent is false.
func (i ImplF) Eval(facts Fact) bool {
	return !i.L.Eval(facts) || i.R.Eval(facts)
}

// Eval returns true if both operands have the same truth value.
func (b IffF) Eval(facts Fact) bool {
	return b.L.Eval(facts) == b.R.Eval(facts)
}

// Eval returns the truth value of the inner formula.
func (g GroupF) Eval(facts Fact) bool {
	return g.F.Eval(facts)
}
