package propositions

type FalseF struct {
}

func F() Formula {
	return FalseF{}
}

func (FalseF) String() string {
	return "F"
}

func (FalseF) Eval(_ map[string]bool) bool {
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
