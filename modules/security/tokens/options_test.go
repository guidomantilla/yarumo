package tokens

import (
	"testing"
	"time"

	jwt "github.com/golang-jsonwebtoken/jsonwebtoken/v5"
)

func TestNewOptions_DefaultsAndOverrides(t *testing.T) {
	// Defaults
	def := NewOptions()
	if def == nil {
		t.Fatalf("expected options, got nil")
	}
	if def.timeout != 24*time.Hour {
		t.Fatalf("expected default timeout 24h, got %v", def.timeout)
	}
	if len(def.cipherKey) == 0 || len(def.signingKey) == 0 || len(def.verifyingKey) == 0 {
		t.Fatalf("expected default keys to be generated")
	}
	if def.signingMethod != jwt.SigningMethodHS512 {
		t.Fatalf("expected default signing method HS512, got %v", def.signingMethod)
	}

	// WithTimeout positive should be ignored (see implementation)
	pos := NewOptions(WithTimeout(1 * time.Minute))
	if pos.timeout != def.timeout {
		t.Fatalf("positive timeout should be ignored, got %v", pos.timeout)
	}

	// WithTimeout negative should be applied
	neg := NewOptions(WithTimeout(-1 * time.Minute))
	if neg.timeout != -1*time.Minute {
		t.Fatalf("negative timeout should be applied, got %v", neg.timeout)
	}

	// WithJwtIssuer
	oi := NewOptions(WithJwtIssuer("issuer-x"))
	if oi.issuer != "issuer-x" {
		t.Fatalf("issuer not applied: %v", oi.issuer)
	}

	// WithJwtKey (also sets verifying)
	key := []byte("k123")
	ok := NewOptions(WithJwtKey(key))
	if string(ok.signingKey) != string(key) || string(ok.verifyingKey) != string(key) {
		t.Fatalf("jsonwebtoken key not applied correctly")
	}

	// WithJwtSigningMethod
	om := NewOptions(WithJwtSigningMethod(jwt.SigningMethodHS256))
	if om.signingMethod != jwt.SigningMethodHS256 {
		t.Fatalf("signing method not applied")
	}

	// WithOpaqueKey
	ock := []byte("opaque-key")
	oo := NewOptions(WithOpaqueKey(ock))
	if string(oo.cipherKey) != string(ock) {
		t.Fatalf("opaque key not applied")
	}
}
