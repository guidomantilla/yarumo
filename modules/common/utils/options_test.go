package utils

import (
    "testing"

    "golang.org/x/text/language"
)

func TestNewOptions_Defaults(t *testing.T) {
    o := NewOptions()
    if o == nil {
        t.Fatalf("NewOptions returned nil")
    }
    if o.charset != AllCharset {
        t.Fatalf("default charset = %q, want %q", o.charset, AllCharset)
    }
    if o.lang != language.English {
        t.Fatalf("default language = %v, want %v", o.lang, language.English)
    }
}

func TestWithCharset(t *testing.T) {
    // non-empty should override
    o := NewOptions(WithCharset("ABC"))
    if o.charset != "ABC" {
        t.Fatalf("charset = %q, want %q", o.charset, "ABC")
    }
    // empty should not override (stays default)
    o2 := NewOptions(WithCharset(""))
    if o2.charset != AllCharset {
        t.Fatalf("empty WithCharset should keep default; got %q, want %q", o2.charset, AllCharset)
    }
}

func TestWithLanguage(t *testing.T) {
    // change from default English to Spanish
    o := NewOptions(WithLanguage(language.Spanish))
    if o.lang != language.Spanish {
        t.Fatalf("language = %v, want %v", o.lang, language.Spanish)
    }
}

func TestCombinedOptions_OrderAndCoexistence(t *testing.T) {
    o := NewOptions(
        WithCharset("UTF-8"),
        WithLanguage(language.German),
    )
    if o.charset != "UTF-8" {
        t.Fatalf("combined: charset = %q, want %q", o.charset, "UTF-8")
    }
    if o.lang != language.German {
        t.Fatalf("combined: language = %v, want %v", o.lang, language.German)
    }
}
