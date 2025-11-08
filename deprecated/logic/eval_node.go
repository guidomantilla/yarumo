package logic

import (
	"github.com/guidomantilla/yarumo/deprecated/logic/propositions"
)

type EvalNode struct {
	Label   string     `json:"label,omitempty"`
	Expr    string     `json:"expr,omitempty"`
	Value   *bool      `json:"value,omitempty"`
	Nodes   []EvalNode `json:"nodes,omitempty"`
	formula propositions.Formula
}

func NewEvalNode(formula propositions.Formula, value bool, child ...EvalNode) *EvalNode {
	return &EvalNode{
		formula: formula,
		Expr:    formula.String(),
		Value:   &value,
		Nodes:   child,
	}
}

func (x EvalNode) Vars() []string {
	return x.formula.Vars()
}

func (x EvalNode) Satisfied() bool {
	return propositions.IsSatisfiable(x.formula)
}

func (x EvalNode) Contradiction() bool {
	return propositions.IsContradiction(x.formula)
}

func (x EvalNode) Tautology() bool {
	return propositions.IsTautology(x.formula)
}

func (x EvalNode) TruthTable() []propositions.Fact {
	return propositions.TruthTable(x.formula)
}

func (x EvalNode) FailCases() []propositions.Fact {
	return propositions.FailCases(x.formula)
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
