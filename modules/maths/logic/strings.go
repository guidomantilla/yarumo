package logic

const (
	trueStr  = "true"
	falseStr = "false"
)

// String returns the variable name.
func (v Var) String() string {
	return string(v)
}

// String returns "true".
func (t TrueF) String() string {
	return trueStr
}

// String returns "false".
func (f FalseF) String() string {
	return falseStr
}

// String returns the negation in canonical form.
func (n NotF) String() string {
	return "!" + n.F.String()
}

// String returns the conjunction in canonical form.
func (a AndF) String() string {
	return "(" + a.L.String() + " & " + a.R.String() + ")"
}

// String returns the disjunction in canonical form.
func (o OrF) String() string {
	return "(" + o.L.String() + " | " + o.R.String() + ")"
}

// String returns the implication in canonical form.
func (i ImplF) String() string {
	return "(" + i.L.String() + " => " + i.R.String() + ")"
}

// String returns the biconditional in canonical form.
func (b IffF) String() string {
	return "(" + b.L.String() + " <=> " + b.R.String() + ")"
}

// String returns the grouped formula in canonical form.
func (g GroupF) String() string {
	return "(" + g.F.String() + ")"
}
