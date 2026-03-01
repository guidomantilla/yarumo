package logic

import "testing"

func TestErrNilFormula(t *testing.T) {
	t.Parallel()

	got := ErrNilFormula.Error()

	expected := "formula is nil"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
