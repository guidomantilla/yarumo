package expressions

import (
	"testing"
)

func TestTypeCompliance(t *testing.T) {
	t.Parallel()

	t.Run("NumberLit implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*NumberLit)(nil)
	})

	t.Run("StringLit implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*StringLit)(nil)
	})

	t.Run("BoolLit implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*BoolLit)(nil)
	})

	t.Run("NilLit implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*NilLit)(nil)
	})

	t.Run("Ident implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*Ident)(nil)
	})

	t.Run("Property implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*Property)(nil)
	})

	t.Run("BinaryOp implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*BinaryOp)(nil)
	})

	t.Run("UnaryOp implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*UnaryOp)(nil)
	})

	t.Run("AndExpr implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*AndExpr)(nil)
	})

	t.Run("OrExpr implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*OrExpr)(nil)
	})

	t.Run("NotExpr implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*NotExpr)(nil)
	})

	t.Run("RangeExpr implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*RangeExpr)(nil)
	})

	t.Run("CallExpr implements Expr", func(t *testing.T) {
		t.Parallel()
		var _ Expr = (*CallExpr)(nil)
	})
}
