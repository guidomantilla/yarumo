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

| Subpackage | Purpose | External deps |
|---|---|---|
| `authn/jwt/` | JWT-backed Authenticator over `modules/crypto/tokens`. | `crypto/tokens` (→ `golang-jwt/v5`). |
| `authn/http/` | Server-side `net/http` Bearer middleware. | `net/http` (stdlib). |
| `authn/grpc/` | Unary + stream gRPC interceptors. | `google.golang.org/grpc`. |

Consumers that wire a non-JWT backend never pull `golang-jwt/v5` into
their build graph. Consumers that only serve gRPC never pull the HTTP
package and vice versa. This is the canonical pattern: keep the root
abstraction free of backend / transport deps.

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
