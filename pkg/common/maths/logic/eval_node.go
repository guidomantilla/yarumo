package logic

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type EvalNode struct {
	Label string     `json:"label,omitempty"`
	Expr  string     `json:"expr,omitempty"`
	Value *bool      `json:"value,omitempty"`
	Nodes []EvalNode `json:"nodes,omitempty"`
	f     propositions.Formula
}

func NewEvalNode(f propositions.Formula, value bool, child ...EvalNode) *EvalNode {
	return &EvalNode{
		f:     f,
		Expr:  f.String(),
		Value: &value,
		Nodes: child,
	}
}

func (x EvalNode) Vars() []string {
	return x.f.Vars()
}

func (x EvalNode) Satisfied() bool {
	return propositions.IsSatisfiable(x.f)
}

func (x EvalNode) Contradiction() bool {
	return propositions.IsContradiction(x.f)
}

func (x EvalNode) Tautology() bool {
	return propositions.IsTautology(x.f)
}

func (x EvalNode) TruthTable() []propositions.Fact {
	return propositions.TruthTable(x.f)
}

func (x EvalNode) FailCases() []propositions.Fact {
	return propositions.FailCases(x.f)
}

func (x EvalNode) Facts() propositions.Fact {
	facts := make(propositions.Fact)
	x.collectFacts(facts)
	return facts
}

func (x EvalNode) collectFacts(facts propositions.Fact) {

	if len(x.Nodes) == 0 && x.Value != nil {
		facts[propositions.Var(x.Expr)] = *x.Value
		return
	}

	for _, child := range x.Nodes {
		child.collectFacts(facts)
	}
}
