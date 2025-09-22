package parser

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
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

func newParser(s string) *parser { return newParserWithOptions(s, ParseOptions{}) }

func (p *parser) next() { p.cur = p.lex.scan() }

func (p *parser) expect(t int) (*token, error) {
	if p.cur.typ != t {
		return nil, newParseError(p.cur.pos, "unexpected token: expected different token type")
	}
	tok := p.cur
	p.next()
	return &tok, nil
}

// Precedence-climbing parser
// precedence: IFF < IMPL < OR < AND < NOT

func (p *parser) parse() (props.Formula, error) {
	return p.parseIff()
}

func (p *parser) parseIff() (props.Formula, error) {
	left, err := p.parseImpl()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tIFF {
		p.next()
		right, err := p.parseImpl()
		if err != nil {
			return nil, err
		}
		left = props.IffF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseImpl() (props.Formula, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tIMPL {
		p.next()
		right, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		left = props.ImplF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseOr() (props.Formula, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tOR {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = props.OrF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (props.Formula, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.typ == tAND {
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = props.AndF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseUnary() (props.Formula, error) {
	switch p.cur.typ {
	case tNOT:
		p.next()
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return props.NotF{F: inner}, nil
	case tLP:
		p.next()
		inner, err := p.parseIff()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tRP); err != nil {
			return nil, err
		}
		return props.GroupF{Inner: inner}, nil
	case tID:
		id := p.cur.lit
		p.next()
		return props.Var(id), nil
	case tTRUE:
		p.next()
		return props.TrueF{}, nil
	case tFALSE:
		p.next()
		return props.FalseF{}, nil
	case tEOF:
		return nil, newParseError(p.cur.pos, "unexpected EOF")
	default:
		return nil, newParseError(p.cur.pos, "unexpected token: '"+p.cur.lit+"'")
	}
}
