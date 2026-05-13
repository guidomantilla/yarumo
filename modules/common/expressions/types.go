// Package expressions provides a typed expression evaluator for business rules.
package expressions

// Context is the evaluation context: data as a nested map.
type Context map[string]any

// Func is a registered function (built-in or custom).
type Func func(args ...any) (any, error)

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

var _ Evaluator = (*evaluator)(nil)

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
