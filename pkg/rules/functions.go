package rules

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

func Unwrap[T any](rules []Rule[T]) (propositions.Formula, predicates.Predicate[T], error) {
	if len(rules) == 0 {
		return nil, nil, fmt.Errorf("no rules provided")
	}
	formula, predicate := rules[0].Formula, rules[0].Predicate
	for _, rule := range rules[1:] {
		formula = formula.And(rule.Formula)
		predicate = predicate.And(rule.Predicate)
	}
	return formula, predicate, nil
}
