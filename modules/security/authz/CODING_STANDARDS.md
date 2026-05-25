# Coding Standards — modules/security/authz/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).
The PACKAGES.md classification is **Shape B** for the parent package
(`authz`) and **Shape B** for the `rbac/` sub-package — see the
`security/authz/` section in [`modules/PACKAGES.md`](../../PACKAGES.md).

## Layout

`modules/security/authz/` ships the contract (`Policy`, `Decision`,
`Request`, `PrincipalReader`, etc.) plus utility functions (`Allow` /
`Deny` / `Abstain`, `ChainPolicies`, `DefaultAuditHook`,
`SilentAuditHook`, `LocalIP`) and the canonical in-module RBAC
implementation under `rbac/`.

Transport adapters live in their own top-level modules under
`modules/extensions/security/authz/`. This keeps
`google.golang.org/grpc` out of the `go.mod` graph of any consumer
that does not import the gRPC interceptor — sub-package isolation
inside a single module still leaves heavy deps in the consumer's
`go.sum` via MVS, so true isolation requires separate `go.mod`
boundaries.

| Module | External deps |
|---|---|
| `modules/security/authz/` (this module) | only `common` |
| `modules/extensions/security/authz/http/` | `net/http` (stdlib) |
| `modules/extensions/security/authz/grpc/` | `google.golang.org/grpc` |

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

The Require middleware in each transport module does NOT use a
hardcoded context key. It takes a `PrincipalReader` option
(`WithPrincipalReader(reader)`) which the consumer wires once at
startup. The reader is typically a closure over the authn-specific
ctx key, but `authz` declares it as an interface so any authn library —
including custom ones — can satisfy it.

This makes the coupling explicit (no implicit shared ctx key) and
lets a single binary stack multiple authn implementations on top of
the same `Require` middleware family.

### Fail closed at construction time

`RequireHTTP` / `RequireUnary` / `RequireStream` (in the extension
modules) panic at construction when policy is nil or action is empty.
The rationale is the same as the `lifecycle.Build` family: a wiring
bug is loud at boot, not silent at first request.

### Default audit hook = log; opt-out = SilentAuditHook

Following the same precedent as `messaging`'s `DefaultErrorHandler`,
`NewOptions` in each transport module installs `authz.DefaultAuditHook`
(logs every Decision via `common/log`) when the caller does not pass
`WithAuditHook`. Callers who want silence must opt out explicitly with
`WithAuditHook(authz.SilentAuditHook)`.

This makes the audit obligation defensive: forgetting to wire the
hook surfaces every Decision in the standard log stream, which is
preferable to silently passing every grant unobserved.

### Closes ticket YA-0164 by encoding the audit hook in the Policy evaluation flow, not as a retrofit

Per ticket YA-0164's framing, the Policy/Require contract bakes the
audit hook into evaluation NOW so future policies do not have to be
retrofitted. Concretely: the audit hook is invoked inside Require
(both HTTP and gRPC) AFTER `policy.Evaluate` and BEFORE the response
is written, with the exact Request and Decision values. Custom
policies can also call their own audit hook from inside Evaluate — the
shape is the same.

## File layout (root `authz` package)

| File | Contents |
|---|---|
| `types.go` | Package doc + Fn aliases (`NewRequestFn`, `AllowFn`, `DenyFn`, `AbstainFn`, `ChainPoliciesFn`, `AuditHookFn`, `LocalIPFn`, `ErrAuthzFn`) + compliance vars + `Effect` const + `Decision` / `Resource` / `Environment` / `Request` structs + `Policy` interface + `PrincipalReader` interface + `PrincipalReaderFn` adapter. |
| `functions.go` | `NewRequest`, `Allow` / `Deny` / `Abstain`, `ChainPolicies` (+ private `chain` impl), `DefaultAuditHook`, `SilentAuditHook`, `LocalIP`. |
| `errors.go` | `AuthzType` const + `Error` struct + sentinels (`ErrAuthzFailed`, `ErrDenied`, `ErrAbstained`, `ErrPolicyNil`, `ErrPrincipalNil`, `ErrPrincipalReaderNil`, `ErrActionEmpty`) + `ErrAuthz` factory. |

No standalone `doc.go` — package doc lives in `types.go` per the
workspace standard.

## rbac/ sub-package

`rbac/` is the canonical Policy implementation for role-based access
control. Lives as a sub-package (not a separate go-module) because it
has no external dependencies beyond the contract itself and the
workspace's `common` module — there is no MVS leak to worry about. It
also carries enough domain-specific vocabulary (`RolesStore`,
`PrincipalIDResolverFn`, `WithRolePermissions`, `WithInheritance`,
inheritance-cycle sentinels, etc.) to warrant its own namespace; the
fusion test ("does this have vocabulary distinct from the parent
contract?") is satisfied.

| File | Contents |
|---|---|
| `types.go` | Package doc + `PrincipalIDResolverFn` + `RolesStore` interface + compliance var. |
| `errors.go` | RBAC domain errors. |
| `options.go` | `Option` / `Options` / `NewOptions` + `With*` funcs (role permissions, inheritance, store, principal id resolver, audit hook). |
| `policy.go` | `policy` struct + `NewPolicy` + Evaluate + wildcard matcher. |
| `store.go` | `InMemoryRolesStore` struct (exposed, not interface-returning, because consumers need direct access to `Assign` / `Unassign`). |
| `internals.go` | `buildClosure` + `dfsClosure` private helpers. |
