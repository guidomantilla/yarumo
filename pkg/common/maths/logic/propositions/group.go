package propositions

type GroupF struct {
	Inner Formula
}

func Group(f Formula) Formula {
	return GroupF{Inner: f}
}

func (g GroupF) String() string {
	return "(" + g.Inner.String() + ")"
}

func (g GroupF) Eval(env map[string]bool) bool {
	return g.Inner.Eval(env)
}

func (g GroupF) Vars() []string {
	return g.Inner.Vars()
}

func (g GroupF) And(f Formula) Formula {
	return AndF{L: g, R: f}
}

func (g GroupF) Or(f Formula) Formula {
	return OrF{L: g, R: f}
}

func (g GroupF) Not() Formula {
	return NotF{F: g}
}

func (g GroupF) Implies(f Formula) Formula {
	return ImplF{L: g, R: f}
}

func (g GroupF) Contrapositive(f Formula) Formula {
	return ImplF{L: f.Not(), R: g.Not()}
}

func (g GroupF) Iff(f Formula) Formula {
	return IffF{L: g, R: f}
}

func (f GroupF) ToNNF() Formula {
	return ToNNF(f)
}

func (f GroupF) ToCNF() Formula {
	return ToCNF(f)
}

func (f GroupF) ToDNF() Formula {
	return ToDNF(f)
}
