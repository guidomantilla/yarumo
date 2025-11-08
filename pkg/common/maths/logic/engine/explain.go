package engine

import (
	"fmt"
	"strings"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"
)

// Explain is a minimal structure for traces.
type Explain struct {
	ID    string
	Expr  string
	Value bool
	Why   string
	Kids  []*Explain
}

// PrettyExplain renders an explanation tree into a string by delegating to PrettyExplainTo.
func PrettyExplain(e *Explain) string {
	var sb strings.Builder
	PrettyExplainTo(&sb, e)
	return sb.String()
}

// --- internal helpers ---

// explainWithRules builds an explanation tree using facts and tries to expand
// variable facts via rules that produced them (When => Then). It uses a visited
// set to avoid infinite recursion in the presence of cycles.
func explainWithRules(f props.Formula, facts FactBase, rules []Rule, seen map[props.Var]bool) (*Explain, bool) {
	switch x := f.(type) {
	case props.TrueF:
		return &Explain{Expr: x.String(), Value: true, Why: "constant true"}, true
	case props.FalseF:
		return &Explain{Expr: x.String(), Value: false, Why: "constant false"}, false
	case props.Var:
		val := facts[x]
		// If true, try to find a supporting rule and expand
		if val && !seen[x] {
			if r, ok := findSupportingRule(x, facts, rules); ok {
				seen[x] = true
				// Build implication expression: (When => Var)
				imp := props.ImplF{L: r.when, R: x}
				ante, av := explainWithRules(r.when, facts, rules, seen)
				// Leaf for the fact itself
				leaf := &Explain{Expr: x.String(), Value: true, Why: fmt.Sprintf("fact: %s=true", x)}
				why := "rule fired"
				if !av {
					why = "implication holds"
				}
				return &Explain{Expr: imp.String(), Value: true, Why: why, Kids: []*Explain{ante, leaf}}, true
			}
		}
		why := fmt.Sprintf("fact: %s=%v", x.String(), val)
		return &Explain{Expr: x.String(), Value: val, Why: why}, val
	case props.NotF:
		kid, v := explainWithRules(x.F, facts, rules, seen)
		return &Explain{Expr: x.String(), Value: !v, Why: fmt.Sprintf("negation of %v", v), Kids: []*Explain{kid}}, !v
	case props.AndF:
		l, lv := explainWithRules(x.L, facts, rules, seen)
		r, rv := explainWithRules(x.R, facts, rules, seen)
		val := lv && rv
		why := "both true"
		if !lv && !rv {
			why = "both false"
		} else if !lv {
			why = "left false"
		} else if !rv {
			why = "right false"
		}
		return &Explain{Expr: x.String(), Value: val, Why: why, Kids: []*Explain{l, r}}, val
	case props.OrF:
		l, lv := explainWithRules(x.L, facts, rules, seen)
		r, rv := explainWithRules(x.R, facts, rules, seen)
		val := lv || rv
		why := "at least one true"
		if !lv && !rv {
			why = "both false"
		}
		return &Explain{Expr: x.String(), Value: val, Why: why, Kids: []*Explain{l, r}}, val
	case props.ImplF:
		l, lv := explainWithRules(x.L, facts, rules, seen)
		r, rv := explainWithRules(x.R, facts, rules, seen)
		val := (!lv) || rv
		why := "implication holds"
		if lv && !rv {
			why = "left true and right false"
		}
		return &Explain{Expr: x.String(), Value: val, Why: why, Kids: []*Explain{l, r}}, val
	case props.IffF:
		l, lv := explainWithRules(x.L, facts, rules, seen)
		r, rv := explainWithRules(x.R, facts, rules, seen)
		val := lv == rv
		why := "both equal"
		if lv != rv {
			why = "different truth values"
		}
		return &Explain{Expr: x.String(), Value: val, Why: why, Kids: []*Explain{l, r}}, val
	case props.GroupF:
		kid, v := explainWithRules(x.Inner, facts, rules, seen)
		return &Explain{Expr: x.String(), Value: v, Why: "group", Kids: []*Explain{kid}}, v
	default:
		return &Explain{Expr: f.String(), Value: false, Why: "unknown node"}, false
	}
}

// findSupportingRule finds the first rule whose Then matches v and whose When
// evaluates to true under the given facts.
func findSupportingRule(v props.Var, facts FactBase, rules []Rule) (*Rule, bool) {
	for i := range rules {
		r := &rules[i]
		if r.then == v && r.when.Eval(props.Fact(facts)) {
			return r, true
		}
	}
	return nil, false
}
