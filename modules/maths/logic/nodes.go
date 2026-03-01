package logic

// Type compliance: all node types implement Formula.
var (
	_ Formula = Var("")
	_ Formula = TrueF{}
	_ Formula = FalseF{}
	_ Formula = NotF{}
	_ Formula = AndF{}
	_ Formula = OrF{}
	_ Formula = ImplF{}
	_ Formula = IffF{}
	_ Formula = GroupF{}
)

// TrueF represents a tautology (top).
type TrueF struct{}

// FalseF represents a contradiction (bottom).
type FalseF struct{}

// NotF represents negation.
type NotF struct {
	F Formula
}

// AndF represents conjunction.
type AndF struct {
	L, R Formula
}

// OrF represents disjunction.
type OrF struct {
	L, R Formula
}

// ImplF represents implication.
type ImplF struct {
	L, R Formula
}

// IffF represents biconditional.
type IffF struct {
	L, R Formula
}

// GroupF represents a parenthesized formula.
type GroupF struct {
	F Formula
}
