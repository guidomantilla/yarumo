package expressions

import (
	"errors"
	"testing"
)

func TestEval(t *testing.T) {
	t.Parallel()

	ctx := Context{
		"age":    30.0,
		"income": 10000.0,
		"active": true,
		"name":   "Alice",
		"customer": map[string]any{
			"age":  25.0,
			"name": "Bob",
			"address": map[string]any{
				"city": "NYC",
			},
		},
		"payments": []any{100.0, 200.0, 300.0},
		"tags":     []any{"vip", "new"},
		"blocked":  false,
		"balance":  -50.0,
	}
	funcs := DefaultFuncs()

	t.Run("number literal", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("42")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 42.0 {
			t.Fatalf("expected 42, got %v", v)
		}
	})

	t.Run("string literal", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"hello"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello" {
			t.Fatalf("expected hello, got %v", v)
		}
	})

	t.Run("bool literal true", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("true")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("nil literal", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("nil")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Fatalf("expected nil, got %v", v)
		}
	})

	t.Run("identifier lookup", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 30.0 {
			t.Fatalf("expected 30, got %v", v)
		}
	})

	t.Run("unknown identifier", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for unknown ident")
		}
		if !errors.Is(err, ErrUnknownField) {
			t.Fatalf("expected ErrUnknownField, got %v", err)
		}
	})

	t.Run("property access", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("customer.age")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 25.0 {
			t.Fatalf("expected 25, got %v", v)
		}
	})

	t.Run("nested property access", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("customer.address.city")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "NYC" {
			t.Fatalf("expected NYC, got %v", v)
		}
	})

	t.Run("addition", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age + 10")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 40.0 {
			t.Fatalf("expected 40, got %v", v)
		}
	})

	t.Run("string concatenation", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"hello" + " world"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello world" {
			t.Fatalf("expected 'hello world', got %v", v)
		}
	})

	t.Run("add type mismatch", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age + name")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for incompatible add")
		}
	})

	t.Run("subtraction", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age - 5")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 25.0 {
			t.Fatalf("expected 25, got %v", v)
		}
	})

	t.Run("multiplication", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("income * 12")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 120000.0 {
			t.Fatalf("expected 120000, got %v", v)
		}
	})

	t.Run("division", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("income / 2")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5000.0 {
			t.Fatalf("expected 5000, got %v", v)
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("income / 0")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for division by zero")
		}
		if !errors.Is(err, ErrDivisionByZero) {
			t.Fatalf("expected ErrDivisionByZero, got %v", err)
		}
	})

	t.Run("modulo", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("7 % 3")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 1.0 {
			t.Fatalf("expected 1, got %v", v)
		}
	})

	t.Run("modulo by zero", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("7 % 0")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for modulo by zero")
		}
	})

	t.Run("arithmetic left operand non-numeric", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("name - 1")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("arithmetic right operand non-numeric", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("1 * name")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("equality true", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age == 30")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("equality false", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age == 25")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("inequality", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age != 25")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("less than", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age < 40")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("less than equal", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age <= 30")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("greater than", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age > 20")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("greater than equal", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age >= 30")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string comparison", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"a" < "b"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("mixed type comparison error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age < name")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for mixed type comparison")
		}
	})

	t.Run("unary negation", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("-age")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != -30.0 {
			t.Fatalf("expected -30, got %v", v)
		}
	})

	t.Run("unary negation type error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("-name")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for non-numeric negation")
		}
	})

	t.Run("AND short circuit false", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("blocked AND active")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("AND both true", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("active AND active")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("AND left not bool", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age AND active")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("AND right not bool", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("active AND age")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("OR short circuit true", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("active OR blocked")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("OR both false", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("blocked OR blocked")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("OR left not bool", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age OR blocked")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("OR right not bool", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("blocked OR age")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("NOT", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("NOT blocked")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("NOT type error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("NOT age")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("range inclusive", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age IN [18..65]")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("range at boundary inclusive", func(t *testing.T) {
		t.Parallel()
		ctx2 := Context{"x": 18.0}
		expr := MustParse("x IN [18..65]")
		v, err := expr.Eval(ctx2, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true for boundary inclusive, got %v", v)
		}
	})

	t.Run("range at boundary exclusive", func(t *testing.T) {
		t.Parallel()
		ctx2 := Context{"x": 18.0}
		expr := MustParse("x IN (18..65)")
		v, err := expr.Eval(ctx2, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false for boundary exclusive, got %v", v)
		}
	})

	t.Run("range out of bounds", func(t *testing.T) {
		t.Parallel()
		ctx2 := Context{"x": 100.0}
		expr := MustParse("x IN [0..50]")
		v, err := expr.Eval(ctx2, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("range subject type error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("name IN [0..10]")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for non-numeric subject")
		}
	})

	t.Run("range lo bound type error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`age IN ["a"..10]`)
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for non-numeric lo bound")
		}
	})

	t.Run("range hi bound type error", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`age IN [0.."z"]`)
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for non-numeric hi bound")
		}
	})

	t.Run("function call sum", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("sum(payments)")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 600.0 {
			t.Fatalf("expected 600, got %v", v)
		}
	})

	t.Run("function call len string", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("len(name)")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("function call len slice", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("len(payments)")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 3.0 {
			t.Fatalf("expected 3, got %v", v)
		}
	})

	t.Run("unknown function", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown()")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for unknown function")
		}
		if !errors.Is(err, ErrUnknownFunc) {
			t.Fatalf("expected ErrUnknownFunc, got %v", err)
		}
	})

	t.Run("complex business rule", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("income * 12 > 100000 AND age IN [18..65] AND active")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("property access error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("customer.nonexistent")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for nonexistent property")
		}
	})

	t.Run("function arg eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("sum(unknown)")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("binary left eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown + 1")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("binary right eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("1 + unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unary eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("-unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("AND left eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown AND true")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("OR left eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown OR false")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("NOT eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("NOT unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("range subject eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown IN [0..10]")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("range lo eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age IN [unknown..10]")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("range hi eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age IN [0..unknown]")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("property access on object eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("unknown.field")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("OR right eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("blocked OR unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("AND right eval error propagates", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("active AND unknown")
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("string comparison lte", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"a" <= "a"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string comparison gt", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"b" > "a"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string comparison gte", func(t *testing.T) {
		t.Parallel()
		expr := MustParse(`"a" >= "a"`)
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("equality mixed types returns false", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age == name")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false for mixed type equality, got %v", v)
		}
	})

	t.Run("inequality mixed types returns true", func(t *testing.T) {
		t.Parallel()
		expr := MustParse("age != name")
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true for mixed type inequality, got %v", v)
		}
	})

	t.Run("unary not via UnaryOp node", func(t *testing.T) {
		t.Parallel()
		expr := &UnaryOp{Op: OpNot, X: &BoolLit{Value: true}}
		v, err := expr.Eval(ctx, funcs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("unary not via UnaryOp type error", func(t *testing.T) {
		t.Parallel()
		expr := &UnaryOp{Op: OpNot, X: &NumberLit{Value: 42}}
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for non-bool unary not")
		}
	})

	t.Run("unknown unary operator returns error", func(t *testing.T) {
		t.Parallel()
		expr := &UnaryOp{Op: OpAdd, X: &NumberLit{Value: 1}}
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for unknown unary op")
		}
	})

	t.Run("unknown binary operator returns error", func(t *testing.T) {
		t.Parallel()
		expr := &BinaryOp{Op: OpNeg, L: &NumberLit{Value: 1}, R: &NumberLit{Value: 2}}
		_, err := expr.Eval(ctx, funcs)
		if err == nil {
			t.Fatal("expected error for unknown binary op")
		}
	})

	t.Run("unknown arithmetic operator returns error", func(t *testing.T) {
		t.Parallel()
		// evalArithmetic is called for OpSub/OpMul/OpDiv/OpMod;
		// pass an unexpected op through a direct call by wrapping.
		result, err := evalArithmetic(OpEq, 1.0, 2.0)
		if err == nil {
			t.Fatalf("expected error, got %v", result)
		}
	})

	t.Run("unknown comparison operator for numbers", func(t *testing.T) {
		t.Parallel()
		_, err := compare(OpAdd, 1.0, 2.0)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unknown comparison operator for strings", func(t *testing.T) {
		t.Parallel()
		_, err := compare(OpAdd, "a", "b")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("number comparison lt", func(t *testing.T) {
		t.Parallel()
		v, err := compare(OpLt, 1.0, 2.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("number comparison gte", func(t *testing.T) {
		t.Parallel()
		v, err := compare(OpGte, 2.0, 2.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string comparison lte direct", func(t *testing.T) {
		t.Parallel()
		v, err := compare(OpLte, "a", "b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string comparison gte direct", func(t *testing.T) {
		t.Parallel()
		v, err := compare(OpGte, "b", "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})
}
