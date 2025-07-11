package maths

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"

type Proposition struct {
	propositions.Formula
}

func NewProposition(variable string) *Proposition {
	return &Proposition{
		Formula: propositions.V(variable),
	}
}

func (p *Proposition) TruthTable() []map[string]bool {
	return propositions.TruthTable(p.Formula)
}

func (p *Proposition) Equivalent(f propositions.Formula) bool {
	return propositions.Equivalent(p.Formula, f)
}

func (p *Proposition) ToNNF() propositions.Formula {
	return propositions.ToNNF(p.Formula)
}

func (p *Proposition) ToCNF() propositions.Formula {
	return propositions.ToCNF(p.Formula)
}

func (p *Proposition) ToDNF() propositions.Formula {
	return propositions.ToDNF(p.Formula)
}

func (p *Proposition) PrettyPrint() {
	propositions.PrintTruthTable(p.Formula)
}

func (p *Proposition) Resolution() bool {
	return propositions.Resolution(p.Formula)
}

func (p *Proposition) ResolutionTrace() bool {
	return propositions.ResolutionTrace(p.Formula)
}
