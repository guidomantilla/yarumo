package propositions

type Formula interface {
	String() string
	Eval(env map[string]bool) bool
	Vars() []string

	And(Formula) Formula
	Or(Formula) Formula
	Not() Formula
	Implies(Formula) Formula
	Contrapositive(Formula) Formula
	Iff(Formula) Formula
}
