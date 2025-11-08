package examples

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/common/maths/logic/engine"
	"github.com/guidomantilla/yarumo/common/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/common/maths/logic/props"
)

func TestRulesJSON_RoundTrip(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
	}
	// Encode to JSON
	var buf bytes.Buffer
	if err := engine.SaveRulesJSON(&buf, rules); err != nil {
		t.Fatalf("save rules json: %v", err)
	}
	first := buf.String()
	if !strings.Contains(first, "\"version\": \"v1\"") {
		t.Fatalf("expected version v1 in json, got: %s", first)
	}

	// Decode back
	gotRules, err := engine.LoadRulesJSON(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("load rules json: %v", err)
	}
	if len(gotRules) != len(rules) {
		t.Fatalf("rule count mismatch: got %d want %d", len(gotRules), len(rules))
	}
	for i := range rules {
		if !rules[i].Equals(gotRules[i]) {
			t.Fatalf("rule %d mismatch after round-trip", i)
		}
	}

	// Re-encode and compare canonical strings (stable modulo whitespace)
	buf2 := bytes.Buffer{}
	if err := engine.SaveRulesJSON(&buf2, gotRules); err != nil {
		t.Fatalf("save rules json 2: %v", err)
	}
	second := buf2.String()
	if first != second {
		// As we use MarshalIndent with stable struct order, these should match exactly
		t.Fatalf("non-stable round-trip json.\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestExplainDTO_JSON(t *testing.T) {
	// Build a small engine and query to get an explanation
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	_ = eng.RunToFixpoint(3)
	ok, why := eng.Query(parser.MustParse("C"))
	if !ok || why == nil {
		t.Fatalf("expected C to be true with explanation")
	}

	// Convert to DTO and marshal JSON
	dto := engine.ToDTO(why)
	b, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		t.Fatalf("marshal explain dto: %v", err)
	}
	js := string(b)
	if !strings.Contains(js, "\"expr\"") || !strings.Contains(js, "\"value\"") {
		t.Fatalf("unexpected explain json: %s", js)
	}
}
