package propositions

import (
	"fmt"
	"strings"
)

type literal struct {
	Name    string
	Negated bool
}

func (l literal) String() string {
	if l.Negated {
		return "¬" + l.Name
	}
	return l.Name
}

type clause []literal

type clauses []clause

func toClauses(f Formula) clauses {
	switch x := f.(type) {
	case AndF:
		return append(toClauses(x.L), toClauses(x.R)...)
	case OrF:
		return clauses{flattenOr(x)}
	case NotF:
		if v, ok := x.F.(Var); ok {
			return clauses{{{Name: string(v), Negated: true}}}
		}
	case Var:
		return clauses{{{Name: string(x), Negated: false}}}
	}
	panic("invalid CNF formula")
}

func flattenOr(f Formula) clause {
	switch x := f.(type) {
	case OrF:
		return append(flattenOr(x.L), flattenOr(x.R)...)
	case NotF:
		if v, ok := x.F.(Var); ok {
			return clause{{Name: string(v), Negated: true}}
		}
	case Var:
		return clause{{Name: string(x), Negated: false}}
	}
	panic("invalid clause structure")
}

func resolve(c1, c2 clause) (clause, bool) {
	for i, l1 := range c1 {
		for j, l2 := range c2 {
			if l1.Name == l2.Name && l1.Negated != l2.Negated {
				newclause := append([]literal{}, c1[:i]...)
				newclause = append(newclause, c1[i+1:]...)
				for k, l := range c2 {
					if k != j {
						newclause = append(newclause, l)
					}
				}
				return removeDuplicates(newclause), true
			}
		}
	}
	return nil, false
}

func removeDuplicates(c clause) clause {
	seen := make(map[string]bool)
	out := make(clause, 0, len(c))
	for _, l := range c {
		key := l.String()
		if !seen[key] {
			seen[key] = true
			out = append(out, l)
		}
	}
	return out
}

func Resolution(f Formula) bool {
	clauses := toClauses(ToCNF(f))
	seen := make(map[string]struct{})

	for {
		n := len(clauses)
		newclauses := []clause{}
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				resolvent, ok := resolve(clauses[i], clauses[j])
				if ok {
					if len(resolvent) == 0 {
						return false // contradicción: fórmula insatisfactible
					}
					k := clauseKey(resolvent)
					if _, found := seen[k]; !found {
						seen[k] = struct{}{}
						newclauses = append(newclauses, resolvent)
					}
				}
			}
		}
		if len(newclauses) == 0 {
			return true // no se puede deducir contradicción → satisfactible
		}
		clauses = append(clauses, newclauses...)
	}
}

func clauseKey(c clause) string {
	s := ""
	for _, l := range c {
		s += l.String() + ","
	}
	return s
}

func IsSatisfiable(f Formula) bool {
	return Resolution(f)
}

func IsContradiction(f Formula) bool {
	return !Resolution(f)
}

func ResolutionTrace(f Formula) bool {
	clauses := toClauses(ToCNF(f))
	seen := make(map[string]struct{})
	step := 1

	fmt.Println("Cláusulas iniciales:")
	for i, c := range clauses {
		fmt.Printf("C%d: %s\n", i+1, clauseString(c))
	}

	for {
		n := len(clauses)
		newclauses := []clause{}
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				resolvent, ok := resolve(clauses[i], clauses[j])
				if ok {
					fmt.Printf("Paso %d: resolver C%d y C%d ⇒ %s\n", step, i+1, j+1, clauseString(resolvent))
					step++
					if len(resolvent) == 0 {
						fmt.Println("❌ Contradicción encontrada: cláusula vacía")
						return false
					}
					k := clauseKey(resolvent)
					if _, found := seen[k]; !found {
						seen[k] = struct{}{}
						newclauses = append(newclauses, resolvent)
					}
				}
			}
		}
		if len(newclauses) == 0 {
			fmt.Println("✅ No se encontró contradicción. La fórmula es satisfactible.")
			return true
		}
		clauses = append(clauses, newclauses...)
	}
}

func clauseString(c clause) string {
	parts := make([]string, len(c))
	for i, l := range c {
		parts[i] = l.String()
	}
	return "(" + strings.Join(parts, " ∨ ") + ")"
}
