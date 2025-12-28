package engine

import (
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// Engine contains current rules and facts.
type Engine struct {
	Facts FactBase
	Rules []Rule
}

// Assert sets a fact to true.
func (e *Engine) Assert(v p.Var) {
	if e.Facts == nil {
		e.Facts = FactBase{}
	}

	e.Facts[v] = true
}

// Retract removes a fact.
func (e *Engine) Retract(v p.Var) {
	if e.Facts != nil {
		delete(e.Facts, v)
	}
}

// FireOnce evaluates rules in a single pass and returns fired rule IDs.
func (e *Engine) FireOnce() (fired []string) {
	if e.Facts == nil {
		e.Facts = FactBase{}
	}

	for _, r := range e.Rules {
		if shouldFire(r, e.Facts) {
			// Fire: set Then to true if it wasn't already true
			if !e.Facts[r.then] {
				e.Facts[r.then] = true
				fired = append(fired, r.id)
			}
		}
	}

	return fired
}

// shouldFire determines whether a rule should fire under the current facts.
// Special handling:
// - If When is an implication (A => B) and Then == B, fire when A is true.
// - If When is a biconditional (A <=> B) and Then == A (or B), fire when the other side is true.
// Otherwise, default to evaluating When directly.
func shouldFire(r Rule, facts FactBase) bool {
	switch w := r.when.(type) {
	case p.ImplF:
		if isVarEqual(w.R, r.then) {
			return w.L.Eval(p.Fact(facts))
		}
	case p.IffF:
		if isVarEqual(w.L, r.then) {
			return w.R.Eval(p.Fact(facts))
		}

		if isVarEqual(w.R, r.then) {
			return w.L.Eval(p.Fact(facts))
		}
	}

	return r.when.Eval(p.Fact(facts))
}

// isVarEqual reports whether formula f is exactly the variable v (possibly wrapped in a GroupF).
func isVarEqual(f p.Formula, v p.Var) bool {
	switch x := f.(type) {
	case p.Var:
		return x == v
	case p.GroupF:
		return isVarEqual(x.Inner, v)
	default:
		return false
	}
}

// RunToFixpoint iterates FireOnce until convergence or maxIters is reached.
func (e *Engine) RunToFixpoint(maxIters int) (fired []string) {
	if maxIters <= 0 {
		maxIters = 1
	}

	for range maxIters {
		step := e.FireOnce()
		if len(step) == 0 {
			break
		}

		fired = append(fired, step...)
	}

	return fired
}

// Query evaluates a goal against the current facts and produces an explanation tree.
// It attempts to reconstruct a derivation using rules: when the goal (or subgoals)
// are variables made true by some rule's Then, it expands the explanation to include
// the corresponding implication (When => Then) and recursively explains the antecedent.
func (e *Engine) Query(goal p.Formula) (bool, *Explain) {
	seen := make(map[p.Var]bool)
	exp, val := explainWithRules(goal, e.Facts, e.Rules, seen)

	return val, exp
}
