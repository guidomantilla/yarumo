package parser

// token types
const (
	tEOF = iota
	tID
	tNOT     // !
	tAND     // &
	tOR      // |
	tIMPL    // =>
	tIFF     // <=>
	tLP      // (
	tRP      // )
	tTRUE    // TRUE literal
	tFALSE   // FALSE literal
	tILLEGAL // illegal/unrecognized token
)

type token struct {
	typ int
	lit string
	pos int
}
