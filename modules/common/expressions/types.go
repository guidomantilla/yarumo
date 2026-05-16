// Package expressions provides a typed expression evaluator for business rules.
package expressions

// Interface compliance + Fn alias compliance for every public free function.
var (
	_ Evaluator = (*evaluator)(nil)

	_ Expr = (*NumberLit)(nil)
	_ Expr = (*StringLit)(nil)
	_ Expr = (*BoolLit)(nil)
	_ Expr = (*NilLit)(nil)
	_ Expr = (*Ident)(nil)
	_ Expr = (*Property)(nil)
	_ Expr = (*BinaryOp)(nil)
	_ Expr = (*UnaryOp)(nil)
	_ Expr = (*AndExpr)(nil)
	_ Expr = (*OrExpr)(nil)
	_ Expr = (*NotExpr)(nil)
	_ Expr = (*RangeExpr)(nil)
	_ Expr = (*CallExpr)(nil)

	_ ParseFn        = Parse
	_ MustParseFn    = MustParse
	_ DefaultFuncsFn = DefaultFuncs
	_ ErrParseFn     = ErrParse
	_ ErrEvalFn      = ErrEval
)

// Context is the evaluation context: data as a nested map.
type Context map[string]any

// Func is a registered function (built-in or custom).
type Func func(args ...any) (any, error)

// ParseFn is the function type for Parse.
type ParseFn func(input string) (Expr, error)

// MustParseFn is the function type for MustParse.
type MustParseFn func(input string) Expr

// DefaultFuncsFn is the function type for DefaultFuncs.
type DefaultFuncsFn func() map[string]Func

// ErrParseFn is the function type for ErrParse.
type ErrParseFn func(pos, end int, msg string, causes ...error) *ParseError

// ErrEvalFn is the function type for ErrEval.
type ErrEvalFn func(msg string, causes ...error) *EvalError

// Expr is the interface for all AST nodes.
type Expr interface {
	Eval(ctx Context, funcs map[string]Func) (any, error)
	String() string
}

// Evaluator parses and evaluates expression strings against a context.
type Evaluator interface {
	// Evaluate parses an input expression and evaluates it against the given context.
	Evaluate(input string, ctx Context) (any, error)
}

// OpKind represents the kind of operator in binary and unary expressions.
type OpKind int

const (
	OpAdd OpKind = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte
	OpNeg
	OpNot
)

var opSymbols = map[OpKind]string{
	OpAdd: "+",
	OpSub: "-",
	OpMul: "*",
	OpDiv: "/",
	OpMod: "%",
	OpEq:  "==",
	OpNeq: "!=",
	OpLt:  "<",
	OpLte: "<=",
	OpGt:  ">",
	OpGte: ">=",
	OpNeg: "-",
	OpNot: "!",
}

// Symbol returns the string representation of the operator.
func (o OpKind) Symbol() string {
	s, ok := opSymbols[o]
	if !ok {
		return "?"
	}
	return s
}
