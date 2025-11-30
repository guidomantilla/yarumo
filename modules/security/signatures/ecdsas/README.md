# `ecdsas` Package  
*A consistent, modern abstraction for ECDSA digital signatures in Go.*

---

## Overview

The `ecdsas` package provides a unified, safe, and extensible interface for signing and verifying data using ECDSA across supported NIST curves.

It builds on Go’s `crypto/ecdsa` primitives, and offers:
- Clear separation of signature formats
- Strict curve–method compatibility
- Deterministic digest handling
- Simple unified API (`Sign` and `Verify`)
- Support for both **RS (raw r||s)** and **ASN1 (DER)** formats

This package **does not implement new hash algorithms**.
Instead, it wraps existing Go ecdsa functions in a safer, easier-to-use API.

---

## Features

---

## Supported Curves & Methods

The package works with the methods you define, such as:

- `ECDSA_with_SHA256_over_P256`
- `ECDSA_with_SHA512_over_P521`

Each method defines:
- Curve
- Hash function
- Key size (for RS-format encoding)

---

## Signature Formats

### 1. RS Format (`r || s`)
Used in:
- JWT / JOSE
- WebAuthn
- FIDO2

Output layout:
```
[r (fixed size)] [s (fixed size)]
```

### 2. ASN1 Format (DER)
Used in:
- X.509
- TLS
- OpenSSL

Encoded as:
```
SEQUENCE {
    r INTEGER
    s INTEGER
}
```

---

## API

### `func Sign(method *Method, key *ecdsa.PrivateKey, data []byte, format Format) ([]byte, error)`

Signs data using ECDSA.

#### Parameters

- `method` - The signature method to use.
- `key` - The private key to sign with.
- `data` - The data to sign.
- `format` - The signature format to use.

#### Behavior


#### Returns

- `[]byte` - The signature bytes.
- `error` - An error if one occurred.

#### Notes



### `func Verify(method *Method, key *ecdsa.PublicKey, signature []byte, data []byte, format Format) (bool, error)`

Verifies an ECDSA signature.

#### Parameters

- `method` - The signature method to use.
- `key` - The public key to verify with.
- `signature` - The signature bytes to verify.
- `data` - The data that was signed.
- `format` - The signature format to use.

#### Behavior


#### Returns

- `bool` - Whether the signature is valid.
- `error` - An error if one occurred.

#### Notes

---

## RS Format Sizes
| Curve | Key Size | RS Signature Size |
|-------|----------|--------------------|
| P-256 | 32 bytes | 64 bytes           |
| P-521 | 66 bytes | 132 bytes          |


---

## Usage Examples

### Example
```go
method := ECDSA_with_SHA256_over_P256

priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
pub := &priv.PublicKey

data := []byte("hello")

sig, _ := Sign(method, priv, data, RS)
ok, _ := Verify(method, pub, sig, data, RS)
```

---

## Why This Package?

---

## Related Packages

This package is typically used together with:

- **HASH** (`hashes`)
- **RSA-PSS** (`rsapss`)
- **HMAC** (`hmacs`)
- **Ed25519** (`ed25519sig`)

