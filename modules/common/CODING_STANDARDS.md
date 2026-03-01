# Coding Standards

Conventions and standards for all packages under `modules/common`:

1. **Bullet proof review** - Read the code, find bugs, edge cases, and panics. Show options/ask to the user. Fix what is found.
2. **Type Compliance** - Make sure that all structs are compliant with the interface or type defined in the package.
   * Example: var _ MyInterface = (*mystruct)(nil)
   * Example: var _ MyFuncTypeFn = MyFunc - > Functions types as contracts
   * Exceptions: `NewOptions`, `With<Field>` functions from the Options pattern do not require function types.
   * Exceptions: Constructors do not require function types.
3. **Public Interface, Private Implementation** - Interfaces should be public, implementation private.
   * If the package exposes an interface that external consumers need to mock, provide a `Pluggable<Interface>` struct with public function fields (e.g. `PluggableClient` with `DoFn`, `LimiterEnabledFn`).
4. **Constructor returns interface** - All constructors should return an interface. The name of the constructor should be func New<InterfaceName>.
   * Use `assert.NotNil` or `assert.NotEmpty` on non-variadic parameters that must not be nil or empty.
   * Struct methods must call `assert.NotNil` on the receiver at the start of the function.
5. **Options** - the fields of the struct should be private. Check a func With<FieldName> to set the field. Each With function must validate its input using `if valid { assign }` — never use guard clauses with early `return` inside the option closure. Invalid input is silently ignored, preserving the default.
   ```go
   // Good — if valid then assign
   func WithTimeout(timeout time.Duration) Option {
       return func(opts *Options) {
           if timeout > 0 {
               opts.timeout = timeout
           }
       }
   }

   // Bad — guard clause with early return
   func WithTimeout(timeout time.Duration) Option {
       return func(opts *Options) {
           if timeout <= 0 {
               return
           }
           opts.timeout = timeout
       }
   }
   ```
6. **Preconfigured Default Singletons** - Check if having a default singleton is necessary or adds value. Check common/crypto package for examples.
   * The function that sets or selects the active default must be named `Use`.
   * If the package has a registry, `Use` selects from it by name. A package-level function (e.g. `Generate`) delegates to the current default.
   * Singleton variables must use Go MixedCaps naming (e.g. `DefaultClient`, `NoopClient`), not `SCREAMING_SNAKE_CASE`.
7. **Linter** - Run `go tool golangci-lint run --fix ./package/...` until 0 issues. Adjust `.golangci.yml` if necessary.
8. **Tests** - Rewrite tests following these rules:
   - One `Test*` per exported function
   - Subtests with `t.Run` for each case
   - `t.Parallel()` in all tests and subtests
   - `t.Fatal` / `t.Fatalf` assertions, no testify
   - No table-driven tests
   - Edge case coverage (nil, zero, empty, negative, etc.)
   - 100% statement coverage
   - When a package has multiple implementations of the same interface (e.g. `client` and `PluggableClient`), use the type name as a prefix in the test name: `TestClient_Do`, `TestPluggableClient_Do`
   - Check `.golangci.yml` in case a linter rule prevents full compliance with these test rules
9. **Documentation** - Every exported symbol must have a doc comment:
   - **Package doc**: `// Package <name> <one-line description>.` in one file per package (prefer `types.go` or the main file).
   - **Functions**: `// <FuncName> <verb>s ...` (third person singular, e.g. "returns", "checks", "creates").
   - **Function types**: `// <TypeName> is the function type for <FuncName>.`
   - **Interfaces**: `// <InterfaceName> defines the interface for ...` Each method must have its own doc comment: `// <MethodName> <verb>s ...`
   - **Structs / Types**: `// <TypeName> <description>.`
   - **Constants**: A single `// <description>.` comment above the `const (` block. Individual constants only need a comment if the group comment is not enough.
   - **Options pattern**: `// Option is a functional option for configuring <Package> Options.`, `// Options holds the configuration for ...`, `// With<Field> sets ...`
   - **Interface contracts**: Interfaces with side effects must document caller responsibilities (e.g. "the caller must close res.Body when err == nil") and concurrency guarantees (e.g. "safe for concurrent use") in the interface doc comment.

## Import Aliases

All imports of `github.com/guidomantilla/yarumo/common/<pkg>` must use the alias `c<last-segment>`:

```go
// Good
import (
    cassert "github.com/guidomantilla/yarumo/common/assert"
    cerrs   "github.com/guidomantilla/yarumo/common/errs"
    ctypes  "github.com/guidomantilla/yarumo/common/types"
    chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
    cslog   "github.com/guidomantilla/yarumo/common/log/slog"
)

// Bad — missing alias
import (
    "github.com/guidomantilla/yarumo/common/assert"
    "github.com/guidomantilla/yarumo/common/errs"
)
```

- The alias is `c` + the last segment of the import path (e.g., `common/crypto/hashes` → `chashes`).
- Blank imports (`_ "..."`) do not need aliases.

## No Inline Assignments

Never combine assignment and condition in a single `if` statement. Always separate the assignment from the check. This applies to:

- **Error checks**: `if err := fn(); err != nil` — use `err := fn()` then `if err != nil`.
- **Map lookups**: `if val, ok := m[key]; ok` — use `val, ok := m[key]` then `if ok`.
- **Type assertions**: `if val, ok := x.(T); ok` — use `val, ok := x.(T)` then `if ok`.

```go
// Bad
if err := doSomething(); err != nil { ... }
if val, ok := myMap[key]; ok { ... }
if val, ok := x.(MyType); ok { ... }

// Good
err := doSomething()
if err != nil { ... }

val, ok := myMap[key]
if ok { ... }

val, ok := x.(MyType)
if ok { ... }
```

## Error Handling Pattern

Packages that contain logic returning errors must follow the `common/errs` pattern:

1. **Define a type constant** for the error domain:
   ```go
   const RequestType = "http-request"
   ```

2. **Define a domain error struct** that embeds `errs.TypedError`:
   ```go
   var _ error = (*Error)(nil)

   type Error struct {
       errs.TypedError
   }
   ```
   Optionally override `Error()` for a custom format. `Unwrap()` and `ErrorType()` are inherited automatically.

3. **Define sentinel errors** for specific failure modes:
   ```go
   var (
       ErrHttpRequestFailed = errors.New("http request failed")
       ErrContextNil        = errors.New("context is nil")
   )
   ```

4. **Define `Err<Operation>` factory functions** that wrap errors into the domain error:
   ```go
   func ErrDo(errs ...error) error {
       return &Error{
           TypedError: cerrs.TypedError{
               Type: RequestType,
               Err:  errors.Join(append(errs, ErrHttpRequestFailed)...),
           },
       }
   }
   ```

5. **Internal helpers** — private functions that return errors must use `errs.Wrap(sentinel, err)` with a sentinel defined in `errors.go`. Never use `fmt.Errorf` or `errors.New` in business code when the package has an `errors.go`. The public API function that calls the helper is responsible for re-wrapping via the factory:
   ```go
   // errors.go — sentinel
   var ErrReadBodyFailed = errors.New("reading response body failed")

   // functions.go — internal helper
   func readBody(r io.Reader) ([]byte, error) {
       data, err := io.ReadAll(r)
       if err != nil {
           return nil, cerrs.Wrap(ErrReadBodyFailed, err)
       }
       return data, nil
   }

   // public API re-wraps with factory
   func Call(...) (..., error) {
       body, err := readBody(resp.Body)
       if err != nil {
           return nil, ErrCall(err)
       }
       ...
   }
   ```

6. **JSON serialization** — use `errs.AsErrorInfo(err)` to convert an error tree into `[]errs.ErrorInfo`, grouped by type:
   ```json
   [
     {
       "type": "http-request",
       "messages": ["connection timeout", "dial tcp failed", "http request failed"]
     }
   ]
   ```

## Crypto Subpackage Standard

Subpackages under `common/crypto` (hashes, signers/hmacs, signers/ecdsas, signers/ed25519, signers/rsapss, ciphers/aead, ciphers/rsaoaep, passwords, tokens, certs) follow the general review criteria above with these **overrides and additions**:

> **Note:** `certs` is a utility package — it provides standalone functions instead of the `Method` + registry pattern. The file structure, Method overrides, and registry sections do not apply to `certs`; it uses `types.go`, `errors.go`, `functions.go`, and `options.go` only.

### File Structure

Each crypto subpackage must contain exactly these files:

| File | Purpose |
|------|---------|
| `types.go` | Package doc, type compliance vars, function type definitions |
| `errors.go` | Error domain constant, `Error` struct, sentinel errors, `Err<Operation>` factories |
| `<name>.go` | `Method` struct, `NewMethod` constructor, predefined method vars, delegating methods |
| `functions.go` | Private implementation functions (`key`, `sign`, `verify`, `digest`, etc.) |
| `options.go` | `Option`, `Options`, `NewOptions`, `With<Fn>` functions |
| `extensions.go` | Thread-safe registry: `Register`, `Get`, `Supported` |

### Override: Criterion 3 — Public Struct, not Interface

Crypto does **not** use public interfaces with private implementations. Instead:

- `Method` is a **public struct** with **private fields**.
- There is no interface to implement — `Method` is the API.
- Pluggable behavior is achieved via function fields injected through Options.

### Override: Criterion 4 — Constructor returns `*Method`

`NewMethod` returns `*Method`, not an interface. Parameters vary by algorithm:

```go
// hashes
func NewMethod(name string, kind crypto.Hash, options ...Option) *Method

// signers (hmacs, ecdsas, rsapss)
func NewMethod(name string, kind crypto.Hash, keySize int, ..., options ...Option) *Method

// signers (ed25519)
func NewMethod(name string, options ...Option) *Method
```

All constructors must take `name` as the first parameter and `...Option` as the last.

### Override: Criterion 6 — Registry, not Singleton

Crypto uses a **multi-instance registry** instead of the `Use` singleton pattern:

```go
// Predefined method vars (package-level)
var SHA256 = NewMethod("SHA256", crypto.SHA256)

// Thread-safe registry in extension.go
func Register(method Method)              // adds to registry
func Get(name string) (*Method, error)    // looks up by name
func Supported() []Method                 // lists all registered

// Registry internals
var methods = map[string]Method{ ... }
var lock = new(sync.RWMutex)
```

- `Register` must use `lock.Lock()`.
- `Get` and `Supported` must use `lock.RLock()` / `lock.RUnlock()` (read-only operations).

### Error Pattern (crypto-specific)

Follows the general error handling pattern with one addition:

- **`ErrAlgorithmNotSupported(name string)`** — contextual factory that includes the algorithm name in the error message. Must return a domain `*Error` wrapping `TypedError` (not plain `fmt.Errorf`).
- **Type constant naming**: `<Algorithm>Method` (e.g., `HmacMethod`, `EcdsaMethod`, `RsaPssMethod`). Exception: hashes uses `HashNotFound`.
- **Operation factories**: `ErrKeyGeneration`, `ErrSigning`, `ErrVerification`, `ErrDigest` — variadic `(errs ...error)`.

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
- **hashes**: `Hash(data) Bytes`
- **symmetric signers** (hmacs): `GenerateKey()`, `Digest(key, data)`, `Validate(key, digest, data)`
- **asymmetric signers** (ecdsas, ed25519, rsapss): `GenerateKey(...)`, `Sign(key, data, ...)`, `Verify(key, signature, data, ...)`

### Examples Package

Each crypto subpackage (or group of related subpackages) must include an `examples/` directory with a `main.go` that serves as a runnable demonstration:

| File | Purpose |
|------|---------|
| `examples/main.go` | `package main` with `main()` — executable demonstration of the package API |

The examples `main.go` must demonstrate:
1. **Predefined methods** — direct use of package-level vars (e.g., `hashes.SHA256`, `caead.AES_256_GCM`).
2. **Standalone functions** — calling public functions directly when available.
3. **Registry lookup** — using `Get(name)` to retrieve a method by name, including error case for unknown names.
4. **Listing supported methods** — using `Supported()` to enumerate all registered methods.

Organizational rules:
- Subpackages at the same level share one examples directory (e.g., `signers/examples/` covers hmacs, ecdsas, ed25519, rsapss; `ciphers/examples/` covers aead, rsaoaep).
- Leaf subpackages get their own (e.g., `hashes/examples/`).
- Examples are **excluded** from `graph.go` imports (not part of the module API).
- Examples are **excluded** from `.testcoverage.yml` paths (no coverage enforcement).
- Examples are **excluded** from `forbidigo` linter in `.golangci.yml` (allowed to use `fmt.Print*`).

## Reviewed Packages

- [x] common/assert
- [x] common/cast
- [x] common/constraints
- [x] common/cron
- [x] common/crypto (hashes, signers/*, ciphers/aead, ciphers/rsaoaep, certs, passwords, tokens)
- [x] common/diagnostics
- [x] common/errs
- [x] common/grpc
- [x] common/http
- [x] common/log
- [x] common/pointer
- [x] common/random
- [x] common/rest
- [x] common/types
- [x] common/uids
- [x] common/utils
