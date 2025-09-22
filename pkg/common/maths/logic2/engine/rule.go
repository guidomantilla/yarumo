package engine

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// Rule represents a propositional rule with a literal consequence (MVP).
type Rule struct {
	id   string
	when props.Formula
	then props.Var
	rule props.Formula
}

func BuildRule(id string, when string, then string) Rule {
	return Rule{
		id:   id,
		when: parser.MustParse(when),
		then: props.Var(then),
		rule: parser.MustParse(fmt.Sprintf("(%s) => %s", when, then)),
	}
}

func (r *Rule) Equals(rule Rule) bool {
	return r.id == r.id && r.then == r.then && props.Equivalent(r.when, rule.when)
}

func (r *Rule) String() string {
	return r.rule.String()
}
