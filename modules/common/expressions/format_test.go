package expressions

import (
	"testing"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	t.Run("NumberLit integer", func(t *testing.T) {
		t.Parallel()
		n := &NumberLit{Value: 42}
		if got := n.String(); got != "42" {
			t.Fatalf("expected 42, got %s", got)
		}
	})

	t.Run("NumberLit decimal", func(t *testing.T) {
		t.Parallel()
		n := &NumberLit{Value: 3.14}
		if got := n.String(); got != "3.14" {
			t.Fatalf("expected 3.14, got %s", got)
		}
	})

	t.Run("StringLit", func(t *testing.T) {
		t.Parallel()
		s := &StringLit{Value: "hello"}
		if got := s.String(); got != `"hello"` {
			t.Fatalf("expected \"hello\", got %s", got)
		}
	})

	t.Run("BoolLit true", func(t *testing.T) {
		t.Parallel()
		b := &BoolLit{Value: true}
		if got := b.String(); got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("BoolLit false", func(t *testing.T) {
		t.Parallel()
		b := &BoolLit{Value: false}
		if got := b.String(); got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("NilLit", func(t *testing.T) {
		t.Parallel()
		n := &NilLit{}
		if got := n.String(); got != "nil" {
			t.Fatalf("expected nil, got %s", got)
		}
	})

	t.Run("Ident", func(t *testing.T) {
		t.Parallel()
		i := &Ident{Name: "age"}
		if got := i.String(); got != "age" {
			t.Fatalf("expected age, got %s", got)
		}
	})

	t.Run("Property", func(t *testing.T) {
		t.Parallel()
		p := &Property{Object: &Ident{Name: "customer"}, Field: "age"}
		if got := p.String(); got != "customer.age" {
			t.Fatalf("expected customer.age, got %s", got)
		}
	})

	t.Run("BinaryOp", func(t *testing.T) {
		t.Parallel()
		b := &BinaryOp{Op: OpMul, L: &Ident{Name: "income"}, R: &NumberLit{Value: 12}}
		if got := b.String(); got != "(income * 12)" {
			t.Fatalf("expected (income * 12), got %s", got)
		}
	})

	t.Run("UnaryOp neg", func(t *testing.T) {
		t.Parallel()
		u := &UnaryOp{Op: OpNeg, X: &Ident{Name: "balance"}}
		if got := u.String(); got != "(-balance)" {
			t.Fatalf("expected (-balance), got %s", got)
		}
	})

	t.Run("AndExpr", func(t *testing.T) {
		t.Parallel()
		a := &AndExpr{L: &Ident{Name: "a"}, R: &Ident{Name: "b"}}
		if got := a.String(); got != "(a AND b)" {
			t.Fatalf("expected (a AND b), got %s", got)
		}
	})

	t.Run("OrExpr", func(t *testing.T) {
		t.Parallel()
		o := &OrExpr{L: &Ident{Name: "a"}, R: &Ident{Name: "b"}}
		if got := o.String(); got != "(a OR b)" {
			t.Fatalf("expected (a OR b), got %s", got)
		}
	})

	t.Run("NotExpr", func(t *testing.T) {
		t.Parallel()
		n := &NotExpr{X: &Ident{Name: "blocked"}}
		if got := n.String(); got != "(NOT blocked)" {
			t.Fatalf("expected (NOT blocked), got %s", got)
		}
	})

	t.Run("RangeExpr inclusive", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "age"}, Lo: &NumberLit{Value: 18}, Hi: &NumberLit{Value: 65},
			LoIncl: true, HiIncl: true,
		}
		if got := r.String(); got != "(age IN [18..65])" {
			t.Fatalf("expected (age IN [18..65]), got %s", got)
		}
	})

	t.Run("RangeExpr exclusive", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "x"}, Lo: &NumberLit{Value: 0}, Hi: &NumberLit{Value: 1},
			LoIncl: false, HiIncl: false,
		}
		if got := r.String(); got != "(x IN (0..1))" {
			t.Fatalf("expected (x IN (0..1)), got %s", got)
		}
	})

	t.Run("RangeExpr mixed", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "x"}, Lo: &NumberLit{Value: 0}, Hi: &NumberLit{Value: 1},
			LoIncl: true, HiIncl: false,
		}
		if got := r.String(); got != "(x IN [0..1))" {
			t.Fatalf("expected (x IN [0..1)), got %s", got)
		}
	})

	t.Run("CallExpr no args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "now", Args: nil}
		if got := c.String(); got != "now()" {
			t.Fatalf("expected now(), got %s", got)
		}
	})

	t.Run("CallExpr with args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "sum", Args: []Expr{&Ident{Name: "payments"}}}
		if got := c.String(); got != "sum(payments)" {
			t.Fatalf("expected sum(payments), got %s", got)
		}
	})

	t.Run("CallExpr multiple args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "contains", Args: []Expr{&Ident{Name: "list"}, &StringLit{Value: "x"}}}
		expected := `contains(list, "x")`
		if got := c.String(); got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})
}
