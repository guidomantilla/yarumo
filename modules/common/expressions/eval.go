package expressions

// Eval returns the numeric value.
func (n *NumberLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return n.Value, nil
}

// Eval returns the string value.
func (s *StringLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return s.Value, nil
}

// Eval returns the boolean value.
func (b *BoolLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return b.Value, nil
}

// Eval returns nil.
func (n *NilLit) Eval(_ Context, _ map[string]Func) (any, error) {
	return nil, nil //nolint:nilnil // nil is the valid result for a nil literal
}

// Eval looks up the identifier in the context.
func (i *Ident) Eval(ctx Context, _ map[string]Func) (any, error) {
	v, ok := ctx[i.Name]
	if !ok {
		return nil, ErrEval("unknown field "+i.Name, ErrUnknownField)
	}
	return v, nil
}

// Eval navigates the property access chain.
func (p *Property) Eval(ctx Context, funcs map[string]Func) (any, error) {
	obj, err := p.Object.Eval(ctx, funcs)
	if err != nil {
		return nil, err
	}
	return resolveProperty(obj, p.Field)
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

// evalBinaryOp dispatches binary operator evaluation.
func evalBinaryOp(op OpKind, lv, rv any) (any, error) {
	switch op { //nolint:exhaustive // only arithmetic, comparison, and equality operators are valid binary operators
	case OpAdd:
		return evalAdd(lv, rv)
	case OpSub, OpMul, OpDiv, OpMod:
		return evalArithmetic(op, lv, rv)
	case OpEq:
		return lv == rv, nil
	case OpNeq:
		return lv != rv, nil
	case OpLt, OpLte, OpGt, OpGte:
		return evalComparison(op, lv, rv)
	default:
		return nil, ErrEval("unknown operator: "+op.Symbol(), ErrTypeMismatch)
	}
}

// evalAdd handles + for numbers (addition) and strings (concatenation).
func evalAdd(lv, rv any) (any, error) {
	ln, lok := toFloat64(lv)
	rn, rok := toFloat64(rv)
	if lok && rok {
		return ln + rn, nil
	}

	ls, lok := toString(lv)
	rs, rok := toString(rv)
	if lok && rok {
		return ls + rs, nil
	}

	return nil, ErrEval("+: incompatible types "+formatValue(lv)+" and "+formatValue(rv), ErrTypeMismatch)
}

// evalArithmetic handles -, *, /, %.
func evalArithmetic(op OpKind, lv, rv any) (any, error) {
	ln, lok := toFloat64(lv)
	if !lok {
		return nil, ErrEval(op.Symbol()+": left operand expected numeric, got "+formatValue(lv), ErrTypeMismatch)
	}

	rn, rok := toFloat64(rv)
	if !rok {
		return nil, ErrEval(op.Symbol()+": right operand expected numeric, got "+formatValue(rv), ErrTypeMismatch)
	}

	switch op { //nolint:exhaustive // only Sub, Mul, Div, Mod reach this function
	case OpSub:
		return ln - rn, nil
	case OpMul:
		return ln * rn, nil
	case OpDiv:
		if rn == 0 {
			return nil, ErrEval("division by zero", ErrDivisionByZero)
		}
		return ln / rn, nil
	case OpMod:
		if rn == 0 {
			return nil, ErrEval("division by zero", ErrDivisionByZero)
		}
		return float64(int64(ln) % int64(rn)), nil
	default:
		return nil, ErrEval("unknown arithmetic operator: "+op.Symbol(), ErrTypeMismatch)
	}
}

// evalComparison handles <, <=, >, >=.
func evalComparison(op OpKind, lv, rv any) (any, error) {
	ln, lok := toFloat64(lv)
	rn, rok := toFloat64(rv)
	if lok && rok {
		return compare(op, ln, rn)
	}

	ls, lok := toString(lv)
	rs, rok := toString(rv)
	if lok && rok {
		return compare(op, ls, rs)
	}

	return nil, ErrEval(op.Symbol()+": incompatible types "+formatValue(lv)+" and "+formatValue(rv), ErrTypeMismatch)
}

func compare[T ~float64 | ~string](op OpKind, l, r T) (any, error) {
	switch op { //nolint:exhaustive // only comparison operators reach this function
	case OpLt:
		return l < r, nil
	case OpLte:
		return l <= r, nil
	case OpGt:
		return l > r, nil
	case OpGte:
		return l >= r, nil
	default:
		return nil, ErrEval("unsupported comparison op: "+op.Symbol(), ErrTypeMismatch)
	}
}
