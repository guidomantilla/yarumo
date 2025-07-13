package propositions

type NotF struct {
	F Formula
}

func (f NotF) String() string {
	return "!" + f.F.String()
}

func (f NotF) Eval(facts Fact) bool {
	return !f.F.Eval(facts)
}

func (f NotF) Vars() []string {
	return f.F.Vars()
}

func (f NotF) And(g Formula) Formula {
	return AndF{L: f, R: g}
}

func (f NotF) Or(g Formula) Formula {
	return OrF{L: f, R: g}
}

func (f NotF) Not() Formula {
	return NotF{F: f}
}

func (f NotF) Implies(g Formula) Formula {
	return ImplF{L: f, R: g}
}

func (f NotF) Contrapositive(g Formula) Formula {
	return ImplF{L: g.Not(), R: f.Not()}
}

func (f NotF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f NotF) ToNNF() Formula {
	return ToNNF(f)
}

func (f NotF) ToCNF() Formula {
	return ToCNF(f)
}

func (f NotF) ToDNF() Formula {
	return ToDNF(f)
}
