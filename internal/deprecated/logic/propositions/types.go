package propositions

type Fact map[Var]bool

type Formula interface {
	String() string
	Eval(facts Fact) bool
	Vars() []string

	And(Formula) Formula
	Or(Formula) Formula
	Not() Formula
	Implies(Formula) Formula
	Contrapositive(Formula) Formula
	Iff(Formula) Formula

	ToNNF() Formula
	ToCNF() Formula
	ToDNF() Formula
}
