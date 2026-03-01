package logic

import "slices"

// Vars returns a slice containing only this variable.
func (v Var) Vars() []Var {
	return []Var{v}
}

// Vars returns nil.
func (t TrueF) Vars() []Var {
	return nil
}

// Vars returns nil.
func (f FalseF) Vars() []Var {
	return nil
}

// Vars returns the variables in the negated formula.
func (n NotF) Vars() []Var {
	return n.F.Vars()
}

// Vars returns the sorted, deduplicated variables from both operands.
func (a AndF) Vars() []Var {
	return mergeVars(a.L.Vars(), a.R.Vars())
}

// Vars returns the sorted, deduplicated variables from both operands.
func (o OrF) Vars() []Var {
	return mergeVars(o.L.Vars(), o.R.Vars())
}

// Vars returns the sorted, deduplicated variables from both operands.
func (i ImplF) Vars() []Var {
	return mergeVars(i.L.Vars(), i.R.Vars())
}

// Vars returns the sorted, deduplicated variables from both operands.
func (b IffF) Vars() []Var {
	return mergeVars(b.L.Vars(), b.R.Vars())
}

// Vars returns the variables in the grouped formula.
func (g GroupF) Vars() []Var {
	return g.F.Vars()
}

func mergeVars(a, b []Var) []Var {
	seen := make(map[Var]struct{}, len(a)+len(b))
	for _, v := range a {
		seen[v] = struct{}{}
	}

	for _, v := range b {
		seen[v] = struct{}{}
	}

	result := make([]Var, 0, len(seen))
	for v := range seen {
		result = append(result, v)
	}

	slices.Sort(result)

	return result
}
