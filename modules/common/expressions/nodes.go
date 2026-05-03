package expressions

// Type compliance checks.
var (
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
)

// NumberLit is a numeric literal.
type NumberLit struct {
	Value float64
}

// StringLit is a string literal.
type StringLit struct {
	Value string
}

// BoolLit is a boolean literal.
type BoolLit struct {
	Value bool
}

// NilLit is a nil literal.
type NilLit struct{}

// Ident is an identifier referencing a context variable.
type Ident struct {
	Name string
}

// Property is a property access expression (e.g. customer.age).
type Property struct {
	Object Expr
	Field  string
}

// BinaryOp is a binary operator expression.
type BinaryOp struct {
	Op OpKind
	L  Expr
	R  Expr
}

// UnaryOp is a unary operator expression.
type UnaryOp struct {
	Op OpKind
	X  Expr
}

// AndExpr is a logical AND expression.
type AndExpr struct {
	L Expr
	R Expr
}

// OrExpr is a logical OR expression.
type OrExpr struct {
	L Expr
	R Expr
}

// NotExpr is a logical NOT expression.
type NotExpr struct {
	X Expr
}

// RangeExpr tests whether a value falls within a range.
type RangeExpr struct {
	X      Expr
	Lo     Expr
	Hi     Expr
	LoIncl bool
	HiIncl bool
}

// CallExpr is a function call expression.
type CallExpr struct {
	Name string
	Args []Expr
}
