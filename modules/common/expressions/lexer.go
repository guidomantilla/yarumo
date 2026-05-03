package expressions

import (
	"unicode"
	"unicode/utf8"
)

// lexer tokenizes an expression string.
type lexer struct {
	input  string
	pos    int
	tokens []token
	err    *ParseError
}

// lex tokenizes the input and returns the token stream plus any error.
func lex(input string) ([]token, *ParseError) {
	l := &lexer{input: input}
	l.run()
	if l.err != nil {
		return nil, l.err
	}
	return l.tokens, nil
}

func (l *lexer) run() { //nolint:cyclop,funlen // lexer dispatch requires one case per token type
	for l.pos < len(l.input) {
		if l.err != nil {
			return
		}

		r, size := utf8.DecodeRuneInString(l.input[l.pos:])

		switch {
		case unicode.IsSpace(r):
			l.pos += size
		case r == '(':
			l.emit(tokLParen, "(", size)
		case r == ')':
			l.emit(tokRParen, ")", size)
		case r == '[':
			l.emit(tokLBracket, "[", size)
		case r == ']':
			l.emit(tokRBracket, "]", size)
		case r == ',':
			l.emit(tokComma, ",", size)
		case r == '+':
			l.emit(tokPlus, "+", size)
		case r == '-':
			l.emit(tokMinus, "-", size)
		case r == '*':
			l.emit(tokStar, "*", size)
		case r == '/':
			l.emit(tokSlash, "/", size)
		case r == '%':
			l.emit(tokPercent, "%", size)
		case r == '.':
			l.lexDot()
		case r == '!':
			l.lexBang()
		case r == '=':
			l.lexEquals()
		case r == '<':
			l.lexLessThan()
		case r == '>':
			l.lexGreaterThan()
		case r == '&':
			l.lexAmpersand()
		case r == '|':
			l.lexPipe()
		case r == '"' || r == '\'':
			l.lexString(r)
		case unicode.IsDigit(r):
			l.lexNumber()
		case unicode.IsLetter(r) || r == '_':
			l.lexWord()
		default:
			l.err = ErrParse(l.pos, l.pos+size, "unexpected character: "+string(r), ErrUnexpectedToken)
		}
	}

	l.tokens = append(l.tokens, token{kind: tokEOF, pos: l.pos})
}

func (l *lexer) emit(kind tokenKind, val string, advance int) {
	l.tokens = append(l.tokens, token{kind: kind, val: val, pos: l.pos})
	l.pos += advance
}

func (l *lexer) lexDot() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '.' {
		l.emit(tokDotDot, "..", 2)
		return
	}
	l.emit(tokDot, ".", 1)
}

func (l *lexer) lexBang() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
		l.emit(tokNeq, "!=", 2)
		return
	}
	l.emit(tokNot, "!", 1)
}

func (l *lexer) lexEquals() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
		l.emit(tokEq, "==", 2)
		return
	}
	l.err = ErrParse(l.pos, l.pos+1, "expected '==' after '='", ErrUnexpectedToken)
}

func (l *lexer) lexLessThan() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
		l.emit(tokLte, "<=", 2)
		return
	}
	l.emit(tokLt, "<", 1)
}

func (l *lexer) lexGreaterThan() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
		l.emit(tokGte, ">=", 2)
		return
	}
	l.emit(tokGt, ">", 1)
}

func (l *lexer) lexAmpersand() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
		l.emit(tokAnd, "&&", 2)
		return
	}
	l.err = ErrParse(l.pos, l.pos+1, "expected '&&' after '&'", ErrUnexpectedToken)
}

func (l *lexer) lexPipe() {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '|' {
		l.emit(tokOr, "||", 2)
		return
	}
	l.err = ErrParse(l.pos, l.pos+1, "expected '||' after '|'", ErrUnexpectedToken)
}

func (l *lexer) lexString(quote rune) {
	start := l.pos
	l.pos++ // skip opening quote

	for l.pos < len(l.input) {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])
		if r == quote {
			val := l.input[start+1 : l.pos]
			l.tokens = append(l.tokens, token{kind: tokString, val: val, pos: start})
			l.pos += size
			return
		}
		if r == '\\' && l.pos+size < len(l.input) {
			l.pos += size // skip backslash
			_, nextSize := utf8.DecodeRuneInString(l.input[l.pos:])
			l.pos += nextSize // skip escaped char
			continue
		}
		l.pos += size
	}

	l.err = ErrParse(start, l.pos, "unclosed string", ErrUnclosedString)
}

func (l *lexer) lexNumber() {
	start := l.pos
	hasDot := false

	for l.pos < len(l.input) {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])
		if r == '.' {
			// Lookahead: ".." is range operator, not decimal.
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '.' {
				break
			}
			if hasDot {
				break
			}
			hasDot = true
			l.pos += size
			continue
		}
		if !unicode.IsDigit(r) {
			break
		}
		l.pos += size
	}

	l.tokens = append(l.tokens, token{kind: tokNumber, val: l.input[start:l.pos], pos: start})
}

var keywords = map[string]tokenKind{
	"AND":   tokAnd,
	"and":   tokAnd,
	"OR":    tokOr,
	"or":    tokOr,
	"NOT":   tokNot,
	"not":   tokNot,
	"IN":    tokIn,
	"in":    tokIn,
	"true":  tokTrue,
	"false": tokFalse,
	"nil":   tokNil,
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
	if !ok {
		kind = tokIdent
	}

	l.tokens = append(l.tokens, token{kind: kind, val: word, pos: start})
}
