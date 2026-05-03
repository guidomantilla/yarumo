package expressions

import (
	"strconv"
	"strings"
)

// String returns the string representation of a NumberLit.
func (n *NumberLit) String() string {
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}

// String returns the string representation of a StringLit.
func (s *StringLit) String() string {
	return strconv.Quote(s.Value)
}

// String returns the string representation of a BoolLit.
func (b *BoolLit) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// String returns the string representation of a NilLit.
func (n *NilLit) String() string {
	return "nil"
}

// String returns the string representation of an Ident.
func (i *Ident) String() string {
	return i.Name
}

// String returns the string representation of a Property.
func (p *Property) String() string {
	return p.Object.String() + "." + p.Field
}

// String returns the string representation of a BinaryOp.
func (b *BinaryOp) String() string {
	return "(" + b.L.String() + " " + b.Op.Symbol() + " " + b.R.String() + ")"
}

// String returns the string representation of a UnaryOp.
func (u *UnaryOp) String() string {
	return "(" + u.Op.Symbol() + u.X.String() + ")"
}

// String returns the string representation of an AndExpr.
func (a *AndExpr) String() string {
	return "(" + a.L.String() + " AND " + a.R.String() + ")"
}

// String returns the string representation of an OrExpr.
func (o *OrExpr) String() string {
	return "(" + o.L.String() + " OR " + o.R.String() + ")"
}

// String returns the string representation of a NotExpr.
func (n *NotExpr) String() string {
	return "(NOT " + n.X.String() + ")"
}

// String returns the string representation of a RangeExpr.
func (r *RangeExpr) String() string {
	lo := "("
	if r.LoIncl {
		lo = "["
	}
	hi := ")"
	if r.HiIncl {
		hi = "]"
	}
	return "(" + r.X.String() + " IN " + lo + r.Lo.String() + ".." + r.Hi.String() + hi + ")"
}

// String returns the string representation of a CallExpr.
func (c *CallExpr) String() string {
	args := make([]string, len(c.Args))
	for i, a := range c.Args {
		args[i] = a.String()
	}
	return c.Name + "(" + strings.Join(args, ", ") + ")"
}
