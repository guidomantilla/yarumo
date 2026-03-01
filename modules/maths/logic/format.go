package logic

// Format returns a Unicode-formatted string representation of the formula.
// Uses standard mathematical symbols: ∧ (and), ∨ (or), ¬ (not), → (implies),
// ↔ (iff), ⊤ (true), ⊥ (false).
func Format(f Formula) string {
	switch v := f.(type) {
	case Var:
		return string(v)
	case TrueF:
		return "⊤"
	case FalseF:
		return "⊥"
	case NotF:
		return "¬" + Format(v.F)
	case AndF:
		return "(" + Format(v.L) + " ∧ " + Format(v.R) + ")"
	case OrF:
		return "(" + Format(v.L) + " ∨ " + Format(v.R) + ")"
	case ImplF:
		return "(" + Format(v.L) + " → " + Format(v.R) + ")"
	case IffF:
		return "(" + Format(v.L) + " ↔ " + Format(v.R) + ")"
	case GroupF:
		return "(" + Format(v.F) + ")"
	default:
		return f.String()
	}
}
