package parser

// token types
const (
	tEOF = iota
	tID
	tNOT  // !
	tAND  // &
	tOR   // |
	tIMPL // =>
	tIFF  // <=>
	tLP   // (
	tRP   // )
	tTRUE // TRUE literal
	tFALSE // FALSE literal
)

type token struct {
	typ int
	lit string
	pos int
}
