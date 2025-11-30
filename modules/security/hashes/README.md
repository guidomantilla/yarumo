# `hashes` Package  
*A lightweight, consistent abstraction for cryptographic hash functions in Go.*

---

## Overview

The `hashes` package provides a unified interface on top of Go’s standard `crypto.Hash` primitives.  
Its goal is to:

- Simplify hashing operations.
- Provide a clean, predictable API for digest computation.
- Enforce hash availability checks.
- Avoid API repetition across higher-level cryptographic modules (ECDSA, RSA-PSS, HMAC, Ed25519 workflows, etc.).

This package **does not implement new hash algorithms**.  
Instead, it wraps the existing Go hash functions in a safer, easier-to-use API.

---

## Features

- Simple `Hash(hash, data)` function for computing digests.
- Strict availability checks via `crypto.Hash.Available()`.
- No surprises: always returns raw digest bytes.
- Works seamlessly with:
  - ECDSA signing and verification
  - RSA-PSS signing and verification
  - HMAC computation and validation
  - Any custom cryptographic module that expects raw hash output

---

## Supported Hash Algorithms

Any hash registered in Go’s `crypto` package is supported, including:

- **SHA-256**
- **SHA-512**
- **SHA3-256**
- **SHA3-512**
- **BLAKE2b-256**
- **BLAKE2b-512**
- (and others registered via `crypto.RegisterHash`)

The package only requires that the provided hash is marked `Available()`.

---

## API

### `func Hash(hash crypto.Hash, data types.Bytes) types.Bytes`

Computes the digest of the given data using the specified hash function.

#### Parameters

- `hash` — A valid `crypto.Hash` identifier (e.g. `crypto.SHA256`).
- `data` — Raw message input.

#### Behavior

- Asserts that the hash is available (panic on unsupported hashes).
- Creates a new hash instance via `hash.New()`.
- Writes the input data.
- Returns the digest via `h.Sum(nil)`.

#### Returns

A `[]byte` representing the cryptographic hash of `data`.

#### Notes

- Hashing always succeeds for standard hash functions.
- Write errors are safely ignored (they never occur for `hash.Hash`).
- If you want streaming hashing, call `hash.New()` directly.

---

## Usage Examples

### Hashing data

```go
import (
    "crypto"
    "github.com/guidomantilla/yarumo/security/hashes"
)

digest := hashes.Hash(crypto.SHA256, []byte("hello world"))
```

### Using the output for signatures (ECDSA / RSA-PSS)

```go
h := hashes.Hash(crypto.SHA512, message)
r, s, err := ecdsa.Sign(rand.Reader, privateKey, h)
```

### Using the output for HMAC

```go
mac := hmac.New(sha256.New, secretKey)
mac.Write(message)
signature := mac.Sum(nil)
```

Or using your own HMAC wrapper:

```go
signature := Digest(crypto.SHA256, key, message)
```

---

## Design Philosophy

The package intentionally keeps hashing **stateless, explicit, and simple**:

- Avoids hiding behavior.
- Avoids managing custom hash objects.
- Offers one consistent entry point for digesting raw data.

It’s meant to be used as a utility within higher-level cryptographic modules.

---

## Error Handling

Hash availability is enforced through assertions:

- If a hash is not supported or not registered, the function panics with:

  > hash function not available. call crypto.RegisterHash(...)

This is by design. Unavailable hashes are considered **developer errors**, not runtime errors.

---

## Why This Package?

Cryptographic modules often need deterministic hashing of messages before signing, verifying, or authenticating.  
Instead of duplicating the `h := hash.New(); h.Write(); h.Sum()` pattern everywhere, this package:

- Centralizes the behavior.
- Reduces errors.
- Makes higher-level code cleaner.

---

## Related Packages

This package is typically used together with:

- **ECDSA** (`ecdsas`)
- **RSA-PSS** (`rsapss`)
- **HMAC** (`hmacs`)
- **Ed25519** (`ed25519sig`)
- Other signature or crypto modules in your security layer


