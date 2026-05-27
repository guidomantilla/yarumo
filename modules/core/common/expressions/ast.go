package expressions

import (
	"strconv"
	"strings"
)

// NumberLit is a numeric literal.
type NumberLit struct {
	Value float64
}

// Eval returns the numeric value.
func (n *NumberLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return n.Value, nil
}

// String returns the string representation of a NumberLit.
func (n *NumberLit) String() string {
	return strconv.FormatFloat(n.Value, 'g', -1, 64)
}

// StringLit is a string literal.
type StringLit struct {
	Value string
}

// Eval returns the string value.
func (s *StringLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return s.Value, nil
}

// String returns the string representation of a StringLit.
func (s *StringLit) String() string {
	return strconv.Quote(s.Value)
}

// BoolLit is a boolean literal.
type BoolLit struct {
	Value bool
}

// Eval returns the boolean value.
func (b *BoolLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return b.Value, nil
}

// String returns the string representation of a BoolLit.
func (b *BoolLit) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NilLit is a nil literal.
type NilLit struct{}

// Eval returns nil.
func (n *NilLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return nil, nil //nolint:nilnil // nil is the valid result for a nil literal
}

// String returns the string representation of a NilLit.
func (n *NilLit) String() string {
	return "nil"
}

// Ident is an identifier referencing a context variable.
type Ident struct {
	Name string
}

// Eval looks up the identifier in the context.
func (i *Ident) Eval(ctx Context, _ map[string]Func) (any, error) {
	v, ok := ctx[i.Name]
	if !ok {
		return nil, ErrEval("unknown field "+i.Name, ErrUnknownField)
	}
	return v, nil
}

// String returns the string representation of an Ident.
func (i *Ident) String() string {
	return i.Name
}

// Property is a property access expression (e.g. customer.age).
type Property struct {
	Object Expr
	Field  string
}

// Eval navigates the property access chain.
func (p *Property) Eval(ctx Context, funcs map[string]Func) (any, error) {
	obj, err := p.Object.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}
	return resolveProperty(obj, p.Field)
}

// String returns the string representation of a Property.
func (p *Property) String() string {
	return p.Object.String() + "." + p.Field
}

// BinaryOp is a binary operator expression.
type BinaryOp struct {
	Op OpKind
	L  Expr
	R  Expr
}

// Eval evaluates the binary operation.
func (b *BinaryOp) Eval(ctx Context, funcs map[string]Func) (any, error) {
	lv, err := b.L.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	rv, err := b.R.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	return evalBinaryOp(b.Op, lv, rv)
}

// String returns the string representation of a BinaryOp.
func (b *BinaryOp) String() string {
	return "(" + b.L.String() + " " + b.Op.Symbol() + " " + b.R.String() + ")"
}

// UnaryOp is a unary operator expression.
type UnaryOp struct {
	Op OpKind
	X  Expr
}

// Eval evaluates the unary operation.
func (u *UnaryOp) Eval(ctx Context, funcs map[string]Func) (any, error) {
	xv, err := u.X.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	switch u.Op { //nolint:exhaustive // only OpNeg and OpNot are valid unary operators
	case OpNeg:
		n, ok := toFloat64(xv)
		if !ok {
			return nil, ErrEval("unary -: expected numeric, got "+formatValue(xv), ErrTypeMismatch)
		}
		return -n, nil
	case OpNot:
		bv, ok := toBool(xv)
		if !ok {
			return nil, ErrEval("unary !: expected bool, got "+formatValue(xv), ErrTypeMismatch)
		}
		return !bv, nil
	default:
		return nil, ErrEval("unknown unary operator: "+u.Op.Symbol(), ErrTypeMismatch)
	}
}

// String returns the string representation of a UnaryOp.
func (u *UnaryOp) String() string {
	return "(" + u.Op.Symbol() + u.X.String() + ")"
}

// AndExpr is a logical AND expression.
type AndExpr struct {
	L Expr
	R Expr
}

// Eval evaluates the logical AND.
func (a *AndExpr) Eval(ctx Context, funcs map[string]Func) (any, error) {
	lv, err := a.L.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	lb, ok := toBool(lv)
	if !ok {
		return nil, ErrEval("AND: left operand expected bool, got "+formatValue(lv), ErrTypeMismatch)
	}

	if !lb {
		return false, nil
	}

	rv, err := a.R.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	rb, ok := toBool(rv)
	if !ok {
		return nil, ErrEval("AND: right operand expected bool, got "+formatValue(rv), ErrTypeMismatch)
	}

	return rb, nil
}

// String returns the string representation of an AndExpr.
func (a *AndExpr) String() string {
	return "(" + a.L.String() + " AND " + a.R.String() + ")"
}

// OrExpr is a logical OR expression.
type OrExpr struct {
	L Expr
	R Expr
}

// Eval evaluates the logical OR.
func (o *OrExpr) Eval(ctx Context, funcs map[string]Func) (any, error) {
	lv, err := o.L.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	lb, ok := toBool(lv)
	if !ok {
		return nil, ErrEval("OR: left operand expected bool, got "+formatValue(lv), ErrTypeMismatch)
	}

	if lb {
		return true, nil
	}

	rv, err := o.R.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	rb, ok := toBool(rv)
	if !ok {
		return nil, ErrEval("OR: right operand expected bool, got "+formatValue(rv), ErrTypeMismatch)
	}

	return rb, nil
}

// String returns the string representation of an OrExpr.
func (o *OrExpr) String() string {
	return "(" + o.L.String() + " OR " + o.R.String() + ")"
}

// NotExpr is a logical NOT expression.
type NotExpr struct {
	X Expr
}

// Eval evaluates the logical NOT.
func (n *NotExpr) Eval(ctx Context, funcs map[string]Func) (any, error) {
	xv, err := n.X.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	bv, ok := toBool(xv)
	if !ok {
		return nil, ErrEval("NOT: expected bool, got "+formatValue(xv), ErrTypeMismatch)
	}

	return !bv, nil
}

// String returns the string representation of a NotExpr.
func (n *NotExpr) String() string {
	return "(NOT " + n.X.String() + ")"
}

// RangeExpr tests whether a value falls within a range.
type RangeExpr struct {
	X      Expr
	Lo     Expr
	Hi     Expr
	LoIncl bool
	HiIncl bool
}

// Eval evaluates the range membership test.
func (r *RangeExpr) Eval(ctx Context, funcs map[string]Func) (any, error) { //nolint:cyclop // range check needs all three sub-evals plus type checks
	xv, err := r.X.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	lov, err := r.Lo.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	hiv, err := r.Hi.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}

	xn, ok := toFloat64(xv)
	if !ok {
		return nil, ErrEval("IN: subject expected numeric, got "+formatValue(xv), ErrTypeMismatch)
	}

	lon, ok := toFloat64(lov)
	if !ok {
		return nil, ErrEval("IN: lower bound expected numeric, got "+formatValue(lov), ErrTypeMismatch)
	}

	hin, ok := toFloat64(hiv)
	if !ok {
		return nil, ErrEval("IN: upper bound expected numeric, got "+formatValue(hiv), ErrTypeMismatch)
	}

	loOk := (r.LoIncl && xn >= lon) || (!r.LoIncl && xn > lon)
	hiOk := (r.HiIncl && xn <= hin) || (!r.HiIncl && xn < hin)

	return loOk && hiOk, nil
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

// CallExpr is a function call expression.
type CallExpr struct {
	Name string
	Args []Expr
}

// Eval evaluates the function call.
func (c *CallExpr) Eval(ctx Context, funcs map[string]Func) (any, error) {
	fn, ok := funcs[c.Name]
	if !ok {
		return nil, ErrEval("unknown function "+c.Name, ErrUnknownFunc)
	}

	args := make([]any, len(c.Args))
	for i, a := range c.Args {
		v, err := a.Eval(ctx, funcs)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}

	return fn(args...)
}

// String returns the string representation of a CallExpr.
func (c *CallExpr) String() string {
	args := make([]string, len(c.Args))
	for i, a := range c.Args {
		args[i] = a.String()
	}
	return c.Name + "(" + strings.Join(args, ", ") + ")"
}
