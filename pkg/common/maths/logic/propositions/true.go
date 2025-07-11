package propositions

type TrueF struct {
}

func T() Formula {
	return TrueF{}
}

func (TrueF) String() string {
	return "V"
}

func (TrueF) Eval(_ map[string]bool) bool {
	return true
}

func (TrueF) Vars() []string {
	return nil
}

func (t TrueF) And(f Formula) Formula {
	return f
}

func (t TrueF) Or(f Formula) Formula {
	return t
}

func (t TrueF) Not() Formula {
	return FalseF{}
}

func (t TrueF) Implies(f Formula) Formula {
	return f
}

func (t TrueF) Contrapositive(f Formula) Formula {
	return f.Not()
}

func (t TrueF) Iff(f Formula) Formula {
	return IffF{L: t, R: f}
}

func (f TrueF) ToNNF() Formula {
	return ToNNF(f)
}

func (f TrueF) ToCNF() Formula {
	return ToCNF(f)
}

func (f TrueF) ToDNF() Formula {
	return ToDNF(f)
}
