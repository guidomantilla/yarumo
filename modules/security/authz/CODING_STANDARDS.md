# Coding Standards — modules/security/authz/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).
The PACKAGES.md classification is **Shape B** for the parent package
(`authz`) and **Shape B** for the `rbac/` sub-package — see the
`security/authz/` section in [`modules/PACKAGES.md`](../../PACKAGES.md).

## Module-specific conventions

### Principal in Request is typed as `any`

`authz.Request.Principal` is declared as `any`. The justification is
deliberate: this module must not depend on `authn` (or any other
authentication library), so the Principal shape is left to the
consumer. Policies cast `req.Principal` to whatever type their authn
layer produces (a service-account struct, an `*authn.Principal`,
`*jwt.Claims`, an API key id, etc.).

Trade-off: policies that need to read principal fields have a runtime
type assertion. This is the price for keeping `authz` reusable across
authn shapes.

### Principal-from-ctx wired explicitly via PrincipalReader

The Require HTTP and gRPC middleware do NOT use a hardcoded context
key. They take a `PrincipalReader` option
(`WithPrincipalReader(reader)`) which the consumer wires once at
startup. The reader is typically `authn.PrincipalReaderFn` (a closure
over the authn-specific ctx key) but `authz` declares it as an
interface so any authn library — including custom ones — can satisfy
it.

This makes the coupling explicit (no implicit shared ctx key) and
lets a single binary stack multiple authn implementations on top of
the same `Require`.

### Fail closed at construction time

`RequireHTTP`, `RequireUnary`, `RequireStream` panic at construction
when policy is nil or action is empty. The rationale is the same as
the `lifecycle.Build` family: a wiring bug is loud at boot, not silent
at first request. Production code that wires these middlewares lives
in `main` or in module init, so panics are caught by `go test` and by
the boot path.

### Default audit hook = log; opt-out = SilentAuditHook

Following the same precedent as `messaging`'s `DefaultErrorHandler`,
`authz.NewOptions` installs `DefaultAuditHook` (logs every Decision
via `common/log`) when the caller does not pass `WithAuditHook`.
Callers who want silence must opt out explicitly with
`WithAuditHook(SilentAuditHook)`.

This makes the audit obligation defensive: forgetting to wire the
hook surfaces every Decision in the standard log stream, which is
preferable to silently passing every grant unobserved.

### Closes ticket YA-0164 by encoding the audit hook in the Policy
### evaluation flow, not as a retrofit

Per ticket YA-0164's framing, the Policy/Require contract bakes the
audit hook into evaluation NOW so future policies do not have to be
retrofitted. Concretely: the audit hook is invoked inside Require
(both HTTP and gRPC) AFTER `policy.Evaluate` and BEFORE the response
is written, with the exact Request and Decision values. Custom
policies can also call their own audit hook from inside Evaluate — the
shape is the same.

## File layout

| File | Contents |
|---|---|
| `doc.go` | Package doc + design notes (Principal=any, PrincipalReader explicit). |
| `types.go` | Effect / Decision / Resource / Environment / Request / Policy / PrincipalReader / PrincipalReaderFn / AuditHookFn. |
| `errors.go` | `Error` struct + sentinels + `ErrAuthz` factory. |
| `functions.go` | `NewRequest`, `Allow`/`Deny`/`Abstain`, `ChainPolicies` (+ private `chain` impl), `DefaultAuditHook`, `SilentAuditHook`, `LocalIP`. |
| `options.go` | `Option` / `Options` / `NewOptions` / `WithPrincipalReader` / `WithAuditHook` / `WithHTTPResourceResolver` / `WithGRPCResourceResolver`. |
| `require_http.go` | `RequireHTTP` middleware + helpers (`buildHTTPRequest`, `clientIP`, `writeAuthzDeny`). |
| `require_grpc.go` | `RequireUnary` / `RequireStream` interceptors + helpers (`buildGRPCRequest`, `peerIP`, `denyStatus`). |
| `graph.go` | Dependency-graph image tooling marker. |

## rbac/ sub-package

`rbac/` is the canonical Policy implementation for role-based access
control:

| File | Contents |
|---|---|
| `types.go` | Package doc + `PrincipalIDResolverFn` + `RolesStore` interface + compliance var. |
| `errors.go` | RBAC domain errors. |
| `options.go` | `Option`/`Options`/`NewOptions` + With\* funcs (role permissions, inheritance, store, principal id resolver, audit hook). |
| `policy.go` | `policy` struct + `NewPolicy` + Evaluate + wildcard matcher. |
| `store.go` | `InMemoryRolesStore` struct (exposed, not interface-returning, because consumers need direct access to `Assign`/`Unassign`). |
| `internals.go` | `buildClosure` + `dfsClosure` private helpers. |
