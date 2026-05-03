package expressions

import (
	"testing"
)

func TestLex(t *testing.T) {
	t.Parallel()

	t.Run("empty input returns EOF", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tokens) != 1 || tokens[0].kind != tokEOF {
			t.Fatal("expected single EOF token")
		}
	})

	t.Run("integer number", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNumber || tokens[0].val != "42" {
			t.Fatalf("expected number 42, got %v", tokens[0])
		}
	})

	t.Run("decimal number", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("3.14")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNumber || tokens[0].val != "3.14" {
			t.Fatalf("expected number 3.14, got %v", tokens[0])
		}
	})

	t.Run("double quoted string", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex(`"hello"`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokString || tokens[0].val != "hello" {
			t.Fatalf("expected string hello, got %v", tokens[0])
		}
	})

	t.Run("single quoted string", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex(`'world'`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokString || tokens[0].val != "world" {
			t.Fatalf("expected string world, got %v", tokens[0])
		}
	})

	t.Run("unclosed string returns error", func(t *testing.T) {
		t.Parallel()
		_, err := lex(`"unclosed`)
		if err == nil {
			t.Fatal("expected error for unclosed string")
		}
	})

	t.Run("escape in string", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex(`"he\"llo"`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokString {
			t.Fatalf("expected string token, got %v", tokens[0].kind)
		}
	})

	t.Run("identifier", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("age")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokIdent || tokens[0].val != "age" {
			t.Fatalf("expected ident age, got %v", tokens[0])
		}
	})

	t.Run("identifier with underscore and digits", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("var_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokIdent || tokens[0].val != "var_1" {
			t.Fatalf("expected ident var_1, got %v", tokens[0])
		}
	})

	t.Run("keywords AND OR NOT IN", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("AND OR NOT IN")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []tokenKind{tokAnd, tokOr, tokNot, tokIn, tokEOF}
		if len(tokens) != len(expected) {
			t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
		}
		for i, e := range expected {
			if tokens[i].kind != e {
				t.Fatalf("token %d: expected %d, got %d", i, e, tokens[i].kind)
			}
		}
	})

	t.Run("lowercase keywords", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("and or not in")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []tokenKind{tokAnd, tokOr, tokNot, tokIn, tokEOF}
		for i, e := range expected {
			if tokens[i].kind != e {
				t.Fatalf("token %d: expected %d, got %d", i, e, tokens[i].kind)
			}
		}
	})

	t.Run("true false nil keywords", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("true false nil")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokTrue {
			t.Fatalf("expected tokTrue, got %d", tokens[0].kind)
		}
		if tokens[1].kind != tokFalse {
			t.Fatalf("expected tokFalse, got %d", tokens[1].kind)
		}
		if tokens[2].kind != tokNil {
			t.Fatalf("expected tokNil, got %d", tokens[2].kind)
		}
	})

	t.Run("operators", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("+ - * / % == != < <= > >=")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []tokenKind{
			tokPlus, tokMinus, tokStar, tokSlash, tokPercent,
			tokEq, tokNeq, tokLt, tokLte, tokGt, tokGte, tokEOF,
		}
		if len(tokens) != len(expected) {
			t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
		}
		for i, e := range expected {
			if tokens[i].kind != e {
				t.Fatalf("token %d: expected %d, got %d", i, e, tokens[i].kind)
			}
		}
	})

	t.Run("dot and dotdot", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex(". ..")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokDot {
			t.Fatalf("expected tokDot, got %d", tokens[0].kind)
		}
		if tokens[1].kind != tokDotDot {
			t.Fatalf("expected tokDotDot, got %d", tokens[1].kind)
		}
	})

	t.Run("parentheses brackets comma", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("( ) [ ] ,")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []tokenKind{tokLParen, tokRParen, tokLBracket, tokRBracket, tokComma, tokEOF}
		for i, e := range expected {
			if tokens[i].kind != e {
				t.Fatalf("token %d: expected %d, got %d", i, e, tokens[i].kind)
			}
		}
	})

	t.Run("&& and ||", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("&& ||")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokAnd {
			t.Fatalf("expected tokAnd, got %d", tokens[0].kind)
		}
		if tokens[1].kind != tokOr {
			t.Fatalf("expected tokOr, got %d", tokens[1].kind)
		}
	})

	t.Run("! alone is tokNot", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNot {
			t.Fatalf("expected tokNot, got %d", tokens[0].kind)
		}
	})

	t.Run("unexpected character", func(t *testing.T) {
		t.Parallel()
		_, err := lex("@")
		if err == nil {
			t.Fatal("expected error for @")
		}
	})

	t.Run("single & is error", func(t *testing.T) {
		t.Parallel()
		_, err := lex("& ")
		if err == nil {
			t.Fatal("expected error for single &")
		}
	})

	t.Run("single | is error", func(t *testing.T) {
		t.Parallel()
		_, err := lex("| ")
		if err == nil {
			t.Fatal("expected error for single |")
		}
	})

	t.Run("single = is error", func(t *testing.T) {
		t.Parallel()
		_, err := lex("= ")
		if err == nil {
			t.Fatal("expected error for single =")
		}
	})

	t.Run("complex expression", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("income * 12 > 120000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// income * 12 > 120000 EOF
		if len(tokens) != 6 {
			t.Fatalf("expected 6 tokens, got %d", len(tokens))
		}
		if tokens[0].kind != tokIdent || tokens[0].val != "income" {
			t.Fatalf("expected ident income, got %v", tokens[0])
		}
		if tokens[1].kind != tokStar {
			t.Fatalf("expected star, got %d", tokens[1].kind)
		}
		if tokens[2].kind != tokNumber || tokens[2].val != "12" {
			t.Fatalf("expected number 12, got %v", tokens[2])
		}
	})

	t.Run("number followed by dotdot", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("18..65")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNumber || tokens[0].val != "18" {
			t.Fatalf("expected number 18, got %v", tokens[0])
		}
		if tokens[1].kind != tokDotDot {
			t.Fatalf("expected dotdot, got %d", tokens[1].kind)
		}
		if tokens[2].kind != tokNumber || tokens[2].val != "65" {
			t.Fatalf("expected number 65, got %v", tokens[2])
		}
	})

	t.Run("token positions are correct", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("a + b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].pos != 0 {
			t.Fatalf("expected pos 0, got %d", tokens[0].pos)
		}
		if tokens[1].pos != 2 {
			t.Fatalf("expected pos 2, got %d", tokens[1].pos)
		}
		if tokens[2].pos != 4 {
			t.Fatalf("expected pos 4, got %d", tokens[2].pos)
		}
	})

	t.Run("whitespace is skipped", func(t *testing.T) {
		t.Parallel()
		tokens, err := lex("  42  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNumber || tokens[0].val != "42" {
			t.Fatalf("expected number 42, got %v", tokens[0])
		}
		if tokens[1].kind != tokEOF {
			t.Fatalf("expected EOF, got %d", tokens[1].kind)
		}
	})

	t.Run("number with double dot breaks correctly", func(t *testing.T) {
		t.Parallel()
		// "1.5" should be a single number, but "1.5.6" should be 1.5 . 6
		tokens, err := lex("1.5.name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[0].kind != tokNumber || tokens[0].val != "1.5" {
			t.Fatalf("expected number 1.5, got %v", tokens[0])
		}
		if tokens[1].kind != tokDot {
			t.Fatalf("expected dot, got %d", tokens[1].kind)
		}
		if tokens[2].kind != tokIdent || tokens[2].val != "name" {
			t.Fatalf("expected ident name, got %v", tokens[2])
		}
	})
}
