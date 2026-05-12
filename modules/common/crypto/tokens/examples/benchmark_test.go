package main

import (
	"crypto/ed25519"
	"testing"

	caead "github.com/guidomantilla/yarumo/common/crypto/ciphers/aead"
	cecdsas "github.com/guidomantilla/yarumo/common/crypto/signers/ecdsas"
	ced25519 "github.com/guidomantilla/yarumo/common/crypto/signers/ed25519"
	crsassas "github.com/guidomantilla/yarumo/common/crypto/signers/rsassas"
	ctokens "github.com/guidomantilla/yarumo/common/crypto/tokens"
)

// benchmarkSubject is the JWT subject claim used by token benchmarks.
const benchmarkSubject = "user-bench"

// benchmarkPayload is the JWT custom payload used by token benchmarks. The
// fields mirror a typical authorization claim set.
var benchmarkPayload = ctokens.Payload{
	"role":  "admin",
	"scope": "read:write",
}

// tokenCase pairs a fully-keyed *ctokens.Method with a stable name used as
// the benchmark sub-name. The keyed methods are built once at suite setup
// and re-used across Generate / Validate benchmarks.
type tokenCase struct {
	name   string
	method *ctokens.Method
}

// buildTokenCases constructs one fully-keyed Method per predefined algorithm.
// Asymmetric keys are generated at the smallest supported size to keep
// suite setup time bounded; benchmark steady-state cost dominates the
// reported per-iteration numbers regardless of setup wall time.
func buildTokenCases(tb testing.TB) []tokenCase {
	tb.Helper()

	rs256Priv, err := crsassas.RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
	if err != nil {
		tb.Fatalf("RS key generation failed: %v", err)
	}

	ps256Priv, err := crsassas.RSASSA_PSS_using_SHA256.GenerateKey(2048)
	if err != nil {
		tb.Fatalf("PS key generation failed: %v", err)
	}

	es256Priv, err := cecdsas.ECDSA_with_SHA256_over_P256.GenerateKey()
	if err != nil {
		tb.Fatalf("ES256 key generation failed: %v", err)
	}

	es384Priv, err := cecdsas.ECDSA_with_SHA384_over_P384.GenerateKey()
	if err != nil {
		tb.Fatalf("ES384 key generation failed: %v", err)
	}

	es512Priv, err := cecdsas.ECDSA_with_SHA512_over_P521.GenerateKey()
	if err != nil {
		tb.Fatalf("ES512 key generation failed: %v", err)
	}

	edPriv, err := ced25519.Ed25519.GenerateKey()
	if err != nil {
		tb.Fatalf("Ed25519 key generation failed: %v", err)
	}

	edPub, ok := edPriv.Public().(ed25519.PublicKey)
	if !ok {
		tb.Fatal("ed25519 public key type assertion failed")
	}

	aesKeyBytes, err := caead.AES_256_GCM.GenerateKey()
	if err != nil {
		tb.Fatalf("AES-256-GCM key generation failed: %v", err)
	}

	// opaqueAEADKey unwraps the stored key via a []byte type assertion, so
	// the ctypes.Bytes return value must be re-typed before being handed off
	// via WithKey.
	aesKey := []byte(aesKeyBytes)

	xchachaKeyBytes, err := caead.XCHACHA20_POLY1305.GenerateKey()
	if err != nil {
		tb.Fatalf("XChaCha20-Poly1305 key generation failed: %v", err)
	}

	xchachaKey := []byte(xchachaKeyBytes)

	return []tokenCase{
		{
			name:   "JWT_HS256",
			method: ctokens.NewMethod("JWT_HS256_bench", ctokens.AlgorithmHS256, ctokens.WithGeneratedKey()),
		},
		{
			name:   "JWT_HS384",
			method: ctokens.NewMethod("JWT_HS384_bench", ctokens.AlgorithmHS384, ctokens.WithGeneratedKey()),
		},
		{
			name:   "JWT_HS512",
			method: ctokens.NewMethod("JWT_HS512_bench", ctokens.AlgorithmHS512, ctokens.WithGeneratedKey()),
		},
		{
			name:   "JWT_RS256",
			method: ctokens.NewMethod("JWT_RS256_bench", ctokens.AlgorithmRS256, ctokens.WithSigningKey(rs256Priv), ctokens.WithVerifyingKey(&rs256Priv.PublicKey)),
		},
		{
			name:   "JWT_PS256",
			method: ctokens.NewMethod("JWT_PS256_bench", ctokens.AlgorithmPS256, ctokens.WithSigningKey(ps256Priv), ctokens.WithVerifyingKey(&ps256Priv.PublicKey)),
		},
		{
			name:   "JWT_ES256",
			method: ctokens.NewMethod("JWT_ES256_bench", ctokens.AlgorithmES256, ctokens.WithSigningKey(es256Priv), ctokens.WithVerifyingKey(&es256Priv.PublicKey)),
		},
		{
			name:   "JWT_ES384",
			method: ctokens.NewMethod("JWT_ES384_bench", ctokens.AlgorithmES384, ctokens.WithSigningKey(es384Priv), ctokens.WithVerifyingKey(&es384Priv.PublicKey)),
		},
		{
			name:   "JWT_ES512",
			method: ctokens.NewMethod("JWT_ES512_bench", ctokens.AlgorithmES512, ctokens.WithSigningKey(es512Priv), ctokens.WithVerifyingKey(&es512Priv.PublicKey)),
		},
		{
			name:   "JWT_EdDSA",
			method: ctokens.NewMethod("JWT_EdDSA_bench", ctokens.AlgorithmEdDSA, ctokens.WithSigningKey(edPriv), ctokens.WithVerifyingKey(edPub)),
		},
		{
			name:   "OPAQUE_AES_256_GCM",
			method: ctokens.NewMethod("OPAQUE_AES_256_GCM_bench", ctokens.AlgorithmOpaqueAESGCM, ctokens.WithKey(aesKey)),
		},
		{
			name:   "OPAQUE_XCHACHA20_POLY1305",
			method: ctokens.NewMethod("OPAQUE_XCHACHA20_POLY1305_bench", ctokens.AlgorithmOpaqueXChaCha20Poly1305, ctokens.WithKey(xchachaKey)),
		},
	}
}

// BenchmarkGenerate measures Method.Generate across every predefined
// algorithm at the default 24 h timeout.
func BenchmarkGenerate(b *testing.B) {
	cases := buildTokenCases(b)

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = c.method.Generate(benchmarkSubject, benchmarkPayload)
			}
		})
	}
}

// BenchmarkValidate measures Method.Validate across every predefined
// algorithm using a single pre-issued token per row.
func BenchmarkValidate(b *testing.B) {
	cases := buildTokenCases(b)

	for _, c := range cases {
		token, err := c.method.Generate(benchmarkSubject, benchmarkPayload)
		if err != nil {
			b.Fatalf("Generate(%s) failed: %v", c.name, err)
		}

		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = c.method.Validate(token)
			}
		})
	}
}
