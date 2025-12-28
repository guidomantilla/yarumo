package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// helper to assert equivalence between a synonym expression and a canonical expression
func mustEqParse(t *testing.T, syn, canon string) {
	t.Helper()

	fs := parser.MustParse(syn)

	fc := parser.MustParse(canon)
	if !p.Equivalent(fs, fc) {
		t.Fatalf("not equivalent: syn=%q -> %s, canon=%q -> %s", syn, fs.String(), canon, fc.String())
	}
}

func TestParserSynonyms_AsciiAndKeywords(t *testing.T) {
	cases := [][2]string{
		{"A AND B", "(A & B)"},
		{"A && B", "(A & B)"},
		{"A OR B", "(A | B)"},
		{"A || B", "(A | B)"},
		{"NOT A", "!A"},
		{"A THEN B", "(A => B)"},
		{"A -> B", "(A => B)"},
		{"A <-> B", "(A <=> B)"},
		{"A IFF B", "(A <=> B)"},
	}
	for _, c := range cases {
		mustEqParse(t, c[0], c[1])
	}
}

func TestParserSynonyms_Unicode(t *testing.T) {
	cases := [][2]string{
		{"A ∧ B", "(A & B)"},
		{"A ∨ B", "(A | B)"},
		{"¬A", "!A"},
		{"A → B", "(A => B)"},
		{"A ⇒ B", "(A => B)"},
		{"A ↔ B", "(A <=> B)"},
		{"A ⇔ B", "(A <=> B)"},
	}
	for _, c := range cases {
		mustEqParse(t, c[0], c[1])
	}
}

func TestParserSynonyms_TrueFalseLiterals(t *testing.T) {
	fTrue := parser.MustParse("TRUE")
	if fTrue.String() != "⊤" {
		t.Fatalf("TRUE should parse to ⊤, got %q", fTrue.String())
	}

	fFalse := parser.MustParse("FALSE")
	if fFalse.String() != "⊥" {
		t.Fatalf("FALSE should parse to ⊥, got %q", fFalse.String())
	}
}

func TestParserSynonyms_IdentifiersNotKeywords(t *testing.T) {
	// Ensure identifiers that contain keyword substrings are not split or reinterpreted.
	f := parser.MustParse("ANDY | Orion")
	if f.String() != "(ANDY | Orion)" {
		t.Fatalf("expected identifiers to be preserved as variables, got %q", f.String())
	}
}

func TestParser_StrictModeDisallowsSynonyms(t *testing.T) {
	// In strict mode, only canonical operators are allowed. Keywords become identifiers.
	if _, err := parser.ParseWith("A AND B", parser.ParseOptions{Strict: true}); err == nil {
		t.Fatalf("expected error in strict mode for 'A AND B'")
	}

	if _, err := parser.ParseWith("A && B", parser.ParseOptions{Strict: true}); err == nil {
		t.Fatalf("expected error in strict mode for 'A && B'")
	}

	if _, err := parser.ParseWith("A ∧ B", parser.ParseOptions{Strict: true}); err == nil {
		t.Fatalf("expected error in strict mode for 'A ∧ B'")
	}
	// Canonical should still work
	if _, err := parser.ParseWith("A & B", parser.ParseOptions{Strict: true}); err != nil {
		t.Fatalf("expected canonical '&' to work in strict mode: %v", err)
	}
	// TRUE becomes an identifier in strict mode (parsed as Var("TRUE"))
	g, err := parser.ParseWith("TRUE", parser.ParseOptions{Strict: true})
	if err != nil {
		t.Fatalf("unexpected error parsing 'TRUE' in strict mode: %v", err)
	}

	if g.String() != "TRUE" {
		t.Fatalf("in strict mode, TRUE should be parsed as identifier, got %q", g.String())
	}
}
