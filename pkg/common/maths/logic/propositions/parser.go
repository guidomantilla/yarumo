package propositions

import (
	"errors"
	"fmt"
	"strings"
)

func ParseFormula(input string) (Formula, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	return p.parse()
}

//

type tokenType int

const (
	tokVar tokenType = iota
	tokAnd
	tokOr
	tokNot
	tokLParen
	tokRParen
	tokImpl
	tokIff
	tokTrue
	tokFalse
)

type token struct {
	typ tokenType
	val string
}

func tokenize(input string) ([]token, error) {
	replacements := []struct{ old, new string }{

		{"(", " ( "},
		{")", " ) "},

		{"IFF", " IFF "},
		{"iff", " IFF "},
		{"↔", " IFF "},
		{"⇔", " IFF "},
		{"<->", " IFF "},
		{"<=>", " IFF "},

		{"THEN", " THEN "},
		{"then", " THEN "},
		{"⇒", " THEN "},
		{"→", " THEN "},
		{"->", " THEN "},
		{"=>", " THEN "},

		{"OR", " OR "},
		{"or", " OR "},
		{"||", " OR "},
		{"|", " OR "},
		{"∨", " OR "},

		{"AND", " AND "},
		{"and", " AND "},
		{"&&", " AND "},
		{"&", " AND "},
		{"∧", " AND "},

		{"NOT", " NOT "},
		{"not", " NOT "},
		{"!", " NOT "},
		{"¬", " NOT "},

		{"TRUE", " TRUE "},
		{"true", " TRUE "},

		{"FALSE", " FALSE "},
		{"false", " FALSE "},
	}

	for _, r := range replacements {
		input = strings.ReplaceAll(input, r.old, r.new)
	}

	fields := strings.Fields(input)
	var tokens []token
	for _, f := range fields {
		switch f {
		case "AND", "and", "&&", "&", "∧":
			tokens = append(tokens, token{typ: tokAnd})
		case "OR", "or", "||", "|", "∨":
			tokens = append(tokens, token{typ: tokOr})
		case "NOT", "not", "!", "¬":
			tokens = append(tokens, token{typ: tokNot})
		case "(":
			tokens = append(tokens, token{typ: tokLParen})
		case ")":
			tokens = append(tokens, token{typ: tokRParen})
		case "THEN", "then", "→", "⇒", "->", "=>":
			tokens = append(tokens, token{typ: tokImpl})
		case "IFF", "iff", "↔", "⇔", "<->", "<=>":
			tokens = append(tokens, token{typ: tokIff})
		case "TRUE", "true":
			tokens = append(tokens, token{typ: tokTrue})
		case "FALSE", "false":
			tokens = append(tokens, token{typ: tokFalse})
		default:
			tokens = append(tokens, token{typ: tokVar, val: f})
		}
	}
	return tokens, nil
}

//

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) next() token {
	if p.pos >= len(p.tokens) {
		return token{}
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{}
	}
	return p.tokens[p.pos]
}

func (p *parser) parse() (Formula, error) {
	return p.parseIff() // IFF > THEN > OR > AND > NOT
}

func (p *parser) parseIff() (Formula, error) {
	left, err := p.parseImpl()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokIff {
		p.next()
		right, err := p.parseIff()
		if err != nil {
			return nil, err
		}
		left = IffF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseImpl() (Formula, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokImpl {
		p.next()
		right, err := p.parseImpl()
		if err != nil {
			return nil, err
		}
		left = ImplF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseOr() (Formula, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokOr {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = OrF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Formula, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokAnd {
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = AndF{L: left, R: right}
	}
	return left, nil
}

func (p *parser) parseUnary() (Formula, error) {
	tok := p.peek()
	switch tok.typ {
	case tokNot:
		p.next()
		sub, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return NotF{F: sub}, nil
	case tokLParen:
		p.next()
		sub, err := p.parseIff()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tokRParen {
			return nil, errors.New("missing closing parenthesis")
		}
		p.next()
		return sub, nil
	case tokVar:
		p.next()
		return Var(tok.val), nil
	case tokTrue:
		p.next()
		return TrueF{}, nil
	case tokFalse:
		p.next()
		return FalseF{}, nil
	default:
		return nil, fmt.Errorf("unexpected token: %+v", tok)
	}
}
