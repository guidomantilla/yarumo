package props

// FormatOptions controls the optional pretty-printing of formulas.
// String() remains the canonical printer; Format provides customization.
//
// Unicode: when true, uses logical symbols (¬ ∧ ∨ → ↔, ⊤, ⊥). When false, uses
// ASCII tokens (! & | => <=>, TRUE, FALSE).
// Spaces: when true, adds spaces around binary operators for readability.
//
// Note: Parenthesization follows the same conservative scheme as String():
// binary operators are wrapped to preserve precedence unambiguously.
type FormatOptions struct {
	Unicode bool
	Spaces  bool
}

// Format renders the formula according to the provided options.
func Format(f Formula, opts FormatOptions) string {
	op := printer{opts: opts}
	return op.print(f)
}

type printer struct{ opts FormatOptions }

func (p printer) sNot() string   { if p.opts.Unicode { return "¬" } else { return "!" } }
func (p printer) sAnd() string   { if p.opts.Unicode { return "∧" } else { return "&" } }
func (p printer) sOr() string    { if p.opts.Unicode { return "∨" } else { return "|" } }
func (p printer) sImpl() string  { if p.opts.Unicode { return "→" } else { return "=>" } }
func (p printer) sIff() string   { if p.opts.Unicode { return "↔" } else { return "<=>" } }
func (p printer) sTrue() string  { if p.opts.Unicode { return "⊤" } else { return "TRUE" } }
func (p printer) sFalse() string { if p.opts.Unicode { return "⊥" } else { return "FALSE" } }
func (p printer) sp() string     { if p.opts.Spaces { return " " } else { return "" } }

func (p printer) print(f Formula) string {
	switch x := f.(type) {
	case TrueF:
		return p.sTrue()
	case FalseF:
		return p.sFalse()
	case Var:
		return string(x)
	case NotF:
		// Unary NOT binds tight; we do not insert extra parentheses beyond inner Stringing
		return p.sNot() + p.wrapUnary(x.F)
	case AndF:
		return "(" + p.print(x.L) + p.sp() + p.sAnd() + p.sp() + p.print(x.R) + ")"
	case OrF:
		return "(" + p.print(x.L) + p.sp() + p.sOr() + p.sp() + p.print(x.R) + ")"
	case ImplF:
		return "(" + p.print(x.L) + p.sp() + p.sImpl() + p.sp() + p.print(x.R) + ")"
	case IffF:
		return "(" + p.print(x.L) + p.sp() + p.sIff() + p.sp() + p.print(x.R) + ")"
	case GroupF:
		return "(" + p.print(x.Inner) + ")"
	default:
		return f.String()
	}
}

// wrapUnary ensures correct visual grouping for a NOT operand when it is a binary operator.
func (p printer) wrapUnary(f Formula) string {
	switch f.(type) {
	case Var, TrueF, FalseF, NotF, GroupF:
		return p.print(f)
	default:
		// Parenthesize complex expressions under NOT for clarity
		return "(" + p.print(f) + ")"
	}
}
