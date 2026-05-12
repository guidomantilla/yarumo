package expressions

import (
	"testing"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	t.Run("NumberLit integer", func(t *testing.T) {
		t.Parallel()
		n := &NumberLit{Value: 42}
		got := n.String()
		if got != "42" {
			t.Fatalf("expected 42, got %s", got)
		}
	})

	t.Run("NumberLit decimal", func(t *testing.T) {
		t.Parallel()
		n := &NumberLit{Value: 3.14}
		got := n.String()
		if got != "3.14" {
			t.Fatalf("expected 3.14, got %s", got)
		}
	})

	t.Run("StringLit", func(t *testing.T) {
		t.Parallel()
		s := &StringLit{Value: "hello"}
		got := s.String()
		if got != `"hello"` {
			t.Fatalf("expected \"hello\", got %s", got)
		}
	})

	t.Run("BoolLit true", func(t *testing.T) {
		t.Parallel()
		b := &BoolLit{Value: true}
		got := b.String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("BoolLit false", func(t *testing.T) {
		t.Parallel()
		b := &BoolLit{Value: false}
		got := b.String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("NilLit", func(t *testing.T) {
		t.Parallel()
		n := &NilLit{}
		got := n.String()
		if got != "nil" {
			t.Fatalf("expected nil, got %s", got)
		}
	})

	t.Run("Ident", func(t *testing.T) {
		t.Parallel()
		i := &Ident{Name: "age"}
		got := i.String()
		if got != "age" {
			t.Fatalf("expected age, got %s", got)
		}
	})

	t.Run("Property", func(t *testing.T) {
		t.Parallel()
		p := &Property{Object: &Ident{Name: "customer"}, Field: "age"}
		got := p.String()
		if got != "customer.age" {
			t.Fatalf("expected customer.age, got %s", got)
		}
	})

	t.Run("BinaryOp", func(t *testing.T) {
		t.Parallel()
		b := &BinaryOp{Op: OpMul, L: &Ident{Name: "income"}, R: &NumberLit{Value: 12}}
		got := b.String()
		if got != "(income * 12)" {
			t.Fatalf("expected (income * 12), got %s", got)
		}
	})

	t.Run("UnaryOp neg", func(t *testing.T) {
		t.Parallel()
		u := &UnaryOp{Op: OpNeg, X: &Ident{Name: "balance"}}
		got := u.String()
		if got != "(-balance)" {
			t.Fatalf("expected (-balance), got %s", got)
		}
	})

	t.Run("AndExpr", func(t *testing.T) {
		t.Parallel()
		a := &AndExpr{L: &Ident{Name: "a"}, R: &Ident{Name: "b"}}
		got := a.String()
		if got != "(a AND b)" {
			t.Fatalf("expected (a AND b), got %s", got)
		}
	})

	t.Run("OrExpr", func(t *testing.T) {
		t.Parallel()
		o := &OrExpr{L: &Ident{Name: "a"}, R: &Ident{Name: "b"}}
		got := o.String()
		if got != "(a OR b)" {
			t.Fatalf("expected (a OR b), got %s", got)
		}
	})

	t.Run("NotExpr", func(t *testing.T) {
		t.Parallel()
		n := &NotExpr{X: &Ident{Name: "blocked"}}
		got := n.String()
		if got != "(NOT blocked)" {
			t.Fatalf("expected (NOT blocked), got %s", got)
		}
	})

	t.Run("RangeExpr inclusive", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "age"}, Lo: &NumberLit{Value: 18}, Hi: &NumberLit{Value: 65},
			LoIncl: true, HiIncl: true,
		}
		got := r.String()
		if got != "(age IN [18..65])" {
			t.Fatalf("expected (age IN [18..65]), got %s", got)
		}
	})

	t.Run("RangeExpr exclusive", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "x"}, Lo: &NumberLit{Value: 0}, Hi: &NumberLit{Value: 1},
			LoIncl: false, HiIncl: false,
		}
		got := r.String()
		if got != "(x IN (0..1))" {
			t.Fatalf("expected (x IN (0..1)), got %s", got)
		}
	})

	t.Run("RangeExpr mixed", func(t *testing.T) {
		t.Parallel()
		r := &RangeExpr{
			X: &Ident{Name: "x"}, Lo: &NumberLit{Value: 0}, Hi: &NumberLit{Value: 1},
			LoIncl: true, HiIncl: false,
		}
		got := r.String()
		if got != "(x IN [0..1))" {
			t.Fatalf("expected (x IN [0..1)), got %s", got)
		}
	})

	t.Run("CallExpr no args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "now", Args: nil}
		got := c.String()
		if got != "now()" {
			t.Fatalf("expected now(), got %s", got)
		}
	})

	t.Run("CallExpr with args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "sum", Args: []Expr{&Ident{Name: "payments"}}}
		got := c.String()
		if got != "sum(payments)" {
			t.Fatalf("expected sum(payments), got %s", got)
		}
	})

	t.Run("CallExpr multiple args", func(t *testing.T) {
		t.Parallel()
		c := &CallExpr{Name: "contains", Args: []Expr{&Ident{Name: "list"}, &StringLit{Value: "x"}}}
		expected := `contains(list, "x")`
		got := c.String()
		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})
}
