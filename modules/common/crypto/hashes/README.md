# `hashes` Package
An opinionated, safe abstraction for cryptographic hash functions in Go.

---

## Overview

The `hashes` package provides a unified interface on top of Go’s standard `crypto.Hash` primitives.
Its goals are to:

- Simplify hashing operations.
- Provide a clean, predictable API for digest computation.
- Enforce hash availability checks.
- Avoid API repetition across higher-level cryptographic modules (ECDSA, RSA-PSS, HMAC, Ed25519 workflows, etc.).

This package does not implement new hash algorithms. Instead, it wraps existing Go hash functions in a safe, easy-to-use API.

---

## Features

- Simple `Hash(hash, data)` function for computing digests.
- Method descriptor with convenience helpers and predefined methods.
- Strict availability checks via `crypto.Hash.Available()`.
- No surprises: always returns raw digest bytes.
- Works seamlessly with:
  - ECDSA signing and verification
  - RSA-PSS signing and verification
  - HMAC computation and validation
  - Any custom cryptographic module that expects raw hash output

---

## Supported Hash Algorithms

Any hash registered in Go’s `crypto` package is supported, including the predefined methods in this package:

- SHA-256 (`crypto.SHA256`)
- SHA-512 (`crypto.SHA512`)
- SHA3-256 (`crypto.SHA3_256`)
- SHA3-512 (`crypto.SHA3_512`)
- BLAKE2b-256 (`crypto.BLAKE2b_256`)
- BLAKE2b-512 (`crypto.BLAKE2b_512`)

The only requirement is that the provided hash is marked `Available()`.

---

## API

### Types

```go
type Method struct {
    name string
    kind crypto.Hash
}

func NewMethod(name string, kind crypto.Hash) *Method
func (m *Method) Name() string
func (m *Method) Hash(data types.Bytes) types.Bytes
```

Predefined methods:

```go
var (
    SHA256      = NewMethod("SHA256", crypto.SHA256)
    SHA512      = NewMethod("SHA512", crypto.SHA512)
    SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256)
    SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512)
    BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256)
    BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512)
)
```

### Hash function

```go
func Hash(hash crypto.Hash, data types.Bytes) types.Bytes
```

Computes the digest of the given data using the specified hash function.

Parameters:
- `hash`: a `crypto.Hash` identifier (e.g., `crypto.SHA256`). Must be available; otherwise it panics via assertion.
- `data`: the input to hash.

Behavior:
- Ensures the hash function is available.
- Creates a new hash instance via `hash.New()`.
- Writes the input data and returns `h.Sum(nil)`.

Notes:
- Write errors are ignored for standard `hash.Hash` implementations.
- For streaming, call `hash.New()` directly and manage the hasher yourself.

### Registry extensions

Besides direct usage of `crypto.Hash`, this package includes a tiny registry to work by name:

```go
func Register(method Method)
func Get(name string) (*Method, error)
func Supported() []Method
```

Errors:
- `ErrAlgorithmNotSupported(name string)` when a method name is unknown.

---

## Usage Examples

### Hashing data (function form)

```go
package main

import (
    "crypto"
    "fmt"

    "github.com/guidomantilla/yarumo/modules/security/hashes"
)

func main() {
    digest := hashes.Hash(crypto.SHA256, []byte("hello world"))
    fmt.Println(len(digest)) // prints 32
}
```

### Hashing data (method form)

```go
package main

import (
    "fmt"

    "github.com/guidomantilla/yarumo/modules/security/hashes"
)

func main() {
    digest := hashes.SHA512.Hash([]byte("hello world"))
    fmt.Println(len(digest)) // prints 64
}
```

### Using the output for signatures (ECDSA / RSA-PSS)

```go
package main

import (
    "crypto"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "fmt"

    h "github.com/guidomantilla/yarumo/modules/security/hashes"
)

func main() {
    // ECDSA over P-256
    priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    msg := []byte("message")
    d := h.Hash(crypto.SHA256, msg)
    r, s, _ := ecdsa.Sign(rand.Reader, priv, d)
    _ = r; _ = s
    fmt.Println("signed")
}
```

### Using the output for HMAC (vanilla example)

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "fmt"
)

func main() {
    mac := hmac.New(sha256.New, []byte("secret"))
    mac.Write([]byte("message"))
    signature := mac.Sum(nil)
    fmt.Println(len(signature))
}
```

---

## Design Philosophy

The package intentionally keeps hashing stateless, explicit, and simple:

- Avoids hiding behavior.
- Avoids managing custom hash objects.
- Offers one consistent entry point for digesting raw data.

It is meant to be used as a utility within higher-level cryptographic modules.

---

## Error Handling

Hash availability is enforced through assertions:

- If a hash is not supported or not registered, the function panics with:

  > hash function not available. call crypto.RegisterHash(...)

This is by design. Unavailable hashes are considered developer errors, not runtime errors.

Additionally, the name-based registry helpers may return a typed error:

- `ErrAlgorithmNotSupported(name)` producing a `hashes.Error` with `Type = hash_function_not_found`.

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

- ECDSA (`ecdsas`)
- RSA-PSS (`rsapss`)
- HMAC (`hmacs`)
- Ed25519 (`ed25519`)
- Other signature or crypto modules in your security layer


