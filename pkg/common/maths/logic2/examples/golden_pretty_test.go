package examples

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/engine"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// TestPrettyExplain_Golden compares PrettyExplain output with a golden file to
// ensure deterministic, user-facing formatting. If this test fails due to an
// intentional formatting change, update the golden file accordingly.
func TestPrettyExplain_Golden(t *testing.T) {
	rules := []engine.Rule{
		{ID: "r1", When: parser.MustParse("A & B"), Then: p.Var("C")},
		{ID: "r2", When: parser.MustParse("C => D"), Then: p.Var("D")},
	}
	e := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	e.Assert(p.Var("A"))
	e.Assert(p.Var("B"))
	_ = e.RunToFixpoint(5)

	ok, why := e.Query(parser.MustParse("D"))
	if !ok || why == nil {
		t.Fatalf("expected D to be true with explanation")
	}

	out := engine.PrettyExplain(why)
	goldenPath := filepath.Join("testdata", "pretty_explain_D.golden")
	b, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	golden := string(b)
	if out != golden {
		t.Fatalf("PrettyExplain output mismatch with golden file.\n--- got ---\n%s\n--- want (golden) ---\n%s", out, golden)
	}
}
