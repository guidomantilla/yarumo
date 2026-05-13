package tokens

import (
	stded25519 "crypto/ed25519"
	"crypto/elliptic"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"

	ecdsas "github.com/guidomantilla/yarumo/common/crypto/signers/ecdsas"
	ed25519signer "github.com/guidomantilla/yarumo/common/crypto/signers/ed25519"
	rsassas "github.com/guidomantilla/yarumo/common/crypto/signers/rsassas"
)

// TestYA0018_AsymmetricAlgorithms verifies that each predefined asymmetric
// JWT algorithm round-trips a Sign + Validate. The keys are produced through
// the signers/* helpers so the test exercises the same generation path real
// callers will use.
func TestYA0018_AsymmetricAlgorithms(t *testing.T) {
	t.Parallel()

	t.Run("RS256 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmRS256, rsassas.RSASSA_PKCS1v15_using_SHA256, 2048)
	})

	t.Run("RS384 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmRS384, rsassas.RSASSA_PKCS1v15_using_SHA384, 2048)
	})

	t.Run("RS512 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmRS512, rsassas.RSASSA_PKCS1v15_using_SHA512, 3072)
	})

	t.Run("PS256 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmPS256, rsassas.RSASSA_PSS_using_SHA256, 2048)
	})

	t.Run("PS384 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmPS384, rsassas.RSASSA_PSS_using_SHA384, 2048)
	})

	t.Run("PS512 roundtrip", func(t *testing.T) {
		t.Parallel()
		runRSARoundtrip(t, AlgorithmPS512, rsassas.RSASSA_PSS_using_SHA512, 3072)
	})

	t.Run("ES256 roundtrip", func(t *testing.T) {
		t.Parallel()
		runECDSARoundtrip(t, AlgorithmES256, ecdsas.ECDSA_with_SHA256_over_P256)
	})

	t.Run("ES384 roundtrip", func(t *testing.T) {
		t.Parallel()
		runECDSARoundtrip(t, AlgorithmES384, ecdsas.ECDSA_with_SHA384_over_P384)
	})

	t.Run("ES512 roundtrip", func(t *testing.T) {
		t.Parallel()
		runECDSARoundtrip(t, AlgorithmES512, ecdsas.ECDSA_with_SHA512_over_P521)
	})

	t.Run("EdDSA roundtrip", func(t *testing.T) {
		t.Parallel()
		runEdDSARoundtrip(t)
	})
}

func runRSARoundtrip(t *testing.T, alg Algorithm, signer *rsassas.Method, bits int) {
	t.Helper()

	priv, err := signer.GenerateKey(bits)
	if err != nil {
		t.Fatalf("rsa GenerateKey(%d) failed: %v", bits, err)
	}

	m := NewMethod("test_"+string(alg), alg,
		WithSigningKey(priv),
		WithVerifyingKey(&priv.PublicKey),
	)

	const subject = "user@asym"
	expected := "role-" + string(alg)
	token, err := m.Generate(subject, Payload{"role": expected})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if payload["role"] != expected {
		t.Fatalf("expected role %q, got %v", expected, payload["role"])
	}
}

func runECDSARoundtrip(t *testing.T, alg Algorithm, signer *ecdsas.Method) {
	t.Helper()

	priv, err := signer.GenerateKey()
	if err != nil {
		t.Fatalf("ecdsa GenerateKey failed: %v", err)
	}

	m := NewMethod("test_"+string(alg), alg,
		WithSigningKey(priv),
		WithVerifyingKey(&priv.PublicKey),
	)

	const subject = "user@asym"
	expected := "role-" + string(alg)
	token, err := m.Generate(subject, Payload{"role": expected})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if payload["role"] != expected {
		t.Fatalf("expected role %q, got %v", expected, payload["role"])
	}

	// Sanity-check the curve actually matches the algorithm name.
	wantCurve := map[Algorithm]elliptic.Curve{
		AlgorithmES256: elliptic.P256(),
		AlgorithmES384: elliptic.P384(),
		AlgorithmES512: elliptic.P521(),
	}
	want, ok := wantCurve[alg]
	if ok && priv.Curve != want {
		t.Fatalf("expected curve %v, got %v", want.Params().Name, priv.Curve.Params().Name)
	}
}

func runEdDSARoundtrip(t *testing.T) {
	t.Helper()

	priv, err := ed25519signer.Ed25519.GenerateKey()
	if err != nil {
		t.Fatalf("ed25519 GenerateKey failed: %v", err)
	}

	pub, ok := priv.Public().(stded25519.PublicKey)
	if !ok {
		t.Fatalf("expected ed25519.PublicKey, got %T", priv.Public())
	}

	m := NewMethod("test_EdDSA", AlgorithmEdDSA,
		WithSigningKey(priv),
		WithVerifyingKey(pub),
	)

	const subject = "user@asym"
	const expected = "role-EdDSA"
	token, err := m.Generate(subject, Payload{"role": expected})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if payload["role"] != expected {
		t.Fatalf("expected role %q, got %v", expected, payload["role"])
	}
}

// TestYA0018_SigningMethodFor_AsymmetricMappings verifies the enum -> jwt
// signing-method mapping for every new asymmetric algorithm constant.
func TestYA0018_SigningMethodFor_AsymmetricMappings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		alg  Algorithm
		want jwt.SigningMethod
	}{
		{"RS256", AlgorithmRS256, jwt.SigningMethodRS256},
		{"RS384", AlgorithmRS384, jwt.SigningMethodRS384},
		{"RS512", AlgorithmRS512, jwt.SigningMethodRS512},
		{"PS256", AlgorithmPS256, jwt.SigningMethodPS256},
		{"PS384", AlgorithmPS384, jwt.SigningMethodPS384},
		{"PS512", AlgorithmPS512, jwt.SigningMethodPS512},
		{"ES256", AlgorithmES256, jwt.SigningMethodES256},
		{"ES384", AlgorithmES384, jwt.SigningMethodES384},
		{"ES512", AlgorithmES512, jwt.SigningMethodES512},
		{"EdDSA", AlgorithmEdDSA, jwt.SigningMethodEdDSA},
	}

	for _, tc := range cases {
		t.Run(tc.name+" maps correctly", func(t *testing.T) {
			t.Parallel()
			got := signingMethodFor(tc.alg)
			if got != tc.want {
				t.Fatalf("signingMethodFor(%q) = %v, want %v", tc.alg, got, tc.want)
			}
		})
	}
}

// TestYA0018_RegistryHasAsymmetricMethods asserts the new predefined
// JWT_* methods are exposed via the package registry.
func TestYA0018_RegistryHasAsymmetricMethods(t *testing.T) {
	t.Parallel()

	names := []string{
		"JWT_RS256", "JWT_RS384", "JWT_RS512",
		"JWT_PS256", "JWT_PS384", "JWT_PS512",
		"JWT_ES256", "JWT_ES384", "JWT_ES512",
		"JWT_EdDSA",
	}

	for _, name := range names {
		t.Run(name+" resolves", func(t *testing.T) {
			t.Parallel()
			m, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%q) failed: %v", name, err)
			}
			if m.Name() != name {
				t.Fatalf("expected name %q, got %q", name, m.Name())
			}
		})
	}
}
