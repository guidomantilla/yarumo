package logic

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type EvalNode struct {
	Expr          string              `json:"expr"`
	Value         bool                `json:"value"`
	Vars          []string            `json:"vars"`
	Satisfied     bool                `json:"satisfied"`
	Contradiction bool                `json:"contradiction"`
	Tautology     bool                `json:"tautology"`
	TruthTable    []propositions.Fact `json:"truth_table"`
	FailCases     []propositions.Fact `json:"fail_cases"`
	Nodes         []EvalNode          `json:"nodes"`
}

func NewEvalNode(x propositions.Formula, value bool, child ...EvalNode) *EvalNode {
	return &EvalNode{
		Expr:          x.String(),
		Value:         value,
		Nodes:         child,
		Vars:          x.Vars(),
		TruthTable:    propositions.TruthTable(x),
		Satisfied:     propositions.IsSatisfiable(x),
		Contradiction: propositions.IsContradiction(x),
		Tautology:     propositions.IsTautology(x),
		FailCases:     propositions.FailCases(x),
	}
}

func (x EvalNode) Facts() propositions.Fact {
	facts := make(propositions.Fact)
	x.collectFacts(facts)
	return facts
}

func (x EvalNode) collectFacts(facts propositions.Fact) {

	if len(x.Nodes) == 0 {
		facts[x.Expr] = x.Value
		return
	}

	for _, child := range x.Nodes {
		child.collectFacts(facts)
	}
}
