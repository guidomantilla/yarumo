package propositions

type IffF struct {
	L, R Formula
}

func (f IffF) String() string {
	return "(" + f.L.String() + " ⇔ " + f.R.String() + ")"
}

func (f IffF) Eval(env map[string]bool) bool {
	return f.L.Eval(env) == f.R.Eval(env)
}

func (f IffF) Vars() []string {
	return union(f.L.Vars(), f.R.Vars())
}

func (f IffF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

func (f IffF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

func (f IffF) Not() Formula {
	return NotF{F: f}
}

func (f IffF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

func (f IffF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

func (f IffF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f IffF) ToNNF() Formula {
	return ToNNF(f)
}

func (f IffF) ToCNF() Formula {
	return ToCNF(f)
}

func (f IffF) ToDNF() Formula {
	return ToDNF(f)
}
