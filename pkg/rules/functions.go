package rules

import (
	"fmt"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Rule[T any] struct {
	Label     string
	Formula   propositions.Formula
	Predicate predicates.Predicate[T]
}

type RuleResult[T any] struct {
	Rule      Rule[T]
	Input     T
	Violated  bool
	Satisfied bool
}

func EvaluateRules[T any](preds logic.Predicates[T], rules []Rule[T], input T) []RuleResult[T] {

	traces := make([]logic.Trace[T], 0)
	var results []RuleResult[T]
	for _, r := range rules {
		result := logic.EvaluateProposition(r.Formula, preds, &input)
		results = append(results, RuleResult[T]{
			Rule:      r,
			Input:     input,
			Violated:  !result.Result,
			Satisfied: result.Result,
		})
	}

	for _, tr := range traces {
		fmt.Printf("Predicate %s => %v\n", tr.Name, tr.Value)
	}

	return results
}

func PrintRuleEvaluation[T any](results []RuleResult[T]) {
	for _, r := range results {
		status := ""
		if r.Satisfied {
			status = "SATISFIED"
		} else if r.Violated {
			status = "VIOLATED"
		}
		fmt.Printf("%s %+v => %s\n", r.Rule.Formula, r.Rule.Label, status)
	}
}
