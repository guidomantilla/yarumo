package examples

import (
	"reflect"
	"testing"

	"github.com/guidomantilla/yarumo/modules/common/maths/logic/parser"
)

// TestVarsOrder_Stable ensures Vars() returns a stable, sorted order of variable names.
func TestVarsOrder_Stable(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"A & B", []string{"A", "B"}},
		{"B & A", []string{"A", "B"}},
		{"(C | A) & B", []string{"A", "B", "C"}},
		{"((Z | Y) & X) | A", []string{"A", "X", "Y", "Z"}},
	}
	for _, c := range cases {
		f := parser.MustParse(c.in)
		got := f.Vars()
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("Vars() not sorted/stable for %q: got %v, want %v", c.in, got, c.want)
		}
	}
}
