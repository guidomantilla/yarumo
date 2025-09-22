package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	s string
	i int // byte index into s (UTF-8 safe)
	strict bool
}

func (l *lexer) next() rune {
	if l.i >= len(l.s) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.s[l.i:])
	l.i += w
	return r
}

func (l *lexer) peek() rune {
	if l.i >= len(l.s) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.s[l.i:])
	return r
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
	// 1) EOF
	if ch == 0 {
		return token{typ: tEOF, pos: pos}
	}

	// 2) Multi-character ASCII operators (longest first)
	if !l.strict {
		if len(l.s)-l.i >= 3 {
			s3 := l.s[l.i : l.i+3]
			switch s3 {
			case "<=>":
				l.i += 3
				return token{typ: tIFF, lit: "<=>", pos: pos}
			case "<->":
				l.i += 3
				return token{typ: tIFF, lit: "<->", pos: pos}
			}
		}
		if len(l.s)-l.i >= 2 {
			s2 := l.s[l.i : l.i+2]
			switch s2 {
			case "=>", "->":
				l.i += 2
				return token{typ: tIMPL, lit: s2, pos: pos}
			case "||":
				l.i += 2
				return token{typ: tOR, lit: s2, pos: pos}
			case "&&":
				l.i += 2
				return token{typ: tAND, lit: s2, pos: pos}
			}
		}
	} else {
		// Strict: only canonical <=> and => are allowed as multi-char
		if len(l.s)-l.i >= 3 && l.s[l.i:l.i+3] == "<=>" {
			l.i += 3
			return token{typ: tIFF, lit: "<=>", pos: pos}
		}
		if len(l.s)-l.i >= 2 && l.s[l.i:l.i+2] == "=>" {
			l.i += 2
			return token{typ: tIMPL, lit: "=>", pos: pos}
		}
	}

	// 3) Single-rune operators (ASCII and Unicode variants)
	switch ch {
	case '!':
		l.next(); return token{typ: tNOT, lit: "!", pos: pos}
	case '&':
		l.next(); return token{typ: tAND, lit: "&", pos: pos}
	case '|':
		l.next(); return token{typ: tOR, lit: "|", pos: pos}
	case '(':
		l.next(); return token{typ: tLP, lit: "(", pos: pos}
	case ')':
		l.next(); return token{typ: tRP, lit: ")", pos: pos}
	}
	if !l.strict {
		switch ch {
		case '¬':
			l.next(); return token{typ: tNOT, lit: "¬", pos: pos}
		case '∧':
			l.next(); return token{typ: tAND, lit: "∧", pos: pos}
		case '∨':
			l.next(); return token{typ: tOR, lit: "∨", pos: pos}
		case '→', '⇒':
			l.next(); return token{typ: tIMPL, lit: string(ch), pos: pos}
		case '↔', '⇔':
			l.next(); return token{typ: tIFF, lit: string(ch), pos: pos}
		}
	}

	// 4) Identifiers / keywords
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
		lit := l.s[start:l.i]
		if !l.strict {
			up := strings.ToUpper(lit)
			switch up {
			case "AND":
				return token{typ: tAND, lit: lit, pos: pos}
			case "OR":
				return token{typ: tOR, lit: lit, pos: pos}
			case "NOT":
				return token{typ: tNOT, lit: lit, pos: pos}
			case "THEN":
				return token{typ: tIMPL, lit: lit, pos: pos}
			case "IFF":
				return token{typ: tIFF, lit: lit, pos: pos}
			case "TRUE":
				return token{typ: tTRUE, lit: lit, pos: pos}
			case "FALSE":
				return token{typ: tFALSE, lit: lit, pos: pos}
			}
		}
		return token{typ: tID, lit: lit, pos: pos}
	}

	// 5) Unknown char: consume and return EOF (parser will handle unexpected token)
	l.next()
	return token{typ: tEOF, pos: pos}
}
