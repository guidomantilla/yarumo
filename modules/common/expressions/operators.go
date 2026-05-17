package expressions

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
