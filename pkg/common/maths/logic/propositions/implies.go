package propositions

type ImplF struct {
	L, R Formula
}

func (f ImplF) String() string {
	return "(" + f.L.String() + " ⇒ " + f.R.String() + ")"
}

func (f ImplF) Eval(facts Fact) bool {
	return !f.L.Eval(facts) || f.R.Eval(facts)
}

func (f ImplF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

func (f ImplF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

func (f ImplF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

func (f ImplF) Not() Formula {
	return NotF{F: f}
}

func (f ImplF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

func (f ImplF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

func (f ImplF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f ImplF) ToNNF() Formula {
	return ToNNF(f)
}

func (f ImplF) ToCNF() Formula {
	return ToCNF(f)
}

func (f ImplF) ToDNF() Formula {
	return ToDNF(f)
}
