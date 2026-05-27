package utils

import (
	"testing"

	"golang.org/x/text/language"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("default charset", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o.charset != AllCharset {
			t.Fatalf("got %q, want %q", o.charset, AllCharset)
		}
	})

	t.Run("default language", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o.lang != language.English {
			t.Fatalf("got %v, want %v", o.lang, language.English)
		}
	})

	t.Run("with combined options", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(
			WithCharset("ABC"),
			WithLanguage(language.German),
		)
		if o.charset != "ABC" || o.lang != language.German {
			t.Fatalf("charset=%q, lang=%v", o.charset, o.lang)
		}
	})
}

func TestWithCharset(t *testing.T) {
	t.Parallel()

	t.Run("non-empty overrides default", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithCharset("ABC"))
		if o.charset != "ABC" {
			t.Fatalf("got %q, want %q", o.charset, "ABC")
		}
	})

	t.Run("empty keeps default", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithCharset(""))
		if o.charset != AllCharset {
			t.Fatalf("got %q, want %q", o.charset, AllCharset)
		}
	})
}

func TestWithLanguage(t *testing.T) {
	t.Parallel()

	t.Run("changes from default", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithLanguage(language.Spanish))
		if o.lang != language.Spanish {
			t.Fatalf("got %v, want %v", o.lang, language.Spanish)
		}
	})
}
