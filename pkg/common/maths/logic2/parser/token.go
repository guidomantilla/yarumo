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
)

type token struct {
	typ int
	lit string
	pos int
}
