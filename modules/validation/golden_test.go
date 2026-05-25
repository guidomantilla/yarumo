package validation

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

// goldenRuleset loads a YAML ruleset from fixtures/golden/ and returns the
// engine bound to it. Test failures abort.
func goldenRuleset(t *testing.T, file string) Engine {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("fixtures", "golden", file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}

	rs, err := LoadYAML(data)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}

	return NewEngine(rs)
}

type userRegistration struct {
	Email           string
	Password        string
	ConfirmPassword string
	Age             int
	TermsAccepted   bool
}

func TestGolden_UserRegistration_Valid(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "user_registration.yaml")

	payload := userRegistration{
		Email:           "ash@kanto.com",
		Password:        "longerthan8",
		ConfirmPassword: "secret123",
		Age:             21,
		TermsAccepted:   true,
	}

	err := eng.Validate(payload, nil)
	if err != nil {
		t.Fatalf("expected no violations, got %v", err)
	}
}

func TestGolden_UserRegistration_MultipleViolations(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "user_registration.yaml")

	payload := userRegistration{
		Email:           "not-an-email",
		Password:        "short",
		ConfirmPassword: "different",
		Age:             5,
		TermsAccepted:   false,
	}

	err := eng.Validate(payload, nil)
	if err == nil {
		t.Fatalf("expected violations, got nil")
	}

	for _, want := range []error{cvalidation.ErrEmailInvalid, cvalidation.ErrMinLen, cvalidation.ErrNotEqual, cvalidation.ErrOutOfRange, cvalidation.ErrFieldRequired} {
		if !errors.Is(err, want) {
			t.Errorf("expected %v wrapped, got %v", want, err)
		}
	}
}

type apiKey struct {
	Name      string
	Scopes    []string
	ExpiresAt string
}

func TestGolden_APIKey_Valid(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "api_key.yaml")

	payload := apiKey{
		Name:      "prod-key",
		Scopes:    []string{"read", "write"},
		ExpiresAt: "2027-01-01T00:00:00Z",
	}

	err := eng.Validate(payload, nil)
	if err != nil {
		t.Fatalf("expected no violations, got %v", err)
	}
}

func TestGolden_APIKey_InvalidScope(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "api_key.yaml")

	payload := apiKey{
		Name:      "prod-key",
		Scopes:    []string{"read", "delete"},
		ExpiresAt: "2027-01-01T00:00:00Z",
	}

	err := eng.Validate(payload, nil)
	if !errors.Is(err, cvalidation.ErrNotInAllowed) {
		t.Fatalf("expected ErrNotInAllowed for delete scope, got %v", err)
	}
}

type webhook struct {
	URL    string
	Events []string
	Secret string
}

func TestGolden_Webhook_Valid(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "webhook.yaml")

	payload := webhook{
		URL:    "https://example.com/hook",
		Events: []string{"order.created", "order.deleted"},
		Secret: "supersecretkey1234567890",
	}

	err := eng.Validate(payload, nil)
	if err != nil {
		t.Fatalf("expected no violations, got %v", err)
	}
}

func TestGolden_Webhook_MissingFields(t *testing.T) {
	t.Parallel()

	eng := goldenRuleset(t, "webhook.yaml")

	payload := webhook{
		URL:    "",
		Events: []string{},
		Secret: "short",
	}

	err := eng.Validate(payload, nil)
	if err == nil {
		t.Fatalf("expected violations, got nil")
	}

	for _, want := range []error{cvalidation.ErrFieldRequired, cvalidation.ErrCollectionEmpty, cvalidation.ErrMinLen} {
		if !errors.Is(err, want) {
			t.Errorf("expected %v wrapped, got %v", want, err)
		}
	}
}
