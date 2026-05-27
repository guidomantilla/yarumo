package expressions

import (
	"errors"
	"testing"
)

func TestToFloat64(t *testing.T) {
	t.Parallel()

	t.Run("float64 passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(3.14)
		if !ok || v != 3.14 {
			t.Fatalf("expected 3.14, got %v", v)
		}
	})

	t.Run("int converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(42)
		if !ok || v != 42.0 {
			t.Fatalf("expected 42.0, got %v", v)
		}
	})

	t.Run("int64 converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(int64(100))
		if !ok || v != 100.0 {
			t.Fatalf("expected 100.0, got %v", v)
		}
	})

	t.Run("float32 converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(float32(1.5))
		if !ok {
			t.Fatal("expected ok")
		}
		if v < 1.4 || v > 1.6 {
			t.Fatalf("expected ~1.5, got %v", v)
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toFloat64("hello")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})

	t.Run("nil fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toFloat64(nil)
		if ok {
			t.Fatal("expected not ok for nil")
		}
	})
}

func TestToBool(t *testing.T) {
	t.Parallel()

	t.Run("true passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toBool(true)
		if !ok || !v {
			t.Fatal("expected true")
		}
	})

	t.Run("false passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toBool(false)
		if !ok || v {
			t.Fatal("expected false")
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toBool("true")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})
}

func TestToString(t *testing.T) {
	t.Parallel()

	t.Run("string passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toString("hello")
		if !ok || v != "hello" {
			t.Fatalf("expected hello, got %s", v)
		}
	})

	t.Run("int fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toString(42)
		if ok {
			t.Fatal("expected not ok for int")
		}
	})
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	t.Run("slice passthrough", func(t *testing.T) {
		t.Parallel()
		input := []any{1.0, 2.0}
		v, ok := toSlice(input)
		if !ok || len(v) != 2 {
			t.Fatalf("expected slice of 2, got %v", v)
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toSlice("abc")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil string", func(t *testing.T) {
		t.Parallel()
		got := formatValue(nil)
		if got != "nil" {
			t.Fatalf("expected nil, got %s", got)
		}
	})

	t.Run("number returns formatted", func(t *testing.T) {
		t.Parallel()
		got := formatValue(42.5)
		if got != "42.5" {
			t.Fatalf("expected 42.5, got %s", got)
		}
	})
}

func TestResolveProperty(t *testing.T) {
	t.Parallel()

	t.Run("simple field access", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{"name": "Alice"}
		v, err := resolveProperty(obj, "name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "Alice" {
			t.Fatalf("expected Alice, got %v", v)
		}
	})

	t.Run("nested field access", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{
			"customer": map[string]any{
				"age": 30.0,
			},
		}
		v, err := resolveProperty(obj, "customer.age")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 30.0 {
			t.Fatalf("expected 30, got %v", v)
		}
	})

	t.Run("nil object returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveProperty(nil, "field")
		if err == nil {
			t.Fatal("expected error for nil object")
		}
		if !errors.Is(err, ErrNilAccess) {
			t.Fatalf("expected ErrNilAccess, got %v", err)
		}
	})

	t.Run("non-map object returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveProperty(42, "field")
		if err == nil {
			t.Fatal("expected error for non-map object")
		}
		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatalf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("unknown field returns error", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{"name": "Alice"}
		_, err := resolveProperty(obj, "age")
		if err == nil {
			t.Fatal("expected error for unknown field")
		}
		if !errors.Is(err, ErrUnknownField) {
			t.Fatalf("expected ErrUnknownField, got %v", err)
		}
	})

	t.Run("Context type works", func(t *testing.T) {
		t.Parallel()
		ctx := Context{"x": 10.0}
		v, err := resolveProperty(ctx, "x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 10.0 {
			t.Fatalf("expected 10.0, got %v", v)
		}
	})
}

func TestBuiltinLen(t *testing.T) {
	t.Parallel()

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLen("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("slice length", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLen([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 3.0 {
			t.Fatalf("expected 3, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLen()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrArgCount) {
			t.Fatalf("expected ErrArgCount, got %v", err)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLen(42)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatalf("expected ErrTypeMismatch, got %v", err)
		}
	})
}

func TestBuiltinSum(t *testing.T) {
	t.Parallel()

	t.Run("sums numeric slice", func(t *testing.T) {
		t.Parallel()
		v, err := builtinSum([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 6.0 {
			t.Fatalf("expected 6, got %v", v)
		}
	})

	t.Run("empty slice returns 0", func(t *testing.T) {
		t.Parallel()
		v, err := builtinSum([]any{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 0.0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinMin(t *testing.T) {
	t.Parallel()

	t.Run("finds minimum", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMin([]any{3.0, 1.0, 2.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 1.0 {
			t.Fatalf("expected 1, got %v", v)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMin([]any{5.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric first element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{"bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric later element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinMax(t *testing.T) {
	t.Parallel()

	t.Run("finds maximum", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMax([]any{1.0, 3.0, 2.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 3.0 {
			t.Fatalf("expected 3, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric first element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{"bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric later element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinAvg(t *testing.T) {
	t.Parallel()

	t.Run("computes average", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAvg([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 2.0 {
			t.Fatalf("expected 2, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinAbs(t *testing.T) {
	t.Parallel()

	t.Run("positive number", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAbs(5.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("negative number", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAbs(-5.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAbs()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAbs("bad")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinContains(t *testing.T) {
	t.Parallel()

	t.Run("string contains substring", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains("hello world", "world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string does not contain", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains("hello", "world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("slice contains element", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains([]any{"a", "b"}, "b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("slice does not contain", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains([]any{"a", "b"}, "c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains("a")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unsupported collection type", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains(42, "x")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("string contains with non-string search", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains("hello", 42)
		if err == nil {
			t.Fatal("expected error for non-string search in string")
		}
	})
}

func TestBuiltinLower(t *testing.T) {
	t.Parallel()

	t.Run("converts to lowercase", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLower("HELLO")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello" {
			t.Fatalf("expected hello, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLower()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-string", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLower(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinUpper(t *testing.T) {
	t.Parallel()

	t.Run("converts to uppercase", func(t *testing.T) {
		t.Parallel()
		v, err := builtinUpper("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "HELLO" {
			t.Fatalf("expected HELLO, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinUpper()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-string", func(t *testing.T) {
		t.Parallel()
		_, err := builtinUpper(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

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
