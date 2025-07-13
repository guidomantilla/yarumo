package propositions

type AndF struct {
	L, R Formula
}

func (f AndF) String() string {
	return "(" + f.L.String() + " ∧ " + f.R.String() + ")"
}

func (f AndF) Eval(facts Fact) bool {
	return f.L.Eval(facts) && f.R.Eval(facts)
}

func (f AndF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

func (f AndF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

func (f AndF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

func (f AndF) Not() Formula {
	return NotF{F: f}
}

func (f AndF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

func (f AndF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

func (f AndF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f AndF) ToNNF() Formula {
	return ToNNF(f)
}

func (f AndF) ToCNF() Formula {
	return ToCNF(f)
}

func (f AndF) ToDNF() Formula {
	return ToDNF(f)
}
