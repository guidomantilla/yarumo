package examples

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic/engine"
	"github.com/guidomantilla/yarumo/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// TestPrettyExplain_Simple builds a tiny rule system and pretty-prints an explanation
// for a queried goal. It asserts a few stable substrings to keep the test robust
// while allowing harmless formatting changes elsewhere.
func TestPrettyExplain_Simple(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	_ = eng.RunToFixpoint(5)

	ok, why := eng.Query(parser.MustParse("D"))
	if !ok || why == nil {
		t.Fatalf("expected D to be true with explanation, got ok=%v, why=%v", ok, why)
	}

	out := engine.PrettyExplain(why)
	if !strings.Contains(out, "=>") {
		t.Fatalf("expected implication to appear in explanation; got:\n%s", out)
	}
	if !strings.Contains(out, "both true") && !strings.Contains(out, "at least one true") {
		t.Fatalf("expected a summary message in explanation; got:\n%s", out)
	}
	if !strings.Contains(out, "fact: A=true") {
		t.Fatalf("expected leaf facts to appear; got:\n%s", out)
	}
}

// TestPrettyExplainTo_WriterDeterministic ensures that PrettyExplainTo writes the same
// deterministic output as PrettyExplain when given the same tree.
func TestPrettyExplainTo_WriterDeterministic(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & (B | C)", "X"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("C"))
	_ = eng.RunToFixpoint(3)

	ok, why := eng.Query(parser.MustParse("X"))
	if !ok || why == nil {
		t.Fatalf("expected X to be true with explanation")
	}

	want := engine.PrettyExplain(why)
	var buf bytes.Buffer
	engine.PrettyExplainTo(&buf, why)
	got := buf.String()
	if got != want {
		t.Fatalf("PrettyExplainTo mismatch with PrettyExplain\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
