package engine

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// Rule represents a propositional rule with a literal consequence (MVP).
type Rule struct {
	id   string
	when p.Formula
	then p.Var
	rule p.Formula
}

func BuildRule(id string, when string, then string) Rule {
	return Rule{
		id:   id,
		when: parser.MustParse(when),
		then: p.Var(then),
		rule: parser.MustParse(fmt.Sprintf("(%s) => %s", when, then)),
	}
}

func (r *Rule) Equals(rule Rule) bool {
	return r.id == rule.id && r.then == rule.then && p.Equivalent(r.when, rule.when)
}

func (r *Rule) String() string {
	return r.rule.String()
}
