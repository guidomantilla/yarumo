package propositions

type OrF struct {
	L, R Formula
}

func (f OrF) String() string {
	return "(" + f.L.String() + " ∨ " + f.R.String() + ")"
}

func (f OrF) Eval(env map[string]bool) bool {
	return f.L.Eval(env) || f.R.Eval(env)
}

func (f OrF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

func (f OrF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

func (f OrF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

func (f OrF) Not() Formula {
	return NotF{F: f}
}

func (f OrF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

func (f OrF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

func (f OrF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f OrF) ToNNF() Formula {
	return ToNNF(f)
}

func (f OrF) ToCNF() Formula {
	return ToCNF(f)
}

func (f OrF) ToDNF() Formula {
	return ToDNF(f)
}
