package parser

// tokenKind identifies the type of a lexer token.
type tokenKind int

// Token kinds produced by the lexer.
const (
	tokEOF    tokenKind = iota
	tokVar              // variable identifier
	tokTrue             // true, T, ⊤
	tokFalse            // false, F, ⊥
	tokNot              // !, ~, ¬, not, NOT
	tokAnd              // &, &&, ∧, and, AND, ^
	tokOr               // |, ||, ∨, or, OR, v
	tokImpl             // =>, ->, →
	tokIff              // <=>, <->, ↔
	tokLParen           // (
	tokRParen           // )
)

// token represents a single lexer token.
type token struct {
	kind tokenKind
	val  string
	pos  int
}
