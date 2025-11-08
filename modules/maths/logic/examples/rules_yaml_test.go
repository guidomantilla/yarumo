package examples

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic/engine"
)

func TestRulesYAML_RoundTrip(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
	}
	// Encode to YAML
	var buf bytes.Buffer
	if err := engine.SaveRulesYAML(&buf, rules); err != nil {
		t.Fatalf("save rules yaml: %v", err)
	}
	first := buf.String()
	if !strings.Contains(first, "version: v1") {
		t.Fatalf("expected version v1 in yaml, got: %s", first)
	}

	// Decode back
	gotRules, err := engine.LoadRulesYAML(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("load rules yaml: %v", err)
	}
	if len(gotRules) != len(rules) {
		t.Fatalf("rule count mismatch: got %d want %d", len(gotRules), len(rules))
	}
	for i := range rules {
		if !rules[i].Equals(gotRules[i]) {
			t.Fatalf("rule %d mismatch after yaml round-trip", i)
		}
	}

	// Re-encode and compare canonicalization (reasonable exactness with yaml.Encoder)
	buf2 := bytes.Buffer{}
	if err := engine.SaveRulesYAML(&buf2, gotRules); err != nil {
		t.Fatalf("save rules yaml 2: %v", err)
	}
	second := buf2.String()
	if first != second {
		// yaml.Encoder should be deterministic given the same struct values and field order
		t.Fatalf("non-stable yaml round-trip.\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}
