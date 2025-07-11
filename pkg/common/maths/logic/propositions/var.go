package propositions

type Var string

func V(name string) Formula {
	return Var(name)
}

func (v Var) String() string {
	return string(v)
}

func (v Var) Eval(env map[string]bool) bool {
	return env[string(v)]
}

func (v Var) Vars() []string {
	return []string{string(v)}
}

func (v Var) And(f Formula) Formula {
	return AndF{L: v, R: f}
}

func (v Var) Or(f Formula) Formula {
	return OrF{L: v, R: f}
}

func (v Var) Not() Formula {
	return NotF{F: v}
}

func (v Var) Implies(f Formula) Formula {
	return ImplF{L: v, R: f}
}

func (v Var) Contrapositive(f Formula) Formula {
	return ImplF{L: f.Not(), R: v.Not()}
}

func (v Var) Iff(f Formula) Formula {
	return IffF{L: v, R: f}
}
