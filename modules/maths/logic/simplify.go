package logic

// Simplify applies algebraic simplification rules to a formula.
// Rules are applied recursively until no further simplifications are possible.
func Simplify(f Formula) Formula {
	prev := f

	for {
		next := simplifyOnce(prev)
		if next.String() == prev.String() {
			return next
		}

		prev = next
	}
}

func simplifyOnce(f Formula) Formula {
	switch v := f.(type) {
	case GroupF:
		// Rule: (A) → A
		return simplifyOnce(v.F)
	case NotF:
		return simplifyNot(v)
	case AndF:
		return simplifyAnd(v)
	case OrF:
		return simplifyOr(v)
	case ImplF:
		// Rule: A => B → !A | B
		return simplifyOnce(OrF{L: NotF{F: v.L}, R: v.R})
	case IffF:
		// Rule: A <=> B → (A & B) | (!A & !B)
		return simplifyOnce(OrF{
			L: AndF(v),
			R: AndF{L: NotF{F: v.L}, R: NotF{F: v.R}},
		})
	default:
		return f
	}
}

func simplifyNot(v NotF) Formula {
	inner := simplifyOnce(v.F)

	switch i := inner.(type) {
	case NotF:
		// Rule: !!A → A
		return simplifyOnce(i.F)
	case TrueF:
		// Rule: !true → false
		return FalseF{}
	case FalseF:
		// Rule: !false → true
		return TrueF{}
	default:
		return NotF{F: inner}
	}
}

func simplifyAnd(v AndF) Formula {
	l := simplifyOnce(v.L)
	r := simplifyOnce(v.R)

	// Rule: A & true → A
	if isTrue(l) {
		return r
	}

	if isTrue(r) {
		return l
	}

	// Rule: A & false → false
	if isFalse(l) || isFalse(r) {
		return FalseF{}
	}

	// Rule: A & A → A
	if l.String() == r.String() {
		return l
	}

	// Rule: A & !A → false
	if isComplement(l, r) {
		return FalseF{}
	}

	return AndF{L: l, R: r}
}

func simplifyOr(v OrF) Formula {
	l := simplifyOnce(v.L)
	r := simplifyOnce(v.R)

	// Rule: A | true → true
	if isTrue(l) || isTrue(r) {
		return TrueF{}
	}

	// Rule: A | false → A
	if isFalse(l) {
		return r
	}

	if isFalse(r) {
		return l
	}

	// Rule: A | A → A
	if l.String() == r.String() {
		return l
	}

	// Rule: A | !A → true
	if isComplement(l, r) {
		return TrueF{}
	}

	return OrF{L: l, R: r}
}

func isTrue(f Formula) bool {
	_, ok := f.(TrueF)
	return ok
}

func isFalse(f Formula) bool {
	_, ok := f.(FalseF)
	return ok
}

func isComplement(a, b Formula) bool {
	n, ok := a.(NotF)
	if ok && n.F.String() == b.String() {
		return true
	}

	n, ok = b.(NotF)

	return ok && n.F.String() == a.String()
}
