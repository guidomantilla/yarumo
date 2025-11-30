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

- Validate inputs: returns an error if `method` or `key` are nil.
- Check compatibility: if the key's curve does not match the method's curve, return an error.
- Hashes `data` using the hash defined by `method` (e.g., SHA-256 or SHA-512).
- For RS format:
  - Uses `ecdsa.Sign` to obtain `(r, s)`.
  - Encodes `r` and `s` as big-endian, fixed-size `keySize` each, concatenated as `r||s`.
- For ASN1 format:
  - Uses `ecdsa.SignASN1` to produce a standard DER signature.
- Returns an error if the requested format is not supported.


#### Returns

- `[]byte` - The signature bytes.
- `error` - An error if one occurred.

#### Notes

- RS produces fixed-size signatures: `2*keySize` bytes (see table below).
- ASN1 produces variable size (DER), suitable for X.509/TLS.
- Does not implement new hashes; it relies on Go's `crypto`.
- Common errors include: invalid method/key, incompatible curve, and unsupported format.



### `func Verify(method *Method, key *ecdsa.PublicKey, signature []byte, data []byte, format Format) (bool, error)`

Verifies an ECDSA signature.

#### Parameters

- `method` - The signature method to use.
- `key` - The public key to verify with.
- `signature` - The signature bytes to verify.
- `data` - The data that was signed.
- `format` - The signature format to use.

#### Behavior

- Validate inputs: return an error if `method` or `key` are nil.
- Check curve compatibility between `key` and `method`.
- Hash `data` using the method's hash.
- For RS:
  - Enforce exact length `2*keySize`.
  - Split `r` and `s` and call `ecdsa.Verify`.
- For ASN1:
  - Call `ecdsa.VerifyASN1` with the DER bytes.
- Distinguish between invalid format (error) and incorrect signature (return `false, nil`).


#### Returns

- `bool` - Whether the signature is valid.
- `error` - An error if one occurred.

#### Notes

- If the signature does not correspond to the message/key, return `(false, nil)`.
- If the format is incorrect or incompatible, return `(false, err)`.
- Make sure to use the same `method` for signing and verification (same curve and hash).

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

- Unified and safe API on top of Go's `crypto/ecdsa`.
- Explicit handling of signature formats (RS and ASN1), avoiding interoperability pitfalls.
- Strict curve–method compatibility to prevent unsafe usage.
- Fixed-size RS encoding for integrations with JWT/JOSE, WebAuthn, and FIDO2.
- Well-classified errors and clear flow: distinguishes invalid format from verification failure.
- Eases testing and maintenance with a single method abstraction (curve, hash, size).

---

## Related Packages

This package is typically used together with:

- **HASH** (`hashes`)
- **RSA-PSS** (`rsapss`)
- **HMAC** (`hmacs`)
- **Ed25519** (`ed25519sig`)

