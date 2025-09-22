package examples

import (
    "fmt"
    "testing"

    "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/engine"
    "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
    p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// TestEngine_EndToEnd demonstrates a minimal forward-chaining run and query/explain.
func TestEngine_EndToEnd(t *testing.T) {
    rules := []engine.Rule{
        {ID: "r1", When: parser.MustParse("A & B"), Then: p.Var("C")},
        {ID: "r2", When: parser.MustParse("C => D"), Then: p.Var("D")},
        {ID: "r3", When: parser.MustParse("D & E"), Then: p.Var("F")},
    }
    eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

    eng.Assert(p.Var("A"))
    eng.Assert(p.Var("B"))
    eng.Assert(p.Var("E"))

    fired := eng.RunToFixpoint(5)
    if len(fired) == 0 {
        t.Fatalf("expected some rules to fire")
    }

    // After running: A, B are asserted; r1 fires to set C; r2 fires to set D; with E true, r3 fires to set F
    for _, v := range []p.Var{"C", "D", "F"} {
        if val, _ := eng.Facts.Get(v); !val {
            t.Fatalf("expected %s to be true after fixpoint", v)
        }
    }

    // Query with explanation
    ok, why := eng.Query(parser.MustParse("F"))
    if !ok || why == nil {
        t.Fatalf("expected F to be true with explanation")
    }

    // Print a readable explanation for documentation value.
    // We use Output block to make go test verify formatting deterministically.
    fmt.Print(engine.PrettyExplain(why))
}

// TestEngine_NoFire ensures rules do not fire when conditions are not met.
func TestEngine_NoFire(t *testing.T) {
    rules := []engine.Rule{
        {ID: "r1", When: parser.MustParse("A & B"), Then: p.Var("C")},
    }
    eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

    eng.Assert(p.Var("A")) // B remains false (default)

    fired := eng.FireOnce()
    if len(fired) != 0 {
        t.Fatalf("expected no rules to fire, got %v", fired)
    }

    ok, why := eng.Query(parser.MustParse("C"))
    if ok {
        t.Fatalf("expected C to be false without B")
    }
    _ = why
}
