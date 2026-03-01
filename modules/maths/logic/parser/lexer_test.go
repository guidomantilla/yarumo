package parser

import "testing"

func TestLex(t *testing.T) {
	t.Parallel()

	t.Run("bare equals sign", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A = B")

		found := false

		for _, tok := range tokens {
			if tok.val == "=" && tok.kind == tokVar {
				found = true
			}
		}

		if !found {
			t.Fatal("expected bare = to be lexed as tokVar")
		}
	})

	t.Run("bare less than", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A < B")

		found := false

		for _, tok := range tokens {
			if tok.val == "<" && tok.kind == tokVar {
				found = true
			}
		}

		if !found {
			t.Fatal("expected bare < to be lexed as tokVar")
		}
	})

	t.Run("bare dash", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A - B")

		found := false

		for _, tok := range tokens {
			if tok.val == "-" && tok.kind == tokVar {
				found = true
			}
		}

		if !found {
			t.Fatal("expected bare - to be lexed as tokVar")
		}
	})

	t.Run("single v as or", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A v B")

		found := false

		for _, tok := range tokens {
			if tok.kind == tokOr {
				found = true
			}
		}

		if !found {
			t.Fatal("expected v to be lexed as tokOr")
		}
	})

	t.Run("F as false", func(t *testing.T) {
		t.Parallel()

		tokens := lex("F")

		if tokens[0].kind != tokFalse {
			t.Fatal("expected F to be lexed as tokFalse")
		}
	})

	t.Run("multi char variable starting with v", func(t *testing.T) {
		t.Parallel()

		tokens := lex("var1")

		if tokens[0].kind != tokVar || tokens[0].val != "var1" {
			t.Fatalf("expected variable var1, got kind=%d val=%s", tokens[0].kind, tokens[0].val)
		}
	})

	t.Run("eof at end", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A")

		last := tokens[len(tokens)-1]
		if last.kind != tokEOF {
			t.Fatal("expected EOF as last token")
		}
	})

	t.Run("less than with equals not followed by greater", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A <= B")

		// <= is not a valid operator, should be lexed as < and then = and then B
		found := false

		for _, tok := range tokens {
			if tok.val == "<" {
				found = true
			}
		}

		if !found {
			t.Fatal("expected < to appear in tokens")
		}
	})

	t.Run("unknown character becomes var", func(t *testing.T) {
		t.Parallel()

		tokens := lex("#")
		if tokens[0].kind != tokVar || tokens[0].val != "#" {
			t.Fatalf("expected # as tokVar, got kind=%d val=%s", tokens[0].kind, tokens[0].val)
		}
	})

	t.Run("single pipe as or", func(t *testing.T) {
		t.Parallel()

		tokens := lex("A | B")

		found := false

		for _, tok := range tokens {
			if tok.kind == tokOr && tok.val == "|" {
				found = true
			}
		}

		if !found {
			t.Fatal("expected | to be lexed as tokOr")
		}
	})
}
