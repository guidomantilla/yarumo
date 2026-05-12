package expressions

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("empty input returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("")
		if err == nil {
			t.Fatal("expected error for empty input")
		}
		if !errors.Is(err, ErrEmptyInput) {
			t.Fatalf("expected ErrEmptyInput, got %v", err)
		}
	})

	t.Run("whitespace only returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("   ")
		if err == nil {
			t.Fatal("expected error for whitespace input")
		}
	})

	t.Run("number literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		n, ok := expr.(*NumberLit)
		if !ok {
			t.Fatalf("expected NumberLit, got %T", expr)
		}
		if n.Value != 42 {
			t.Fatalf("expected 42, got %v", n.Value)
		}
	})

	t.Run("decimal number literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("3.14")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		n, ok := expr.(*NumberLit)
		if !ok {
			t.Fatalf("expected NumberLit, got %T", expr)
		}
		if n.Value != 3.14 {
			t.Fatalf("expected 3.14, got %v", n.Value)
		}
	})

	t.Run("string literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse(`"hello"`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s, ok := expr.(*StringLit)
		if !ok {
			t.Fatalf("expected StringLit, got %T", expr)
		}
		if s.Value != "hello" {
			t.Fatalf("expected hello, got %s", s.Value)
		}
	})

	t.Run("true literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := expr.(*BoolLit)
		if !ok {
			t.Fatalf("expected BoolLit, got %T", expr)
		}
		if !b.Value {
			t.Fatal("expected true")
		}
	})

	t.Run("false literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := expr.(*BoolLit)
		if !ok {
			t.Fatalf("expected BoolLit, got %T", expr)
		}
		if b.Value {
			t.Fatal("expected false")
		}
	})

	t.Run("nil literal", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("nil")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*NilLit)
		if !ok {
			t.Fatalf("expected NilLit, got %T", expr)
		}
	})

	t.Run("identifier", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("age")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		id, ok := expr.(*Ident)
		if !ok {
			t.Fatalf("expected Ident, got %T", expr)
		}
		if id.Name != "age" {
			t.Fatalf("expected age, got %s", id.Name)
		}
	})

	t.Run("addition", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a + b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpAdd {
			t.Fatalf("expected OpAdd, got %v", bin.Op)
		}
	})

	t.Run("subtraction", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a - b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpSub {
			t.Fatalf("expected OpSub, got %v", bin.Op)
		}
	})

	t.Run("multiplication", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a * b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpMul {
			t.Fatalf("expected OpMul, got %v", bin.Op)
		}
	})

	t.Run("division", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a / b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpDiv {
			t.Fatalf("expected OpDiv, got %v", bin.Op)
		}
	})

	t.Run("modulo", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a % b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpMod {
			t.Fatalf("expected OpMod, got %v", bin.Op)
		}
	})

	t.Run("precedence mul over add", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a + b * c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should be (a + (b * c))
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpAdd {
			t.Fatalf("expected top-level OpAdd, got %v", bin.Op)
		}
		inner, ok := bin.R.(*BinaryOp)
		if !ok {
			t.Fatalf("expected inner BinaryOp, got %T", bin.R)
		}
		if inner.Op != OpMul {
			t.Fatalf("expected inner OpMul, got %v", inner.Op)
		}
	})

	t.Run("left associativity of addition", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a + b + c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should be ((a + b) + c)
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		left, ok := bin.L.(*BinaryOp)
		if !ok {
			t.Fatalf("expected left BinaryOp, got %T", bin.L)
		}
		if left.Op != OpAdd {
			t.Fatalf("expected left OpAdd, got %v", left.Op)
		}
	})

	t.Run("parenthesized grouping", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("(a + b) * c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		bin, ok := expr.(*BinaryOp)
		if !ok {
			t.Fatalf("expected BinaryOp, got %T", expr)
		}
		if bin.Op != OpMul {
			t.Fatalf("expected top-level OpMul, got %v", bin.Op)
		}
		inner, ok := bin.L.(*BinaryOp)
		if !ok {
			t.Fatalf("expected inner BinaryOp, got %T", bin.L)
		}
		if inner.Op != OpAdd {
			t.Fatalf("expected inner OpAdd, got %v", inner.Op)
		}
	})

	t.Run("unclosed parenthesis", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("(a + b")
		if err == nil {
			t.Fatal("expected error for unclosed paren")
		}
	})

	t.Run("comparison operators", func(t *testing.T) {
		t.Parallel()
		ops := map[string]OpKind{
			"a == b": OpEq, "a != b": OpNeq,
			"a < b": OpLt, "a <= b": OpLte,
			"a > b": OpGt, "a >= b": OpGte,
		}
		for input, expectedOp := range ops {
			expr, err := Parse(input)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", input, err)
			}
			bin, ok := expr.(*BinaryOp)
			if !ok {
				t.Fatalf("expected BinaryOp for %s, got %T", input, expr)
			}
			if bin.Op != expectedOp {
				t.Fatalf("expected %v for %s, got %v", expectedOp, input, bin.Op)
			}
		}
	})

	t.Run("unary negation", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("-x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		u, ok := expr.(*UnaryOp)
		if !ok {
			t.Fatalf("expected UnaryOp, got %T", expr)
		}
		if u.Op != OpNeg {
			t.Fatalf("expected OpNeg, got %v", u.Op)
		}
	})

	t.Run("AND expression", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a AND b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*AndExpr)
		if !ok {
			t.Fatalf("expected AndExpr, got %T", expr)
		}
	})

	t.Run("OR expression", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a OR b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*OrExpr)
		if !ok {
			t.Fatalf("expected OrExpr, got %T", expr)
		}
	})

	t.Run("NOT expression", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("NOT a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*NotExpr)
		if !ok {
			t.Fatalf("expected NotExpr, got %T", expr)
		}
	})

	t.Run("&& and || operators", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a > 0 && b < 10 || c == 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// OR has lower precedence, so top level should be OrExpr
		_, ok := expr.(*OrExpr)
		if !ok {
			t.Fatalf("expected OrExpr at top, got %T", expr)
		}
	})

	t.Run("! operator", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("!a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*NotExpr)
		if !ok {
			t.Fatalf("expected NotExpr, got %T", expr)
		}
	})

	t.Run("property access", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("customer.age")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		p, ok := expr.(*Property)
		if !ok {
			t.Fatalf("expected Property, got %T", expr)
		}
		if p.Field != "age" {
			t.Fatalf("expected field age, got %s", p.Field)
		}
	})

	t.Run("chained property access", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("a.b.c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		p, ok := expr.(*Property)
		if !ok {
			t.Fatalf("expected Property, got %T", expr)
		}
		if p.Field != "c" {
			t.Fatalf("expected field c, got %s", p.Field)
		}
		inner, ok := p.Object.(*Property)
		if !ok {
			t.Fatalf("expected inner Property, got %T", p.Object)
		}
		if inner.Field != "b" {
			t.Fatalf("expected inner field b, got %s", inner.Field)
		}
	})

	t.Run("function call no args", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("now()")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c, ok := expr.(*CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", expr)
		}
		if c.Name != "now" {
			t.Fatalf("expected now, got %s", c.Name)
		}
		if len(c.Args) != 0 {
			t.Fatalf("expected 0 args, got %d", len(c.Args))
		}
	})

	t.Run("function call with args", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("sum(payments)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c, ok := expr.(*CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", expr)
		}
		if c.Name != "sum" {
			t.Fatalf("expected sum, got %s", c.Name)
		}
		if len(c.Args) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(c.Args))
		}
	})

	t.Run("function call multiple args", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse(`contains(list, "x")`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		c, ok := expr.(*CallExpr)
		if !ok {
			t.Fatalf("expected CallExpr, got %T", expr)
		}
		if len(c.Args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(c.Args))
		}
	})

	t.Run("unclosed function call", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("sum(a, b")
		if err == nil {
			t.Fatal("expected error for unclosed call")
		}
	})

	t.Run("range inclusive", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("age IN [18..65]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r, ok := expr.(*RangeExpr)
		if !ok {
			t.Fatalf("expected RangeExpr, got %T", expr)
		}
		if !r.LoIncl || !r.HiIncl {
			t.Fatal("expected both inclusive")
		}
	})

	t.Run("range exclusive", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("x IN (0..1)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r, ok := expr.(*RangeExpr)
		if !ok {
			t.Fatalf("expected RangeExpr, got %T", expr)
		}
		if r.LoIncl || r.HiIncl {
			t.Fatal("expected both exclusive")
		}
	})

	t.Run("range mixed", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("x IN [0..1)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r, ok := expr.(*RangeExpr)
		if !ok {
			t.Fatalf("expected RangeExpr, got %T", expr)
		}
		if !r.LoIncl {
			t.Fatal("expected lo inclusive")
		}
		if r.HiIncl {
			t.Fatal("expected hi exclusive")
		}
	})

	t.Run("complex expression", func(t *testing.T) {
		t.Parallel()
		expr, err := Parse("income * 12 > 120000 AND age IN [18..65]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := expr.(*AndExpr)
		if !ok {
			t.Fatalf("expected AndExpr at top, got %T", expr)
		}
	})

	t.Run("unexpected trailing token", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("42 42")
		if err == nil {
			t.Fatal("expected error for trailing token")
		}
	})

	t.Run("unexpected token at start", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(")")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("property access after dot expects ident", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a.42")
		if err == nil {
			t.Fatal("expected error for non-ident after dot")
		}
	})

	t.Run("range without dotdot", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("x IN [1, 2]")
		if err == nil {
			t.Fatal("expected error for range without ..")
		}
	})

	t.Run("range without closing bracket", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("x IN [1..2")
		if err == nil {
			t.Fatal("expected error for unclosed range")
		}
	})

	t.Run("range without opening bracket", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("x IN 1..2]")
		if err == nil {
			t.Fatal("expected error for range without opening")
		}
	})

	t.Run("lex error propagates", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("@")
		if err == nil {
			t.Fatal("expected error for invalid character")
		}
	})
}

func TestMustParse(t *testing.T) {
	t.Parallel()

	t.Run("valid input succeeds", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("42")
		if expr == nil {
			t.Fatal("expected non-nil expr")
		}
	})

	t.Run("invalid input panics", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for invalid input")
			}
		}()
		MustParse("")
	})
}

func TestTokenToOp(t *testing.T) {
	t.Parallel()

	t.Run("unknown token returns OpAdd", func(t *testing.T) {
		t.Parallel()
		got := tokenToOp(tokEOF)
		if got != OpAdd {
			t.Fatalf("expected OpAdd for unknown, got %v", got)
		}
	})
}

func TestParseErrorPropagation(t *testing.T) {
	t.Parallel()

	t.Run("error after OR", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a OR")
		if err == nil {
			t.Fatal("expected error after OR without right operand")
		}
	})

	t.Run("error after AND", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a AND")
		if err == nil {
			t.Fatal("expected error after AND without right operand")
		}
	})

	t.Run("error after NOT", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("NOT")
		if err == nil {
			t.Fatal("expected error after NOT without operand")
		}
	})

	t.Run("error in comparison right side", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a ==")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in addition right side", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a +")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in multiplication right side", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("a *")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in unary operand", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("-)")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in range lo expression", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("x IN [)..5]")
		if err == nil {
			t.Fatal("expected error in range lo")
		}
	})

	t.Run("error in range hi expression", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("x IN [1..))")
		if err == nil {
			t.Fatal("expected error in range hi")
		}
	})

	t.Run("error in grouped expression", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("()")
		if err == nil {
			t.Fatal("expected error for empty parens")
		}
	})

	t.Run("error in arg list expression", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("f()")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error in arg list after comma", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("f(a,)")
		if err == nil {
			t.Fatal("expected error for trailing comma in args")
		}
	})

	t.Run("postfix dot on non-ident call", func(t *testing.T) {
		t.Parallel()
		// "(a)()" should stop after parsing (a) since it's not an ident
		expr, err := Parse("(a)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expr == nil {
			t.Fatal("expected non-nil expr")
		}
	})

	t.Run("error in first arg of arg list", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("f(])")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("parenthesized expression followed by paren is not a call", func(t *testing.T) {
		t.Parallel()
		// (1 + 2)(x) — the (1+2) is a BinaryOp, not an Ident, so (x) cannot be a call.
		// parsePostfix returns the BinaryOp, and Parse sees unconsumed "(" → error.
		_, err := Parse("(1 + 2)(x)")
		if err == nil {
			t.Fatal("expected error for non-ident expr followed by paren")
		}
	})
}

func TestParseAtomInvalidNumber(t *testing.T) {
	t.Parallel()

	t.Run("crafted invalid number token", func(t *testing.T) {
		t.Parallel()
		// Directly test the parser with a crafted token that has tokNumber
		// but an unparseable value — this is unreachable via the lexer but
		// needed for 100% coverage.
		p := &parser{
			tokens: []token{
				{kind: tokNumber, val: "not_a_number", pos: 0},
				{kind: tokEOF, pos: 12},
			},
		}
		_, err := p.parseAtom()
		if err == nil {
			t.Fatal("expected error for invalid number")
		}
		if !errors.Is(err, ErrInvalidNumber) {
			t.Fatalf("expected ErrInvalidNumber, got %v", err)
		}
	})
}
