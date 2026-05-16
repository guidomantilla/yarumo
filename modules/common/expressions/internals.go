package expressions

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// --- coercion helpers ----------------------------------------------------

// toFloat64 attempts to convert a value to float64.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// toBool attempts to convert a value to bool.
func toBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

// toString attempts to convert a value to string.
func toString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// toSlice attempts to convert a value to []any.
func toSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}

// formatValue formats a value for display in error messages.
func formatValue(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprint(v)
}

// resolveProperty navigates nested map[string]any by dot-separated field.
func resolveProperty(obj any, field string) (any, error) {
	if obj == nil {
		return nil, ErrEval("cannot access field "+field+" on nil", ErrNilAccess)
	}

	m, ok := obj.(map[string]any)
	if !ok {
		ctx, isCtx := obj.(Context)
		if !isCtx {
			return nil, ErrEval("cannot access field "+field+" on "+formatValue(obj), ErrTypeMismatch)
		}
		m = map[string]any(ctx)
	}

	parts := strings.SplitN(field, ".", 2)
	val, exists := m[parts[0]]
	if !exists {
		return nil, ErrEval("unknown field "+parts[0], ErrUnknownField)
	}

	if len(parts) == 1 {
		return val, nil
	}

	return resolveProperty(val, parts[1])
}

// --- binary operator dispatch -------------------------------------------

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

// compare evaluates a comparison operator over two values of the same ordered type.
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

// --- builtin functions registry ------------------------------------------

// requireSliceArg validates that a function received exactly one slice argument
// and converts all elements to float64.
func requireSliceArg(name string, args []any) ([]float64, error) {
	if len(args) != 1 {
		return nil, ErrEval(name+": expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toSlice(args[0])
	if !ok {
		return nil, ErrEval(name+": expected slice, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	nums := make([]float64, 0, len(s))

	for i, v := range s {
		n, nok := toFloat64(v)
		if !nok {
			return nil, ErrEval(name+": element "+strconv.Itoa(i)+" is not numeric", ErrTypeMismatch)
		}

		nums = append(nums, n)
	}

	return nums, nil
}

// builtinLen returns the length of a string or []any.
func builtinLen(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("len: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	switch v := args[0].(type) {
	case string:
		return float64(len(v)), nil
	case []any:
		return float64(len(v)), nil
	default:
		return nil, ErrEval("len: unsupported type "+formatValue(args[0]), ErrTypeMismatch)
	}
}

// builtinSum returns the sum of a numeric slice.
func builtinSum(args ...any) (any, error) {
	nums, err := requireSliceArg("sum", args)
	if err != nil {
		return nil, err
	}

	total := 0.0
	for _, n := range nums {
		total += n
	}

	return total, nil
}

// builtinMin returns the minimum value of a non-empty numeric slice.
func builtinMin(args ...any) (any, error) {
	nums, err := requireSliceArg("min", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("min: empty slice", ErrArgCount)
	}

	result := nums[0]

	for _, n := range nums[1:] {
		if n < result {
			result = n
		}
	}

	return result, nil
}

// builtinMax returns the maximum value of a non-empty numeric slice.
func builtinMax(args ...any) (any, error) {
	nums, err := requireSliceArg("max", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("max: empty slice", ErrArgCount)
	}

	result := nums[0]

	for _, n := range nums[1:] {
		if n > result {
			result = n
		}
	}

	return result, nil
}

// builtinAvg returns the arithmetic mean of a non-empty numeric slice.
func builtinAvg(args ...any) (any, error) {
	nums, err := requireSliceArg("avg", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("avg: empty slice", ErrArgCount)
	}

	total := 0.0
	for _, n := range nums {
		total += n
	}

	return total / float64(len(nums)), nil
}

// builtinAbs returns the absolute value of a number.
func builtinAbs(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("abs: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	n, ok := toFloat64(args[0])
	if !ok {
		return nil, ErrEval("abs: expected numeric, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return math.Abs(n), nil
}

// builtinContains reports membership: substring in a string, or element in a []any.
func builtinContains(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, ErrEval("contains: expected 2 arguments, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	switch collection := args[0].(type) {
	case string:
		substr, ok := toString(args[1])
		if !ok {
			return nil, ErrEval("contains: expected string argument, got "+formatValue(args[1]), ErrTypeMismatch)
		}
		return strings.Contains(collection, substr), nil
	case []any:
		return slices.Contains(collection, args[1]), nil
	default:
		return nil, ErrEval("contains: unsupported collection type "+formatValue(args[0]), ErrTypeMismatch)
	}
}

// builtinLower returns the lowercased version of a string.
func builtinLower(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("lower: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toString(args[0])
	if !ok {
		return nil, ErrEval("lower: expected string, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return strings.ToLower(s), nil
}

// builtinUpper returns the uppercased version of a string.
func builtinUpper(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("upper: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toString(args[0])
	if !ok {
		return nil, ErrEval("upper: expected string, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return strings.ToUpper(s), nil
}

// --- lexer / parser entry helpers ----------------------------------------

// lex runs the lexer on input and returns the resulting token stream or a parse error.
func lex(input string) ([]token, *ParseError) {
	l := &lexer{input: input}
	l.run()
	if l.err != nil {
		return nil, l.err
	}
	return l.tokens, nil
}

// tokenToOp maps a binary-operator token kind to its OpKind enum value.
func tokenToOp(kind tokenKind) OpKind { //nolint:cyclop // mapping function needs one case per token
	switch kind { //nolint:exhaustive // only operator tokens are mapped
	case tokPlus:
		return OpAdd
	case tokMinus:
		return OpSub
	case tokStar:
		return OpMul
	case tokSlash:
		return OpDiv
	case tokPercent:
		return OpMod
	case tokEq:
		return OpEq
	case tokNeq:
		return OpNeq
	case tokLt:
		return OpLt
	case tokLte:
		return OpLte
	case tokGt:
		return OpGt
	case tokGte:
		return OpGte
	default:
		return OpAdd
	}
}
