# `hmacs` Package
An opinionated, safe abstraction for HMAC computation and validation in Go.

---

## Overview

The `hmacs` package provides a tiny, consistent API to generate random keys,
compute HMAC digests, and validate them. It is built on top of Go’s `crypto/hmac`
and standard hash functions, while following the same conventions used by the
other security packages in this repository (ECDSA, RSA-PSS, Ed25519, Hashes).

It focuses on:
- Clear API: GenerateKey, Digest, Validate (method and free-function forms)
- Safe defaults: cryptographically secure random keys with appropriate lengths
- Strict validation of hash availability (via assertions)
- Predictable behavior and small name-based registry for extensions

This package does not implement new hash algorithms; it uses Go’s registered
hash functions under `crypto`.

---

## Features

- Method descriptor covering: name, hash function, key size
- Secure key generation via CSPRNG-sized keys
- HMAC digest computation and constant-time validation
- Input validation through assertions for hash availability
- Small name-based registry (Register/Get/Supported) for extension by name

---

## Supported Methods

Two predefined methods are provided:

- `HMAC_with_SHA256` (key size: 32 bytes)
- `HMAC_with_SHA512` (key size: 64 bytes)

Each method defines:
- Hash: `crypto.SHA256` or `crypto.SHA512`
- KeySize: recommended random key length for the method

---

## API

### Types

```go
type Method struct {
    name    string
    kind    crypto.Hash
    keySize int
}

func NewMethod(name string, kind crypto.Hash, keySize int) *Method
func (m *Method) Name() string
```

Predefined:

```go
var (
    HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, 32)
    HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, 64)
)
```

### Key Generation

```go
func (m *Method) GenerateKey() types.Bytes
```

Generates a random key using a cryptographically secure random source with the
size recommended by the method.

Behavior:
- Validates that `m` is not nil.
- Returns `[]byte` of length `m.keySize`.

### HMAC Digest

Method receiver form:

```go
func (m *Method) Digest(key types.Bytes, data types.Bytes) types.Bytes
```

Free-function form:

```go
func Digest(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes
```

Parameters:
- `hash`: a `crypto.Hash` identifier (e.g., `crypto.SHA256`). Must be available; otherwise it panics via assert.
- `key`: secret key used by HMAC.
- `data`: message to authenticate.

Behavior:
- Ensures the hash function is available.
- Initializes an HMAC with the given key and writes the data.
- Returns `h.Sum(nil)` as the HMAC value.

Returns:
- HMAC value as a byte slice.

Notes:
- Write errors are ignored for standard `hash.Hash` implementations.
- The function does not return errors; it panics only if the hash is unavailable.

### HMAC Validation

Method receiver form:

```go
func (m *Method) Validate(key types.Bytes, digest types.Bytes, data types.Bytes) bool
```

Free-function form:

```go
func Validate(hash crypto.Hash, key types.Bytes, digest types.Bytes, data types.Bytes) bool
```

Parameters:
- `hash`: a `crypto.Hash` identifier; must be available.
- `key`: secret key used to compute the HMAC.
- `digest`: expected HMAC value.
- `data`: original message.

Behavior:
- Recomputes the HMAC using `Digest(...)` and compares it with `digest`
  using `hmac.Equal` (constant-time comparison).

Returns:
- `true` if the digests match; `false` otherwise.

Notes:
- Panics only if the hash function is not registered (via assert).

---

## Registry (extensions)

Name-based helpers to work with method names:

```go
func Register(method Method)
func Get(name string) (*Method, error)
func Supported() []Method
```

Errors:
- `ErrAlgorithmNotSupported(name string)` when a method name is unknown.

The registry is pre-populated with `HMAC_with_SHA256` and `HMAC_with_SHA512`.

---

## Usage Examples

### Compute and verify an HMAC (method form)

```go
package main

import (
    "fmt"

    hmacs "github.com/guidomantilla/yarumo/modules/security/signatures/hmacs"
)

func main() {
    method := hmacs.HMAC_with_SHA256

    key := method.GenerateKey()          // 32 bytes
    data := []byte("hello world")

    mac := method.Digest(key, data)
    ok := method.Validate(key, mac, data)
    fmt.Println("valid?", ok)
}
```

### Compute and verify an HMAC (free-function form)

```go
package main

import (
    "crypto"
    "fmt"

    hmacs "github.com/guidomantilla/yarumo/modules/security/signatures/hmacs"
)

func main() {
    key := []byte("super-secret-key-32-bytes........") // example only
    data := []byte("payload")

    mac := hmacs.Digest(crypto.SHA256, key, data)
    ok := hmacs.Validate(crypto.SHA256, key, mac, data)
    fmt.Println("valid?", ok)
}
```

---

## Design Philosophy

The package keeps the HMAC API minimal and predictable:

- Avoids hiding behavior; all parameters are explicit.
- Avoids custom error types for digest/validate functions; misuse is treated as a developer error via assertions (hash availability).
- Provides a consistent interface aligned with other security modules in this repository.

Use this package as a building block wherever message authentication is required.

---

## Error Handling

HMAC digest and validation rely on hash availability enforced through assertions:

- If a hash is not supported or not registered, the functions panic with:

  > hash function not available. call crypto.RegisterHash(...)

Additionally, the name-based registry helpers may return a typed error:

- `ErrAlgorithmNotSupported(name)` producing a `hmacs.Error` with `Type = hmac_function_not_found`.

No other error wrappers are exposed by this package because digesting and validation do not return errors in the standard Go APIs.

---

## Related Packages

This package is typically used together with:

- HASH (`hashes`)
- ECDSA (`ecdsas`)
- RSA-PSS (`rsapss`)
- Ed25519 (`ed25519`)
