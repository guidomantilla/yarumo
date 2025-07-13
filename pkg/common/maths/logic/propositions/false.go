package propositions

type FalseF struct {
}

func F() Formula {
	return FalseF{}
}

func (FalseF) String() string {
	return "F"
}

func (FalseF) Eval(_ Fact) bool {
	return false
}

func (FalseF) Vars() []string {
	return nil
}

func (f FalseF) And(g Formula) Formula {
	return f
}

func (f FalseF) Or(g Formula) Formula {
	return g
}

func (f FalseF) Not() Formula {
	return TrueF{}
}

func (f FalseF) Implies(g Formula) Formula {
	return TrueF{}
}

func (f FalseF) Contrapositive(g Formula) Formula {
	return g.Not()
}

func (f FalseF) Iff(g Formula) Formula {
	return IffF{L: f, R: g}
}

func (f FalseF) ToNNF() Formula {
	return ToNNF(f)
}

func (f FalseF) ToCNF() Formula {
	return ToCNF(f)
}

func (f FalseF) ToDNF() Formula {
	return ToDNF(f)
}
