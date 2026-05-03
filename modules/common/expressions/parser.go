package expressions

import (
	"strconv"
	"strings"
)

// Parse parses an expression string into an AST.
func Parse(input string) (Expr, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, ErrParse(0, 0, "empty input", ErrEmptyInput)
	}

	tokens, lexErr := lex(trimmed)
	if lexErr != nil {
		return nil, lexErr
	}

	p := &parser{tokens: tokens, input: trimmed}

	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.peek().kind != tokEOF {
		tok := p.peek()
		return nil, ErrParse(tok.pos, tok.pos+len(tok.val), "unexpected token: "+tok.val, ErrUnexpectedToken)
	}

	return expr, nil
}

// MustParse parses an expression string and panics on error.
func MustParse(input string) Expr {
	expr, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return expr
}

// parser is a recursive descent parser for expressions.
type parser struct {
	tokens []token
	pos    int
	input  string
}

func (p *parser) peek() token { return p.tokens[p.pos] }

func (p *parser) advance() token { t := p.tokens[p.pos]; p.pos++; return t }

func (p *parser) expect(kind tokenKind) (token, error) {
	tok := p.peek()
	if tok.kind != kind {
		return tok, ErrParse(tok.pos, tok.pos+len(tok.val), "unexpected token: "+tok.val, ErrUnexpectedToken)
	}
	return p.advance(), nil
}

// parseOr: and_expr { OR and_expr }.
func (p *parser) parseOr() (Expr, error) {
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
		left = &OrExpr{L: left, R: right}
	}

	return left, nil
}

// parseAnd: not_expr { AND not_expr }.
func (p *parser) parseAnd() (Expr, error) {
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
		left = &AndExpr{L: left, R: right}
	}

	return left, nil
}

// parseNot: NOT not_expr | comparison.
func (p *parser) parseNot() (Expr, error) {
	if p.peek().kind == tokNot {
		p.advance()
		x, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &NotExpr{X: x}, nil
	}

	return p.parseComparison()
}

// parseComparison: addition [ cmp_op addition | IN range ].
func (p *parser) parseComparison() (Expr, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	switch p.peek().kind { //nolint:exhaustive // only comparison operators and IN are handled
	case tokEq, tokNeq, tokLt, tokLte, tokGt, tokGte:
		op := tokenToOp(p.advance().kind)
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		return &BinaryOp{Op: op, L: left, R: right}, nil

	case tokIn:
		p.advance()
		return p.parseRange(left)
	}

	return left, nil
}

// parseRange: ("["|"(") addition ".." addition ("]"|")").
func (p *parser) parseRange(subject Expr) (Expr, error) {
	loIncl := false
	tok := p.peek()

	switch tok.kind { //nolint:exhaustive // only bracket and paren openers are valid
	case tokLBracket:
		loIncl = true
		p.advance()
	case tokLParen:
		p.advance()
	default:
		return nil, ErrParse(tok.pos, tok.pos+len(tok.val), "expected '[' or '(' for range", ErrUnexpectedToken)
	}

	lo, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	_, err = p.expect(tokDotDot)
	if err != nil {
		return nil, ErrParse(p.peek().pos, p.peek().pos+len(p.peek().val), "expected '..' in range", ErrUnexpectedToken)
	}

	hi, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	hiIncl := false
	tok = p.peek()

	switch tok.kind { //nolint:exhaustive // only bracket and paren closers are valid
	case tokRBracket:
		hiIncl = true
		p.advance()
	case tokRParen:
		p.advance()
	default:
		return nil, ErrParse(tok.pos, tok.pos+len(tok.val), "expected ']' or ')' to close range", ErrUnclosedBracket)
	}

	return &RangeExpr{X: subject, Lo: lo, Hi: hi, LoIncl: loIncl, HiIncl: hiIncl}, nil
}

// parseAddition: multiplication { ("+"|"-") multiplication }.
func (p *parser) parseAddition() (Expr, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokPlus || p.peek().kind == tokMinus {
		op := tokenToOp(p.advance().kind)
		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: op, L: left, R: right}
	}

	return left, nil
}

// parseMultiplication: unary { ("*"|"/"|"%") unary }.
func (p *parser) parseMultiplication() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokStar || p.peek().kind == tokSlash || p.peek().kind == tokPercent {
		op := tokenToOp(p.advance().kind)
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOp{Op: op, L: left, R: right}
	}

	return left, nil
}

// parseUnary: "-" unary | postfix.
func (p *parser) parseUnary() (Expr, error) {
	if p.peek().kind == tokMinus {
		p.advance()
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOp{Op: OpNeg, X: x}, nil
	}

	return p.parsePostfix()
}

// parsePostfix: atom { "." IDENT | "(" [arglist] ")" }.
func (p *parser) parsePostfix() (Expr, error) {
	expr, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for {
		switch p.peek().kind { //nolint:exhaustive // only dot and lparen are postfix operators
		case tokDot:
			p.advance()
			tok, err := p.expect(tokIdent)
			if err != nil {
				return nil, ErrParse(p.peek().pos, p.peek().pos+len(p.peek().val), "expected identifier after '.'", ErrUnexpectedToken)
			}
			expr = &Property{Object: expr, Field: tok.val}

		case tokLParen:
			ident, ok := expr.(*Ident)
			if !ok {
				return expr, nil
			}
			p.advance()
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(tokRParen)
			if err != nil {
				return nil, ErrParse(p.peek().pos, p.peek().pos+len(p.peek().val), "expected ')' after arguments", ErrUnclosedParen)
			}
			expr = &CallExpr{Name: ident.Name, Args: args}

		default:
			return expr, nil
		}
	}
}

// parseArgList: [ expression { "," expression } ].
func (p *parser) parseArgList() ([]Expr, error) {
	if p.peek().kind == tokRParen {
		return nil, nil
	}

	var args []Expr
	first, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	args = append(args, first)

	for p.peek().kind == tokComma {
		p.advance()
		arg, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

// parseAtom: NUMBER | STRING | "true" | "false" | "nil" | IDENT | "(" expression ")".
func (p *parser) parseAtom() (Expr, error) { //nolint:cyclop // atom parsing requires one case per token type
	tok := p.peek()

	switch tok.kind { //nolint:exhaustive // only atom-starting tokens are handled; others fall to default
	case tokNumber:
		p.advance()
		n, err := strconv.ParseFloat(tok.val, 64)
		if err != nil {
			return nil, ErrParse(tok.pos, tok.pos+len(tok.val), "invalid number: "+tok.val, ErrInvalidNumber)
		}
		return &NumberLit{Value: n}, nil

	case tokString:
		p.advance()
		return &StringLit{Value: tok.val}, nil

	case tokTrue:
		p.advance()
		return &BoolLit{Value: true}, nil

	case tokFalse:
		p.advance()
		return &BoolLit{Value: false}, nil

	case tokNil:
		p.advance()
		return &NilLit{}, nil

	case tokIdent:
		p.advance()
		return &Ident{Name: tok.val}, nil

	case tokLParen:
		p.advance()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(tokRParen)
		if err != nil {
			return nil, ErrParse(p.peek().pos, p.peek().pos+len(p.peek().val), "unclosed parenthesis", ErrUnclosedParen)
		}
		return expr, nil

	case tokEOF:
		return nil, ErrParse(tok.pos, tok.pos, "unexpected end of input", ErrUnexpectedEnd)

	default:
		return nil, ErrParse(tok.pos, tok.pos+len(tok.val), "unexpected token: "+tok.val, ErrUnexpectedToken)
	}
}

// tokenToOp converts a token kind to an OpKind.
func tokenToOp(kind tokenKind) OpKind { //nolint:cyclop // mapping function needs one case per token
	switch kind { //nolint:exhaustive // only operator tokens are mapped
	case tokPlus:
		return OpAdd
	case tokMinus:
		return OpSub
	case tokStar:
		return OpMul
	case tokSlash:
		return OpDiv
	case tokPercent:
		return OpMod
	case tokEq:
		return OpEq
	case tokNeq:
		return OpNeq
	case tokLt:
		return OpLt
	case tokLte:
		return OpLte
	case tokGt:
		return OpGt
	case tokGte:
		return OpGte
	default:
		return OpAdd
	}
}
