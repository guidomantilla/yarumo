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
4. **Constructor returns interface (when an interface makes sense)** — when the type defines a public abstraction with multiple implementations or a genuine extension point, the constructor returns the interface, and its name follows `func New<InterfaceName>`. Canonical examples in the repo: `NewEngine`, `NewClient`, `NewServer`, `NewRepository`, `NewBaseWorker`, `NewScheduler`, `NewUID`, `NewValidator`.

   A constructor returns `*Struct` (or value `Struct`) — **not** an interface — in these cases:

   1. **Options structs.** `NewOptions` returns `*Options`. Options carry configuration data, not an abstraction; the Options pattern requires the struct exposed directly. Same for variants like `*CSROptions`, `*SelfSignedOptions`.

   2. **Concrete data structures with rich method sets.** Containers, graphs, trees, FSMs, accumulators — the type IS the data, there is no abstraction to hide. Hiding it behind an interface would be ceremony with zero polymorphism. Canonical examples: `compute/math/graph/` (`*DAG`, `*Directed`, `*Bipartite`, `*Undirected`, `*Tree`, `*MultigraphDirected`, `*MultigraphUndirected`), `compute/math/fsm/` (`*Machine`), `compute/math/stats/` (`*WindowedStats`), `compute/math/markov/` (`*Chain`). Plus crypto helpers: `*Generator` (`modules/crypto/passwords/generator/`), `*DelegatingEncoder` (`modules/crypto/passwords/`).

   3. **"Pluggable struct" pattern.** When the type achieves polymorphism via **function fields configured through Options**, the struct itself plays the role of both public type and mock point — no separate interface is needed. Different "implementations" appear as different **instances** of the same struct, configured differently. Canonical example: crypto's `*Method` (11 packages in `modules/crypto/`: `hashes`, `kdfs`, `ciphers/{aead,hybrid,rsaoaep}`, `signers/{hmacs,ecdsas,ed25519,rsassas}`, `passwords`, `tokens`). The function field (`hashFn`, `kdfFn`, `signFn`, …) is the mockability mechanism; the `With<Xxx>Fn(...)` option is how callers swap behavior. There is intentionally no `Pluggable<X>` wrapper in these packages — the struct IS the pluggable.

   4. **Wrappers over stdlib / external types.** When the constructor's job is to compose / validate inputs and hand back a type owned by another package, return that external type directly. Example: `NewPool(...) (*x509.CertPool, error)` in `modules/crypto/certs/`. This exception also covers wrappers that **extend** the stdlib type with extra methods (e.g. `Trace`, `Fatal`) and return a **package-owned struct mirror** instead of the stdlib type — declaring an interface in the wrapper package would force the parent (the one that owns the abstract contract) to import the wrapper for compliance, closing an import cycle. Canonical example: `modules/log/slog/` returns `*Logger` (own struct) so `modules/common/log/` can declare `_ Logger = (*cslog.Logger)(nil)` against its own `Logger` interface via structural typing, keeping the import flow one-way.

   In all four cases:
   * `assert.NotNil` / `assert.NotEmpty` still applies to non-variadic required parameters.
   * Struct methods still call `assert.NotNil` on the receiver at the start of the function.
   * Type compliance vars still apply where relevant.

   When **not** to use these exceptions: if the type genuinely has multiple distinct implementations (real polymorphism, real extension point), expose an interface and follow the main rule.
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
6. **Preconfigured Default Singletons** - Check if having a default singleton is necessary or adds value. Check `modules/crypto` package for examples.
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
    chashes "github.com/guidomantilla/yarumo/crypto/hashes"
    cslog   "github.com/guidomantilla/yarumo/log/slog"
)

// Bad — missing alias
import (
    "github.com/guidomantilla/yarumo/common/assert"
    "github.com/guidomantilla/yarumo/common/errs"
)
```

- The alias is `c` + the last segment of the import path (e.g., `crypto/hashes` → `chashes`).
- Blank imports (`_ "..."`) do not need aliases.

**Override**: when two packages share the same last segment (e.g., `common/random` and `crypto/random`), an explicit non-default alias is used:
- `crypto/random` → `crandom` (canonical — frequent inside `modules/crypto/*`).
- `common/random` → `cfrandom` ("common-fast-random") — reserved for the `math/rand/v2`-backed fast variant; signals non-secure at a glance.

## No Inline Assignments

Never combine assignment and condition in a single `if` statement. Always separate the assignment from the check. This rule applies to **every** form of `if init; cond`, regardless of what the init is — error returns, map lookups, type assertions, function-call results, anything. There is no test-code exception.

- **Error checks**: `if err := fn(); err != nil` — use `err := fn()` then `if err != nil`.
- **Map lookups**: `if val, ok := m[key]; ok` — use `val, ok := m[key]` then `if ok`.
- **Type assertions**: `if val, ok := x.(T); ok` — use `val, ok := x.(T)` then `if ok`.
- **Assignment-then-compare** (including tests): `if got := f(); got != x` — use `got := f()` then `if got != x`.

```go
// Bad
if err := doSomething(); err != nil { ... }
if val, ok := myMap[key]; ok { ... }
if val, ok := x.(MyType); ok { ... }
if got := node.String(); got != "42" { ... }

// Good
err := doSomething()
if err != nil { ... }

val, ok := myMap[key]
if ok { ... }

val, ok := x.(MyType)
if ok { ... }

got := node.String()
if got != "42" { ... }
```

### Enforcement

This rule is enforced by the custom `inlineassign` analyzer located in
`tools/lint/inlineassign/`. It is a `go/analysis` pass that flags any
`*ast.IfStmt` with a non-nil `Init` clause, covering the three forbidden forms
above plus any other inline assignment in an `if` statement.

Run it locally from the workspace root:

```bash
make lint-inline
```

or directly via `go vet`:

```bash
go build -o /tmp/inlineassign ./tools/lint/inlineassign/cmd/inlineassign
cd modules/common && go vet -vettool=/tmp/inlineassign ./...
```

The `lint` Makefile target depends on `lint-inline`, so `make lint` exercises
both `golangci-lint` and the inline-assignment analyzer.

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

The crypto subpackage standard previously documented here has moved to its
own module: see [`modules/crypto/CODING_STANDARDS.md`](../crypto/CODING_STANDARDS.md).
Crypto is no longer part of `modules/common`.

## `common/lifecycle/` — Lone Goroutine-Dispatching Exception

`modules/common/` follows the principle "pure libraries, no lifecycle, no
side effects, no goroutines spawned by package functions". `common/lifecycle/`
is the **single, deliberate exception**: it is the workspace's lifecycle
primitive, so it must operate at the lifecycle boundary.

Concretely, the following are allowed **only inside `common/lifecycle/`**
within `modules/common/`:

- Free functions that spawn background goroutines (`go component.Start(...)`).
- Functions that depend on `common/log` to emit boundary log lines
  (`starting up` / `stopping` / `stopped` / `failed to start` / `shutdown failed`).
- `Build*` constructors that return `(Component, CloseFn, error)` — wiring
  a Component with its start goroutine and its teardown callback.

The justification is that the `Component` interface itself lives here, so
the canonical wiring helpers belong to the same package — moving them to
a top-level module would create a circular ownership problem (the wiring
of the primitive cannot live further from the primitive than its consumers
do).

Every other package under `modules/common/` MUST remain free of:
- background goroutine dispatch,
- log calls at the boundary (a leaf may log on error paths, but not as
  part of routine flow),
- builder-shaped constructors that fire side effects.

If a feature genuinely needs lifecycle, it belongs in its own top-level
module (`modules/http/`, `modules/cron/`, `modules/grpc/`,
`modules/diagnostics/`, etc.). Code review should treat any `go ...` or
any `Build<X>` returning `(Component, CloseFn, error)` outside
`common/lifecycle/` and outside `modules/<top-level>/` as a red flag.

When extending `common/lifecycle/` itself, prefer:
- Wrapping the existing `Start`/`Stop` helpers rather than re-implementing the
  goroutine + channel pattern.
- Leaving `Component` constructors (`NewComponent`, `NewBaseComponent`) as
  pure factories with no side effects; concentrate the side effects in the
  `Build*` family so they remain auditable in one place.

## Reviewed Packages

- [x] common/assert
- [x] common/cast
- [x] common/constraints
- [x] common/diagnostics
- [x] common/errs
- [x] common/grpc
- [x] common/http
- [x] common/pointer
- [x] common/random
- [x] common/rest
- [x] common/types
- [x] common/uids
- [x] common/utils
