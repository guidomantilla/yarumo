# Coding Standards — modules/security/authn/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).
Its PACKAGES.md classification is **Shape B** (root package defines the
`Authenticator` interface; concrete implementations live in
subpackages). See [`modules/PACKAGES.md`](../../PACKAGES.md) section
"Módulo `modules/security/authn/`".

## Module-specific conventions

### Subpackage isolation

The root `authn` package owns only the abstraction — `Principal`,
`Authenticator`, `WithPrincipal`, `FromContext`, the error domain. Each
transport / backend lives in its own subpackage:

| Package | Path | External deps |
|---|---|---|
| `authn` (root) | `modules/security/authn/` | `crypto/tokens` (→ `golang-jwt/v5`). Ships the `Authenticator` contract + the canonical `tokenAuthenticator` impl that works with all 15 algorithms (JWT + opaque AEAD). |
| `authn/http` | `modules/extensions/security/authn/http/` (separate module) | `net/http` (stdlib). |
| `authn/grpc` | `modules/extensions/security/authn/grpc/` (separate module) | `google.golang.org/grpc`. |

The two transport adapters live in their own top-level modules under
`modules/extensions/security/authn/`. This keeps `google.golang.org/grpc`
out of the `go.mod` graph of any consumer that does not import the gRPC
adapter — sub-package isolation inside a single module still leaves
heavy deps in the consumer's `go.sum` via MVS, so true isolation
requires separate `go.mod` boundaries.

The `tokenAuthenticator` impl stays inside `security/authn` because it
only depends on the contract itself plus `crypto/tokens` (a workspace
module). It is the canonical backend; future in-module backends would
sit beside it in the root package, not as nested subpackages.

### Failure contract

All Authenticator implementations MUST wrap failures through
`authn.ErrAuthentication(...)`. Transport middleware can then collapse
every flavor of failure into a uniform 401 / `codes.Unauthenticated`
response without inspecting concrete error types. `errors.Is(err,
authn.ErrAuthenticationFailed)` is true for every error this module
produces.

### Stateless components

The module owns no lifecycle and spawns no goroutines:
- `Principal` is an immutable value type.
- `Authenticator` is a stateless validator.
- Transport middleware are pure function composition.

If a future Authenticator wants to cache validation results or run a
background refresh loop, it must live in its own top-level module
(`modules/managed/<name>/`) and expose an `Authenticator`
implementation to wire into this contract.
