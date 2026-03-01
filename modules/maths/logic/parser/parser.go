package parser

import (
	"strings"

	"github.com/guidomantilla/yarumo/maths/logic"
)

// Parse parses a propositional logic formula string into a Formula.
func Parse(input string) (logic.Formula, error) {
	return ParseWith(input)
}

// ParseWith parses a propositional logic formula with the given options.
func ParseWith(input string, opts ...Option) (logic.Formula, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, &ParseError{Pos: 0, Col: 1, Msg: ErrEmptyInput.Error()}
	}

	options := NewOptions(opts...)
	tokens := lex(trimmed)
	p := &parser{tokens: tokens, options: options, input: trimmed}

	f, err := p.parseIff()
	if err != nil {
		return nil, err
	}

	if p.tokens[p.pos].kind != tokEOF {
		tok := p.tokens[p.pos]

		return nil, &ParseError{
			Pos: tok.pos,
			Col: tok.pos + 1,
			Msg: ErrUnexpectedToken.Error() + ": " + tok.val,
		}
	}

	return f, nil
}

// MustParse parses a formula string and panics on error.
func MustParse(input string) logic.Formula {
	f, err := Parse(input)
	if err != nil {
		panic(err)
	}

	return f
}

type parser struct {
	tokens  []token
	pos     int
	options Options
	input   string
}

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) advance() {
	p.pos++
}

// Precedence (lowest to highest): iff < impl < or < and < not < atom.

func (p *parser) parseIff() (logic.Formula, error) {
	left, err := p.parseImpl()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokIff {
		p.advance()

		right, err := p.parseImpl()
		if err != nil {
			return nil, err
		}

		left = logic.IffF{L: left, R: right}
	}

	return left, nil
}

func (p *parser) parseImpl() (logic.Formula, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	// Implication is right-associative.
	if p.peek().kind == tokImpl {
		p.advance()

		right, err := p.parseImpl()
		if err != nil {
			return nil, err
		}

		return logic.ImplF{L: left, R: right}, nil
	}

	return left, nil
}

func (p *parser) parseOr() (logic.Formula, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokOr {
		p.advance()

		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		left = logic.OrF{L: left, R: right}
	}

	return left, nil
}

func (p *parser) parseAnd() (logic.Formula, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokAnd {
		p.advance()

		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}

		left = logic.AndF{L: left, R: right}
	}

	return left, nil
}

func (p *parser) parseNot() (logic.Formula, error) {
	if p.peek().kind == tokNot {
		p.advance()

		f, err := p.parseNot()
		if err != nil {
			return nil, err
		}

		return logic.NotF{F: f}, nil
	}

	return p.parseAtom()
}

func (p *parser) parseAtom() (logic.Formula, error) {
	tok := p.peek()

	switch tok.kind { //nolint:exhaustive // only atom-level tokens handled; operators are caught by default
	case tokVar:
		p.advance()

		return logic.Var(tok.val), nil
	case tokTrue:
		p.advance()

		return logic.TrueF{}, nil
	case tokFalse:
		p.advance()

		return logic.FalseF{}, nil
	case tokLParen:
		return p.parseGroup()
	case tokEOF:
		return nil, &ParseError{
			Pos: tok.pos,
			Col: tok.pos + 1,
			Msg: ErrUnexpectedEnd.Error(),
		}
	default:
		return nil, &ParseError{
			Pos: tok.pos,
			Col: tok.pos + 1,
			Msg: ErrUnexpectedToken.Error() + ": " + tok.val,
		}
	}
}

func (p *parser) parseGroup() (logic.Formula, error) {
	p.advance() // consume '('

	f, err := p.parseIff()
	if err != nil {
		return nil, err
	}

	if p.peek().kind != tokRParen {
		return nil, &ParseError{
			Pos: p.peek().pos,
			Col: p.peek().pos + 1,
			Msg: ErrUnclosedParen.Error(),
		}
	}

	p.advance() // consume ')'

	return logic.GroupF{F: f}, nil
}
