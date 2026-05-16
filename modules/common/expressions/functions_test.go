package expressions

import "testing"

func TestDefaultFuncs(t *testing.T) {
	t.Parallel()

	t.Run("returns all 9 built-in functions", func(t *testing.T) {
		t.Parallel()
		funcs := DefaultFuncs()
		expected := []string{"len", "sum", "min", "max", "avg", "abs", "contains", "lower", "upper"}
		for _, name := range expected {
			_, ok := funcs[name]
			if !ok {
				t.Fatalf("expected function %s", name)
			}
		}
	})
}
