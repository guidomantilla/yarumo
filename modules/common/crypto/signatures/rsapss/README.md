# `rsapss` Package
An opinionated, safe abstraction for RSA-PSS digital signatures in Go.

---

## Overview

The `rsapss` package exposes a small, consistent API to generate RSA keys and to sign/verify messages using RSA-PSS (as defined in RFC 8017). It is built on top of Go’s `crypto/rsa` and leverages the centralized hashing helpers from `modules/security/hashes`.

It focuses on:
- Clear API: GenerateKey, Sign, Verify (method and free-function forms)
- Safe defaults: PSS salt length equals hash size, vetted hash functions
- Strict validation: method, key size allowances, and error taxonomy
- Consistent wrapping of errors for better observability

This package does not implement new hash algorithms; it uses Go’s standard hashes via the `hashes` package.

---

## Features

- Method descriptor covering: name, hash function, salt length, allowed RSA key sizes
- Secure key generation through `rsa.GenerateKey`
- RSA-PSS signing and verification with explicit salt length and hash settings
- Input validation and descriptive, typed errors
- Small name-based registry (Register/Get/Supported) for extension by name

---

## Supported Methods

Two predefined methods are provided:

- `RSASSA_PSS_SHA256` (allowed key sizes: 2048, 3072, 4096 bits; salt length = hash length)
- `RSASSA_PSS_SHA512` (allowed key sizes: 3072, 4096 bits; salt length = hash length)

Each method defines:
- Hash: `crypto.SHA256` or `crypto.SHA512`
- SaltLength: `rsa.PSSSaltLengthEqualsHash`
- AllowedKeySizes: a whitelist of modulus sizes (in bits)

---

## API

### Types

```go
type Method struct {
    name            string
    kind            crypto.Hash
    saltLength      int
    allowedKeySizes []int
}

func NewMethod(name string, kind crypto.Hash, saltLength int, allowedKeySizes ...int) *Method
func (m *Method) Name() string
```

Predefined:

```go
var (
    RSASSA_PSS_SHA256 = NewMethod("RSASSA_PSS_SHA256", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, 2048, 3072, 4096)
    RSASSA_PSS_SHA512 = NewMethod("RSASSA_PSS_SHA512", crypto.SHA512, rsa.PSSSaltLengthEqualsHash, 3072, 4096)
)
```

### Key Generation

```go
func (m *Method) GenerateKey(size int) (*rsa.PrivateKey, error)
```

Generates a new RSA private key using a CSPRNG.

Behavior:
- Validates that `m` is not nil.
- Enforces that `size` is included in `m.allowedKeySizes`.
- Returns a private key for the requested modulus size.

Errors:
- Wraps the root cause with `ErrKeyGeneration(...)`.
- Returns `ErrKeySizeNotAllowed` if `size` is not whitelisted by the method.

### Signing

Two equivalent entry points are provided:

```go
func (m *Method) Sign(key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error)
func Sign(method *Method, key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error)
```

Parameters:
- `method`: must not be nil.
- `key`: RSA private key; must not be nil and its modulus bit length must be in `method.allowedKeySizes`.
- `data`: message to sign (raw message; the method applies hashing internally).

Behavior:
- Validates inputs and allowed key size.
- Computes the digest via `hashes.Hash(method.kind, data)`.
- Calls `rsa.SignPSS` with salt length `method.saltLength` and hash `method.kind`.

Returns:
- `[]byte` RSA-PSS signature.
- Error if inputs are invalid or signing fails (`ErrSignFailed` wrapped inside the free function; wrapped with `ErrSigning(...)` in the method receiver variant).

### Verification

Two equivalent entry points are provided:

```go
func (m *Method) Verify(key *rsa.PublicKey, signature, data types.Bytes) (bool, error)
func Verify(method *Method, key *rsa.PublicKey, signature, data types.Bytes) (bool, error)
```

Parameters:
- `method`: must not be nil.
- `key`: RSA public key; must not be nil and its modulus bit length must be in `method.allowedKeySizes`.
- `signature`: RSA-PSS signature to verify.
- `data`: original signed message.

Behavior:
- Validates inputs and allowed key size.
- Computes the digest via `hashes.Hash(method.kind, data)`.
- Calls `rsa.VerifyPSS` with the method’s salt length and hash.
- Distinguishes invalid signature from other errors:
  - `(false, nil)` for an invalid signature (`rsa.ErrVerification`).
  - `(false, err)` for other verification errors (wrapped with `ErrVerifyFailed` in the free function; wrapped with `ErrVerification(...)` in the method receiver variant).

Returns:
- `true, nil` if signature is valid.
- `false, nil` if signature is invalid.
- `false, err` if method/key are invalid or there is an internal verification error.

---

## Registry (extensions)

Simple helpers to work by method name:

```go
func Register(method Method)
func Get(name string) (*Method, error)
func Supported() []Method
```

Errors:
- `ErrAlgorithmNotSupported(name string)` when a method name is unknown.

The registry is pre-populated with `RSASSA_PSS_SHA256` and `RSASSA_PSS_SHA512`.

---

## Usage Example

```go
package main

import (
    "crypto/rand"
    "crypto/rsa"

    pss "github.com/guidomantilla/yarumo/modules/security/signatures/rsapss"
)

func main() {
    method := pss.RSASSA_PSS_SHA256

    // Generate a 2048-bit RSA key (allowed by RSASSA_PSS_SHA256)
    priv, _ := rsa.GenerateKey(rand.Reader, 2048)
    pub := &priv.PublicKey

    data := []byte("message to sign")

    // Sign
    sig, _ := pss.Sign(method, priv, data)

    // Verify
    ok, _ := pss.Verify(method, pub, sig, data)
    _ = ok // use result
}
```

Notes:
- For production code, always check returned errors.
- You can also call `method.GenerateKey(size)` to generate a key with size restrictions enforced by the method.

---

## Why This Package?

- Provides a unified, safe API on top of `crypto/rsa` specifically for PSS.
- Enforces modern best practices (RSA-PSS with salt length equal to hash length).
- Prevents misuse through explicit key-size whitelisting per method.
- Clear error taxonomy and wrapping for observability.

---

## Errors

Common base errors:
- `ErrMethodIsNil`
- `ErrKeyIsNil`
- `ErrKeyLengthIsInvalid`
- `ErrSignFailed`
- `ErrVerifyFailed`
- `ErrKeySizeNotAllowed`

Wrappers:
- `ErrKeyGeneration(errs ...error)`
- `ErrSigning(errs ...error)`
- `ErrVerification(errs ...error)`

Registry error:
- `ErrAlgorithmNotSupported(name string)`

These produce a typed `rsapss.Error` with context (`Type = rsa_pss_method`).

---

## Related Packages

Typically used with:
- HASH (`hashes`)
- ECDSA (`ecdsas`)
- Ed25519 (`ed25519`)
- HMAC (`hmacs`)
