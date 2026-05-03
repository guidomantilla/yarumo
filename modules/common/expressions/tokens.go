package expressions

// tokenKind represents the type of a lexical token.
type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNumber
	tokString
	tokIdent
	tokTrue
	tokFalse
	tokNil
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokPercent
	tokEq
	tokNeq
	tokLt
	tokLte
	tokGt
	tokGte
	tokAnd
	tokOr
	tokNot
	tokIn
	tokDot
	tokDotDot
	tokComma
	tokLParen
	tokRParen
	tokLBracket
	tokRBracket
)

// token represents a single lexical token.
type token struct {
	kind tokenKind
	val  string
	pos  int
}
