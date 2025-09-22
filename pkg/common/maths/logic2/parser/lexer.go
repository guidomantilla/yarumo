package parser

import (
	"strings"
	"unicode"
)

type lexer struct {
	s string
	i int
}

func (l *lexer) next() rune {
	if l.i >= len(l.s) {
		return 0
	}
	r, w := rune(l.s[l.i]), 1
	l.i += w
	return r
}

func (l *lexer) peek() rune {
	if l.i >= len(l.s) {
		return 0
	}
	return rune(l.s[l.i])
}

func (l *lexer) skipWS() {
	for unicode.IsSpace(l.peek()) {
		l.next()
	}
}

func (l *lexer) scan() token {
	l.skipWS()
	ch := l.peek()
	pos := l.i
	switch ch {
	case 0:
		return token{typ: tEOF, pos: pos}
	case '!':
		l.next()
		return token{typ: tNOT, lit: "!", pos: pos}
	case '&':
		l.next()
		return token{typ: tAND, lit: "&", pos: pos}
	case '|':
		l.next()
		return token{typ: tOR, lit: "|", pos: pos}
	case '(':
		l.next()
		return token{typ: tLP, lit: "(", pos: pos}
	case ')':
		l.next()
		return token{typ: tRP, lit: ")", pos: pos}
	case '<':
		// expect <=>
		if strings.HasPrefix(l.s[l.i:], "<=>") {
			l.i += 3
			return token{typ: tIFF, lit: "<=>", pos: pos}
		}
		// fallthrough to error
	case '=':
		// expect =>, but our grammar uses only after another '=' from '<=>' or starting with '=' is error
		if strings.HasPrefix(l.s[l.i:], "=>") {
			l.i += 2
			return token{typ: tIMPL, lit: "=>", pos: pos}
		}
		return token{typ: tEOF, pos: pos}
	default:
		if unicode.IsLetter(ch) || ch == '_' {
			start := l.i
			for {
				r := l.peek()
				if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
					l.next()
					continue
				}
				break
			}
			return token{typ: tID, lit: l.s[start:l.i], pos: pos}
		}
	}
	// Unknown char
	l.next()
	return token{typ: tEOF, pos: pos}
}
