# Coding Standards — `modules/core/crypto`

The crypto module follows the **general review criteria** from
`modules/core/common/CODING_STANDARDS.md` (criteria 1–9 + "No Inline Assignments"
+ "Error Handling Pattern" + Import Aliases) with the **overrides and
additions** documented below.

## Crypto Subpackage Standard

### Inventory

The crypto module ships **14 subpackages**, grouped as follows:

| Category | Subpackages |
|----------|-------------|
| Utility (no `Method` pattern) | `random/`, `certs/`, `passwords/generator/` |
| Method-based (single algorithm family) | `hashes/`, `kdfs/`, `passwords/`, `tokens/` |
| Method-based (cipher families) | `ciphers/aead/`, `ciphers/hybrid/`, `ciphers/rsaoaep/` |
| Method-based (signer families) | `signers/ecdsas/`, `signers/ed25519/`, `signers/hmacs/`, `signers/rsassas/` |

The 11 method-based subpackages follow the **full Crypto Subpackage Standard**
(file structure, Method pattern, registry, errors, examples, text codec).

The 3 utility subpackages follow the general repo standards (`modules/core/common/CODING_STANDARDS.md`)
with one local twist documented below per package; they do **not** use the
`Method` + registry pattern, do not ship `extensions.go`, and do not implement
`encoding.TextMarshaler` / `TextUnmarshaler`.

### `passwords/generator/` placement

`passwords/generator/` is a **subpackage of `passwords/`**, intentionally. It
provides the standalone `Generator` struct used to build random passwords
(distinct from the password-hashing `Method` machinery in the parent). Both
packages are part of the canonical password workflow:

- `passwords/` — hashing / verification `Method` registry (bcrypt/scrypt/argon2/pbkdf2).
- `passwords/generator/` — `*Generator` struct that emits new password strings.

Keeping them under the same namespace makes the relationship explicit at
import sites (`cpasswords` + `cgenerator` both speak about passwords).
`passwords/generator/` is a Shape A-ish utility: it exposes a `*Generator`
struct with rich methods (criterion 4 exception 2 — "Concrete data structures
with rich method sets"), but **no** `Method` and **no** registry.

### File Structure

Each **method-based** crypto subpackage must contain exactly these files:

| File | Purpose |
|------|---------|
| `types.go` | Package doc, type compliance vars, function type definitions |
| `errors.go` | Error domain constant, `Error` struct, sentinel errors, `Err<Operation>` factories |
| `<name>.go` | `Method` struct, `NewMethod` constructor, predefined method vars, delegating methods |
| `functions.go` | Implementation functions (`key`, `sign`, `verify`, `digest`, etc.) |
| `options.go` | `Option`, `Options`, `NewOptions`, `With<Fn>` functions |
| `extensions.go` | Thread-safe registry: `Register`, `Get`, `Supported` |
| `text_codec.go` | `MarshalText` / `UnmarshalText` on `*Method` (for YAML/JSON/TOML config) |

For utility subpackages (`random/`, `certs/`, `passwords/generator/`), the
file set is reduced to the general Shape: `types.go`, `errors.go` (when the
package owns a domain error), `functions.go`, `options.go` (when an Options
pattern is involved). `<name>.go` and `extensions.go` do not apply. `certs/`
and `passwords/generator/` also do not need `text_codec.go`; `random/` has no
exported type to (un)marshal either.

#### Visibility of `functions.go`

`functions.go` may contain either **private implementation helpers** (the
common case for method-based subpackages — `hash`, `sign`, `verify`,
`digest`) **or public top-level functions** when those are the package's
primary API.

- **Method-based packages** typically expose only private helpers in
  `functions.go`; the public surface is the `Method` struct from `<name>.go`.
  Exception: `hashes/functions.go` ships **public** `Hash(...)` and
  `Compute(...)` convenience helpers because hashing is naturally function-
  oriented (a one-shot pure transform) and callers shouldn't be forced through
  the registry for the common case.
- **Utility packages** (`certs/`, `random/`, `passwords/generator/`) put all
  their public free functions in `functions.go` per the general Shape A rules.

The rule of thumb: public functions live in `functions.go` only when they
represent **the** package API (i.e. the package is function-oriented), not
when they are incidental helpers — incidental helpers live with their
consumer or in `internals.go` (per common Shape A rules).

### Override: Criterion 3 — Public Struct, not Interface

Crypto does **not** use public interfaces with private implementations.
Instead:

- `Method` is a **public struct** with **private fields**.
- There is no interface to implement — `Method` is the API.
- Pluggable behavior is achieved via function fields injected through Options.

### Override: Criterion 4 — Constructor returns `*Method`

> This is a concrete instance of Exception 3 (Pluggable struct pattern) from
> criterion 4 of the common standard.

`NewMethod` returns `*Method`, not an interface. Parameters vary by algorithm:

```go
// hashes
func NewMethod(name string, kind crypto.Hash, options ...Option) *Method

// signers (hmacs, ecdsas, rsassas)
func NewMethod(name string, kind crypto.Hash, keySize int, ..., options ...Option) *Method

// signers (ed25519)
func NewMethod(name string, options ...Option) *Method
```

All constructors must take `name` as the first parameter and `...Option` as
the last.

### Override: Criterion 6 — Registry, not Singleton

Crypto uses a **multi-instance registry** instead of the `Use` singleton
pattern:

```go
// Predefined method vars (package-level)
var SHA256 = NewMethod("SHA256", crypto.SHA256)

// Thread-safe registry in extensions.go
func Register(method Method)              // adds to registry
func Get(name string) (*Method, error)    // looks up by name
func Supported() []Method                 // lists all registered

// Registry internals
var methods = map[string]Method{ ... }
var lock = new(sync.RWMutex)
```

- `Register` must use `lock.Lock()`.
- `Get` and `Supported` must use `lock.RLock()` / `lock.RUnlock()` (read-only
  operations).

`Get(name)` returns a **snapshot** of the registered `Method`: the returned
`*Method` points to a copy taken at lookup time, so later `Register` calls do
not mutate previously returned pointers, and callers needing fresh state must
call `Get` again. This contract is documented canonically on each subpackage's
`Get` doc comment (see `extensions.go` in `hashes`, `kdfs`, `passwords`,
`tokens`, `ciphers/aead`, `ciphers/hybrid`, `ciphers/rsaoaep`, `signers/hmacs`,
`signers/ecdsas`, `signers/rsassas`, `signers/ed25519`).

### Error Pattern (crypto-specific)

Follows the general error handling pattern with one addition:

- **`ErrAlgorithmNotSupported(name string)`** — contextual factory that
  includes the algorithm name in the error message. Must return a domain
  `*Error` wrapping `TypedError` (not plain `fmt.Errorf`).
- **Type constant naming**: `<Algorithm>Method` (e.g., `HmacMethod`,
  `EcdsaMethod`, `RsassasMethod`). Exception: `hashes` uses `HashNotFound`.
- **Operation factories**: `ErrKeyGeneration`, `ErrSigning`, `ErrVerification`,
  `ErrDigest` — variadic `(errs ...error)`.

### Method Operations

Each `Method` struct delegates to pluggable function fields:

```go
func (m *Method) <Operation>(...) (..., error) {
    assert.NotNil(m, "method is nil")
    assert.NotNil(m.<fn>, "method <fn> is nil")

    result, err := m.<fn>(m, ...)
    if err != nil {
        return ..., Err<Operation>(err)
    }

    return result, nil
}
```

Common operations by category:
- **hashes**: `Hash(data) (Bytes, error)`
- **symmetric signers** (hmacs): `GenerateKey()`, `Digest(key, data)`,
  `Validate(key, digest, data)`
- **asymmetric signers** (ecdsas, ed25519, rsassas): `GenerateKey(...)`,
  `Sign(key, data, ...)`, `Verify(key, signature, data, ...)`

### Text Codec (Method-based subpackages only)

Each method-based subpackage ships a `text_codec.go` that implements
`encoding.TextMarshaler` and `encoding.TextUnmarshaler` on `*Method`:

```go
var (
    _ encoding.TextMarshaler   = (*Method)(nil)
    _ encoding.TextUnmarshaler = (*Method)(nil)
)

func (m *Method) MarshalText() ([]byte, error)
func (m *Method) UnmarshalText(data []byte) error
```

`MarshalText` returns the registry name of the method so callers can
serialise algorithm choices into YAML/JSON/TOML. `UnmarshalText` resolves the
name through the package registry via `Get` and overwrites the receiver with
the resolved method. Both methods assert the receiver is non-nil.

Utility subpackages (`certs/`, `random/`, `passwords/generator/`) do not own
a `Method` type and therefore do not ship `text_codec.go`.

### Examples Package

Each crypto subpackage (or group of related subpackages) must include an
`examples/` directory with a `main.go` that serves as a runnable
demonstration:

| File | Purpose |
|------|---------|
| `examples/main.go` | `package main` with `main()` — executable demonstration of the package API |

The examples `main.go` must demonstrate:

1. **Predefined methods** — direct use of package-level vars (e.g.,
   `hashes.SHA256`, `caead.AES_256_GCM`).
2. **Standalone functions** — calling public functions directly when
   available.
3. **Registry lookup** — using `Get(name)` to retrieve a method by name,
   including error case for unknown names.
4. **Listing supported methods** — using `Supported()` to enumerate all
   registered methods.

Organizational rules:

- Subpackages at the same level share one examples directory (e.g.,
  `signers/examples/` covers hmacs, ecdsas, ed25519, rsassas;
  `ciphers/examples/` covers aead and rsaoaep, with `ciphers/hybrid/examples/`
  on its own).
- Leaf subpackages get their own (e.g., `hashes/examples/`).
- Examples are **excluded** from `graph.go` imports (not part of the module
  API).
- Examples are **excluded** from `.testcoverage.yml` paths (no coverage
  enforcement).
- Examples are **excluded** from `forbidigo` linter in `.golangci.yml`
  (allowed to use `fmt.Print*`).

## Reviewed Packages

- [x] crypto/certs
- [x] crypto/ciphers/aead
- [x] crypto/ciphers/hybrid
- [x] crypto/ciphers/rsaoaep
- [x] crypto/hashes
- [x] crypto/kdfs
- [x] crypto/passwords
- [x] crypto/passwords/generator
- [x] crypto/random
- [x] crypto/signers/ecdsas
- [x] crypto/signers/ed25519
- [x] crypto/signers/hmacs
- [x] crypto/signers/rsassas
- [x] crypto/tokens
