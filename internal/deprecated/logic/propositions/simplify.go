package propositions

// Simplify reduce a formula to its simplest form.
func Simplify(f Formula) Formula {
	prev := Formula(nil)
	cur := simplifyOnce(f)
	for !structuralEqual(cur, prev) {
		prev = cur
		cur = simplifyOnce(cur)
	}
	return cur
}

func simplifyOnce(f Formula) Formula {
	switch x := f.(type) {
	case GroupF:
		return Simplify(x.Inner)

	case TrueF, FalseF, Var:
		return x

	case NotF:
		inner := Simplify(x.F)
		switch y := inner.(type) {
		case TrueF:
			return FalseF{}
		case FalseF:
			return TrueF{}
		case NotF: // ¬(¬A) → A
			return Simplify(y.F)
		default:
			return NotF{F: inner}
		}

	case AndF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		// Constantes
		if isFalse(L) || isFalse(R) {
			return FalseF{}
		}
		if isTrue(L) {
			return R
		}
		if isTrue(R) {
			return L
		}
		// Idempotencia
		if structuralEqual(L, R) {
			return L
		}
		// Complemento: A ∧ ¬A → F
		if isNegationOf(L, R) || isNegationOf(R, L) {
			return FalseF{}
		}
		// Absorción: A ∧ (A ∨ B) → A; (A ∨ B) ∧ A → A
		if rOr, ok := R.(OrF); ok {
			if structuralEqual(L, rOr.L) || structuralEqual(L, rOr.R) {
				return L
			}
		}
		if lOr, ok := L.(OrF); ok {
			if structuralEqual(R, lOr.L) || structuralEqual(R, lOr.R) {
				return R
			}
		}
		return AndF{L: L, R: R}

	case OrF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		// Constantes
		if isTrue(L) || isTrue(R) {
			return TrueF{}
		}
		if isFalse(L) {
			return R
		}
		if isFalse(R) {
			return L
		}
		// Idempotencia
		if structuralEqual(L, R) {
			return L
		}
		// Complemento: A ∨ ¬A → T
		if isNegationOf(L, R) || isNegationOf(R, L) {
			return TrueF{}
		}
		// Absorción: A ∨ (A ∧ B) → A; (A ∧ B) ∨ A → A
		if rAnd, ok := R.(AndF); ok {
			if structuralEqual(L, rAnd.L) || structuralEqual(L, rAnd.R) {
				return L
			}
		}
		if lAnd, ok := L.(AndF); ok {
			if structuralEqual(R, lAnd.L) || structuralEqual(R, lAnd.R) {
				return R
			}
		}
		return OrF{L: L, R: R}

	case ImplF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		// Casos con constantes y trivialidades
		if isFalse(L) {
			return TrueF{}
		}
		if isTrue(L) {
			return R
		}
		if isTrue(R) {
			return TrueF{}
		}
		if isFalse(R) {
			return NotF{F: L}
		}
		if structuralEqual(L, R) {
			return TrueF{}
		}
		return ImplF{L: L, R: R}

	case IffF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		if structuralEqual(L, R) {
			return TrueF{}
		}
		if isTrue(L) {
			return R
		}
		if isTrue(R) {
			return L
		}
		if isFalse(L) {
			return NotF{F: R}
		}
		if isFalse(R) {
			return NotF{F: L}
		}
		return IffF{L: L, R: R}
	}
	return f
}

// --- Helpers ---

func isTrue(f Formula) bool { _, ok := f.(TrueF); return ok }

func isFalse(f Formula) bool { _, ok := f.(FalseF); return ok }

// isNegationOf checks if a is the negation of b.
func isNegationOf(a Formula, b Formula) bool {
	na, ok := a.(NotF)
	if !ok {
		return false
	}
	return structuralEqual(na.F, b)
}

// structuralEqual compares two formulas for structural equality.
func structuralEqual(a, b Formula) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch x := a.(type) {
	case Var:
		y, ok := b.(Var)
		return ok && x == y
	case TrueF:
		_, ok := b.(TrueF)
		return ok
	case FalseF:
		_, ok := b.(FalseF)
		return ok
	case NotF:
		y, ok := b.(NotF)
		return ok && structuralEqual(x.F, y.F)
	case AndF:
		y, ok := b.(AndF)
		return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case OrF:
		y, ok := b.(OrF)
		return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case ImplF:
		y, ok := b.(ImplF)
		return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case IffF:
		y, ok := b.(IffF)
		return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case GroupF:
		// No debería quedar si Simplify ya los elimina, pero lo soportamos
		y, ok := b.(GroupF)
		return ok && structuralEqual(x.Inner, y.Inner)
	default:
		return false
	}
}
