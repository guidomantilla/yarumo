# `ecdsas` Package
An opinionated, safe abstraction for ECDSA digital signatures in Go.

---

## Overview

The `ecdsas` package provides a unified and simple API to generate keys, sign and verify data using ECDSA over NIST curves, built on top of Go's `crypto/ecdsa`.

It focuses on:
- Clear and minimal API (GenerateKey, Sign, Verify)
- Explicit separation of signature formats (RS and ASN1/DER)
- Strict input validation (method/key/format/sizes)
- Consistent error types and wrapping

This package does not implement new hash functions; it relies on Go's primitives and applies the hash specified by the method before signing/verifying.

---

## Features

- Simple method descriptor: Method with name, hash, curve, and size
- Secure key generation via `ecdsa.GenerateKey`
- Signature format support for RS (fixed r||s) and ASN.1 DER
- Validations and descriptive errors

---

## Supported Methods

Two predefined methods are provided as examples:

- ECDSA_with_SHA256_over_P256
- ECDSA_with_SHA512_over_P521

Each method defines:
- Curve: elliptic curve (e.g., P-256, P-521)
- Hash: hash function to apply (e.g., SHA-256, SHA-512)
- keySize: base size for RS serialization (r and s as fixed-size big-endian)

---

## Signature Formats

- RS (r || s): two fixed-size segments, typically used in JOSE/JWT, WebAuthn, FIDO2
- ASN1 (DER): standard DER sequence with INTEGER r and INTEGER s, used in X.509/TLS/OpenSSL

RS details:
- Layout: [r (keySize)] [s (keySize)] in big-endian, left-padded with zeros if necessary
- Size: 2*keySize

---

## API

### Types

```go
type Method struct {
    name    string
    kind    crypto.Hash
    keySize int
    curve   elliptic.Curve
}

func NewMethod(name string, kind crypto.Hash, keySize int, curve elliptic.Curve) *Method
func (m *Method) Name() string
```

Predefined:

```go
var (
    ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, 32, elliptic.P256())
    ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, 66, elliptic.P521())
)
```

### Key Generation

```go
func (m *Method) GenerateKey() (*ecdsa.PrivateKey, error)
```

Generates a new ECDSA private key using a CSPRNG. The curve used is the one defined by the method.

Behavior:
- Validates that m is not nil
- Returns a private key whose PublicKey matches m.curve

Errors:
- Wraps the root cause with ErrKeyGeneration(...)

### Signing

Two equivalent entry points are provided:

```go
func (m *Method) Sign(key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error)
func Sign(method *Method, key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error)
```

Parameters:
- method: must not be nil
- key: ECDSA private key; must not be nil and must use the same curve as method.curve
- data: message to sign
- format: RS or ASN1

Behavior:
- Validates inputs and curve compatibility
- Applies the hash indicated by the method to data
- RS: uses ecdsa.Sign and serializes r||s in fixed-size big-endian (2*keySize)
- ASN1: uses ecdsa.SignASN1 to obtain standard DER

Returns:
- []byte signature in the requested format
- error if inputs/formats are invalid or if signing fails (wrapped with ErrSigning(...) in the receiver variant)

### Verification

Two equivalent entry points are provided:

```go
func (m *Method) Verify(key *ecdsa.PublicKey, signature, data types.Bytes, format Format) (bool, error)
func Verify(method *Method, key *ecdsa.PublicKey, signature, data types.Bytes, format Format) (bool, error)
```

Parameters:
- method: must not be nil
- key: ECDSA public key; must not be nil and must use the same curve as method.curve
- signature: RS or ASN1 signature
- data: original message
- format: RS or ASN1

Behavior:
- Validates inputs and curve compatibility
- Applies the hash indicated by the method to data
- RS: enforces exact length 2*keySize, splits r and s, and calls ecdsa.Verify
- ASN1: calls ecdsa.VerifyASN1 with the DER bytes
- Distinguishes incorrect signature from invalid input: (false, nil) for invalid signature; (false, err) for invalid inputs/format

Returns:
- true, nil if the signature is valid
- false, nil if the signature is invalid
- false, err if the format/sizes are invalid (wrapped with ErrVerification(...) in the receiver variant)

---

## Sizes (RS)

- P-256: keySize = 32 => RS signature size 64 bytes
- P-521: keySize = 66 => RS signature size 132 bytes

ASN1 is DER-encoded and its size can vary.

---

## Usage Example

```go
package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"

    sig "github.com/guidomantilla/yarumo/modules/security/signatures/ecdsas"
)

func main() {
    method := sig.ECDSA_with_SHA256_over_P256

    priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    pub := &priv.PublicKey

    data := []byte("hello")

    // Sign
    sigBytes, _ := sig.Sign(method, priv, data, sig.RS)

    // Verify
    ok, _ := sig.Verify(method, pub, sigBytes, data, sig.RS)
    _ = ok
}
```

---

## Why This Package?

- Unified and safe API on top of `crypto/ecdsa`
- Explicit handling of signature formats (RS and ASN1) to avoid interoperability pitfalls
- Strict curve–method compatibility to prevent unsafe usages
- Fixed-size RS useful for JWT/JOSE, WebAuthn, and FIDO2
- Clear error taxonomy with wrapping for better observability

---

## Errors

Common base errors:
- ErrMethodIsNil
- ErrKeyIsNil
- ErrKeyCurveIsInvalid
- ErrSignatureInvalid
- ErrSignFailed
- ErrFormatUnsupported

Wrappers:
- ErrKeyGeneration(errs ...error)
- ErrSigning(errs ...error)
- ErrVerification(errs ...error)

These produce a typed `ecdsas.Error` with context (Type = ecdsa_method). `ErrAlgorithmNotSupported(name string)` also exists for extension usages.

---

## Related Packages

Commonly integrates with:
- HASH (`hashes`)
- RSA-PSS (`rsapss`)
- HMAC (`hmacs`)
- Ed25519 (`ed25519`)

