package parser

import (
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

type parser struct {
	lex *lexer
	cur token
}

func newParserWithOptions(s string, opts ParseOptions) *parser {
	p := &parser{lex: &lexer{s: s, strict: opts.Strict}}
	p.next()

	return p
}

func newParser(s string) *parser { //nolint:unused
	return newParserWithOptions(s, ParseOptions{})
}

func (parser *parser) next() { parser.cur = parser.lex.scan() }

func (parser *parser) expect(t int) (*token, error) {
	if parser.cur.typ != t {
		return nil, newParseError(parser.cur.pos, "unexpected token: expected different token type")
	}

	tok := parser.cur
	parser.next()

	return &tok, nil
}

// Precedence-climbing parser
// precedence: IFF < IMPL < OR < AND < NOT

func (parser *parser) parse() (p.Formula, error) {
	return parser.parseIff()
}

func (parser *parser) parseIff() (p.Formula, error) {
	left, err := parser.parseImpl()
	if err != nil {
		return nil, err
	}

	for parser.cur.typ == tIFF {
		parser.next()

		right, err := parser.parseImpl()
		if err != nil {
			return nil, err
		}

		left = p.IffF{L: left, R: right}
	}

	return left, nil
}

func (parser *parser) parseImpl() (p.Formula, error) {
	left, err := parser.parseOr()
	if err != nil {
		return nil, err
	}

	for parser.cur.typ == tIMPL {
		parser.next()

		right, err := parser.parseOr()
		if err != nil {
			return nil, err
		}

		left = p.ImplF{L: left, R: right}
	}

	return left, nil
}

func (parser *parser) parseOr() (p.Formula, error) {
	left, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}

	for parser.cur.typ == tOR {
		parser.next()

		right, err := parser.parseAnd()
		if err != nil {
			return nil, err
		}

		left = p.OrF{L: left, R: right}
	}

	return left, nil
}

func (parser *parser) parseAnd() (p.Formula, error) {
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}

	for parser.cur.typ == tAND {
		parser.next()

		right, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}

		left = p.AndF{L: left, R: right}
	}

	return left, nil
}

func (parser *parser) parseUnary() (p.Formula, error) {
	switch parser.cur.typ {
	case tNOT:
		parser.next()

		inner, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}

		return p.NotF{F: inner}, nil
	case tLP:
		parser.next()

		inner, err := parser.parseIff()
		if err != nil {
			return nil, err
		}

		if _, err := parser.expect(tRP); err != nil {
			return nil, err
		}

		return p.GroupF{Inner: inner}, nil
	case tID:
		id := parser.cur.lit
		parser.next()

		return p.Var(id), nil
	case tTRUE:
		parser.next()
		return p.TrueF{}, nil
	case tFALSE:
		parser.next()
		return p.FalseF{}, nil
	case tEOF:
		return nil, newParseError(parser.cur.pos, "unexpected EOF")
	default:
		return nil, newParseError(parser.cur.pos, "unexpected token: '"+parser.cur.lit+"'")
	}
}
