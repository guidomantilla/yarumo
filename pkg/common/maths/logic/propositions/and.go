package propositions

type AndF struct {
	L, R Formula
}

func (f AndF) String() string {
	return "(" + f.L.String() + " ∧ " + f.R.String() + ")"
}

func (f AndF) Eval(env map[string]bool) bool {
	return f.L.Eval(env) && f.R.Eval(env)
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
