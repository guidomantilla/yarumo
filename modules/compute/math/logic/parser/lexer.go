package parser

import (
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	input  string
	pos    int
	tokens []token
}

func lex(input string) []token {
	l := &lexer{input: input}
	l.run()

	return l.tokens
}

func (l *lexer) run() {
	for l.pos < len(l.input) {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])

		switch {
		case unicode.IsSpace(r):
			l.pos += size
		case r == '(':
			l.emit(tokLParen, "(", size)
		case r == ')':
			l.emit(tokRParen, ")", size)
		case r == '!' || r == '~':
			l.emit(tokNot, string(r), size)
		case r == '¬':
			l.emit(tokNot, "¬", size)
		case r == '∧':
			l.emit(tokAnd, "∧", size)
		case r == '∨':
			l.emit(tokOr, "∨", size)
		case r == '→':
			l.emit(tokImpl, "→", size)
		case r == '↔':
			l.emit(tokIff, "↔", size)
		case r == '⊤':
			l.emit(tokTrue, "⊤", size)
		case r == '⊥':
			l.emit(tokFalse, "⊥", size)
		case r == '&':
			l.lexAmpersand()
		case r == '|':
			l.lexPipe()
		case r == '=':
			l.lexEquals()
		case r == '<':
			l.lexLessThan()
		case r == '-':
			l.lexDash()
		case r == '^':
			l.emit(tokAnd, "^", size)
		case unicode.IsLetter(r):
			l.lexWord()
		default:
			l.emit(tokVar, string(r), size)
		}
	}

	l.tokens = append(l.tokens, token{kind: tokEOF, pos: l.pos})
}

func (l *lexer) emit(kind tokenKind, val string, advance int) {
	l.tokens = append(l.tokens, token{kind: kind, val: val, pos: l.pos})
	l.pos += advance
}

func (l *lexer) lexAmpersand() {
	start := l.pos
	l.pos++

	if l.pos < len(l.input) && l.input[l.pos] == '&' {
		l.pos++
	}

	l.tokens = append(l.tokens, token{kind: tokAnd, val: l.input[start:l.pos], pos: start})
}

func (l *lexer) lexPipe() {
	start := l.pos
	l.pos++

	if l.pos < len(l.input) && l.input[l.pos] == '|' {
		l.pos++
	}

	l.tokens = append(l.tokens, token{kind: tokOr, val: l.input[start:l.pos], pos: start})
}

func (l *lexer) lexEquals() {
	start := l.pos
	l.pos++

	if l.pos < len(l.input) && l.input[l.pos] == '>' {
		l.pos++
		l.tokens = append(l.tokens, token{kind: tokImpl, val: "=>", pos: start})

		return
	}

	l.tokens = append(l.tokens, token{kind: tokVar, val: "=", pos: start})
}

func (l *lexer) lexLessThan() {
	start := l.pos
	l.pos++

	if l.pos+1 < len(l.input) && l.input[l.pos] == '=' && l.input[l.pos+1] == '>' {
		l.pos += 2
		l.tokens = append(l.tokens, token{kind: tokIff, val: "<=>", pos: start})

		return
	}

	if l.pos+1 < len(l.input) && l.input[l.pos] == '-' && l.input[l.pos+1] == '>' {
		l.pos += 2
		l.tokens = append(l.tokens, token{kind: tokIff, val: "<->", pos: start})

		return
	}

	l.tokens = append(l.tokens, token{kind: tokVar, val: "<", pos: start})
}

func (l *lexer) lexDash() {
	start := l.pos
	l.pos++

	if l.pos < len(l.input) && l.input[l.pos] == '>' {
		l.pos++
		l.tokens = append(l.tokens, token{kind: tokImpl, val: "->", pos: start})

		return
	}

	l.tokens = append(l.tokens, token{kind: tokVar, val: "-", pos: start})
}

var keywords = map[string]tokenKind{ //nolint:gochecknoglobals // keyword lookup table
	"true":  tokTrue,
	"TRUE":  tokTrue,
	"T":     tokTrue,
	"false": tokFalse,
	"FALSE": tokFalse,
	"F":     tokFalse,
	"not":   tokNot,
	"NOT":   tokNot,
	"and":   tokAnd,
	"AND":   tokAnd,
	"or":    tokOr,
	"OR":    tokOr,
	"v":     tokOr,
}

func (l *lexer) lexWord() {
	start := l.pos

	for l.pos < len(l.input) {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}

		l.pos += size
	}

	word := l.input[start:l.pos]

	kind, ok := keywords[word]
	if ok {
		l.tokens = append(l.tokens, token{kind: kind, val: word, pos: start})

		return
	}

	l.tokens = append(l.tokens, token{kind: tokVar, val: word, pos: start})
}
