# `ed25519` Package
An opinionated, safe abstraction for Ed25519 digital signatures in Go.

---

## Overview

The `ed25519` package provides a unified and simple API to generate keys, sign and verify data using Ed25519 (EdDSA over Curve25519) built on top of Go's `crypto/ed25519`.

It focuses on:
- Clear and minimal API (`GenerateKey`, `Sign`, `Verify`)
- Deterministic signatures (Ed25519 inherent property)
- Strict input validation (method/key/signature sizes)
- Consistent error wrapping and typed errors

This package does not implement new hash functions. Ed25519 performs hashing internally as part of the signature algorithm.

---

## Features

- Simple method descriptor: `Method` with a single predefined instance `Ed25519`.
- Safe key generation via `crypto/rand`.
- Deterministic signing (`ed25519.Sign`).
- Constant-time verification (`ed25519.Verify`).
- Input validation and descriptive errors.

---

## Supported Methods

This package currently defines a single method:

- `Ed25519`

Unlike ECDSA packages with multiple curves and signature encodings, Ed25519 defines a single scheme with a fixed-size signature and no alternative encodings (e.g., no RS or ASN.1/DER toggles).

---

## Signature Format

- Format: raw Ed25519 signature bytes
- Size: 64 bytes (`ed25519.SignatureSize`)

There is no ASN.1/DER or r||s representation to select. Ed25519 always outputs a fixed-size 64-byte signature.

---

## API

### Types

```go
type Method struct {
    name string
}

func NewMethod(name string) *Method
func (m *Method) Name() string
```

Predefined:

```go
var Ed25519 = NewMethod("Ed25519")
```

### Key Generation

```go
func (m *Method) GenerateKey() (ed25519.PrivateKey, error)
```

Generates a new Ed25519 private key using a cryptographically secure RNG.

Behavior:
- Validates `m` is not nil.
- Returns a 64-byte private key (`ed25519.PrivateKey`). The associated public key can be obtained from it (`priv.Public().(ed25519.PublicKey)` or by splitting the returned values from `crypto/ed25519.GenerateKey`).

Errors:
- Wraps the root cause with `ErrKeyGeneration(...)`.

### Signing

Two equivalent entry points are provided:

```go
func (m *Method) Sign(key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error)
func Sign(method *Method, key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error)
```

Parameters:
- `method`: must not be nil.
- `key`: pointer to an Ed25519 private key; must not be nil and must be `ed25519.PrivateKeySize` in length.
- `data`: message to sign.

Behavior:
- Validates inputs.
- Produces a deterministic Ed25519 signature of size 64 bytes.

Returns:
- `[]byte` signature of size `ed25519.SignatureSize`.
- An error if inputs are invalid (wrapped with `ErrSigning(...)` in the method receiver variant).

Notes:
- Ed25519 performs hashing internally; no external digesting is applied.

### Verification

Two equivalent entry points are provided:

```go
func (m *Method) Verify(key *ed25519.PublicKey, signature, data types.Bytes) (bool, error)
func Verify(method *Method, key *ed25519.PublicKey, signature, data types.Bytes) (bool, error)
```

Parameters:
- `method`: must not be nil.
- `key`: pointer to an Ed25519 public key; must not be nil and must be `ed25519.PublicKeySize` in length.
- `signature`: must be exactly `ed25519.SignatureSize` (64 bytes).
- `data`: original signed message.

Behavior:
- Validates inputs and sizes.
- Checks the signature using `ed25519.Verify` (constant-time).
- Distinguishes invalid signature from invalid inputs: returns `(false, nil)` for an invalid signature, and `(false, err)` for invalid inputs.

Returns:
- `true, nil` if signature is valid.
- `false, nil` if signature is invalid.
- `false, err` for invalid method/key/signature sizes (wrapped with `ErrVerification(...)` in the method receiver variant).

Notes:
- As with signing, no external hash is used; Ed25519 includes hashing.

---

## Sizes

- Public key: 32 bytes (`ed25519.PublicKeySize`)
- Private key: 64 bytes (`ed25519.PrivateKeySize`)
- Signature: 64 bytes (`ed25519.SignatureSize`)

---

## Usage Example

```go
package main

import (
    cryped "crypto/ed25519"
    "crypto/rand"

    edsig "github.com/guidomantilla/yarumo/modules/security/signers/ed25519"
)

func main() {
    // Create method
    method := edsig.Ed25519

    // Generate key
    // You can also use method.GenerateKey() to get a fresh private key
    _, priv, _ := cryped.GenerateKey(rand.Reader)
    pub := priv.Public().(cryped.PublicKey)

    data := []byte("hello")

    // Sign
    sig, _ := edsig.Sign(method, &priv, data)

    // Verify
    ok, _ := edsig.Verify(method, &pub, sig, data)
    _ = ok // use result
}
```

---

## Why This Package?

- Unified and safe API on top of Go's `crypto/ed25519`.
- Explicit size checks for keys and signatures to prevent misuse.
- Clear error taxonomy with wrapping functions for better observability.
- Deterministic signatures and constant-time verification.

---

## Errors

Common error values and wrappers:

- Base errors:
  - `ErrMethodIsNil`
  - `ErrKeyIsNil`
  - `ErrKeyLengthIsInvalid`
  - `ErrSignatureLengthInvalid`
- Wrappers:
  - `ErrKeyGeneration(errs ...error)`
  - `ErrSigning(errs ...error)`
  - `ErrVerification(errs ...error)`

These produce an `ed25519.Error` (typed error) with context.

---

## Related Packages

This package commonly integrates with:

- HASH (`hashes`)
- RSA-PSS (`rsapss`)
- HMAC (`hmacs`)
- ECDSA (`ecdsas`)
