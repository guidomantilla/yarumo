package logic

// ToNNF converts a formula to Negation Normal Form.
// In NNF, negations appear only before variables, and the only connectives
// are AND, OR, and NOT. Implications and biconditionals are eliminated.
func ToNNF(f Formula) Formula {
	return toNNF(f, false)
}

// ToCNF converts a formula to Conjunctive Normal Form.
// The result is a conjunction of disjunctions (AND of ORs).
func ToCNF(f Formula) Formula {
	return distributeCNF(ToNNF(f))
}

// ToDNF converts a formula to Disjunctive Normal Form.
// The result is a disjunction of conjunctions (OR of ANDs).
func ToDNF(f Formula) Formula {
	return distributeDNF(ToNNF(f))
}

func toNNF(f Formula, neg bool) Formula {
	switch v := f.(type) {
	case Var:
		if neg {
			return NotF{F: v}
		}

		return v
	case TrueF:
		if neg {
			return FalseF{}
		}

		return v
	case FalseF:
		if neg {
			return TrueF{}
		}

		return v
	case NotF:
		return toNNF(v.F, !neg)
	case AndF:
		if neg {
			return OrF{L: toNNF(v.L, true), R: toNNF(v.R, true)}
		}

		return AndF{L: toNNF(v.L, false), R: toNNF(v.R, false)}
	case OrF:
		if neg {
			return AndF{L: toNNF(v.L, true), R: toNNF(v.R, true)}
		}

		return OrF{L: toNNF(v.L, false), R: toNNF(v.R, false)}
	case ImplF:
		if neg {
			// !(A => B) ≡ A & !B
			return AndF{L: toNNF(v.L, false), R: toNNF(v.R, true)}
		}

		// A => B ≡ !A | B
		return OrF{L: toNNF(v.L, true), R: toNNF(v.R, false)}
	case IffF:
		if neg {
			// !(A <=> B) ≡ (A & !B) | (!A & B)
			return OrF{
				L: AndF{L: toNNF(v.L, false), R: toNNF(v.R, true)},
				R: AndF{L: toNNF(v.L, true), R: toNNF(v.R, false)},
			}
		}

		// A <=> B ≡ (!A | B) & (!B | A)
		return AndF{
			L: OrF{L: toNNF(v.L, true), R: toNNF(v.R, false)},
			R: OrF{L: toNNF(v.R, true), R: toNNF(v.L, false)},
		}
	case GroupF:
		return toNNF(v.F, neg)
	default:
		return f
	}
}

func distributeCNF(f Formula) Formula {
	switch v := f.(type) {
	case AndF:
		return AndF{L: distributeCNF(v.L), R: distributeCNF(v.R)}
	case OrF:
		l := distributeCNF(v.L)
		r := distributeCNF(v.R)

		return distributeOrOverAnd(l, r)
	default:
		return f
	}
}

func distributeOrOverAnd(l, r Formula) Formula {
	la, ok := l.(AndF)
	if ok {
		// (A & B) | C ≡ (A | C) & (B | C)
		return AndF{
			L: distributeOrOverAnd(la.L, r),
			R: distributeOrOverAnd(la.R, r),
		}
	}

	ra, ok := r.(AndF)
	if ok {
		// A | (B & C) ≡ (A | B) & (A | C)
		return AndF{
			L: distributeOrOverAnd(l, ra.L),
			R: distributeOrOverAnd(l, ra.R),
		}
	}

	return OrF{L: l, R: r}
}

func distributeDNF(f Formula) Formula {
	switch v := f.(type) {
	case OrF:
		return OrF{L: distributeDNF(v.L), R: distributeDNF(v.R)}
	case AndF:
		l := distributeDNF(v.L)
		r := distributeDNF(v.R)

		return distributeAndOverOr(l, r)
	default:
		return f
	}
}

func distributeAndOverOr(l, r Formula) Formula {
	lo, ok := l.(OrF)
	if ok {
		// (A | B) & C ≡ (A & C) | (B & C)
		return OrF{
			L: distributeAndOverOr(lo.L, r),
			R: distributeAndOverOr(lo.R, r),
		}
	}

	ro, ok := r.(OrF)
	if ok {
		// A & (B | C) ≡ (A & B) | (A & C)
		return OrF{
			L: distributeAndOverOr(l, ro.L),
			R: distributeAndOverOr(l, ro.R),
		}
	}

	return AndF{L: l, R: r}
}
