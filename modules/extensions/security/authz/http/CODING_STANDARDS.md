# Coding Standards — modules/extensions/security/authz/http/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ RequireHTTPFn = RequireHTTP` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `func(http.Handler) http.Handler`; closure captures `*Options` privately |
| 4 | Constructor returns function value | Yes | `RequireHTTP(policy, action, opts...) func(http.Handler) http.Handler` |
| 5 | Options | Yes | `Options` + `WithPrincipalReader`, `WithAuditHook`, `WithResourceResolver`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | Callers construct middleware instances directly via `RequireHTTP`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Thin adapter on security/authz

This module is a **thin transport adapter**: it accepts a pre-configured
`authz.Policy` and wraps each inbound HTTP request in a middleware that
calls `Policy.Evaluate`. All policy logic, error sentinels (`ErrPolicyNil`,
`ErrActionEmpty`, `ErrAuthz`), `PrincipalReader`, `AuditHookFn`, and the
`Request`/`Decision`/`Resource`/`Environment` contract live in
`modules/security/authz/`. This module owns only:

- Reading the principal from `ctx` via the configured `PrincipalReader`.
- Translating the HTTP request (path, headers, RemoteAddr,
  X-Forwarded-For) into an `authz.Request`.
- Translating Deny / Abstain `Decision` into a 403 response with
  `Decision.Reason` echoed in both a JSON envelope and the
  `X-Authz-Reason` header.

### Override: Sibling to `extensions/security/authz/grpc/`

Same module layout, same shape, same fail-closed-at-construction rule
(nil policy or empty action panics). Each of the two terminates authz
at a specific transport layer; the contract module
`modules/security/authz/` stays transport-agnostic.

### Override: Fail closed loud at construction

`RequireHTTP` panics on `policy == nil` or `action == ""`. Both are
wiring bugs that almost always indicate a misconfigured route; failing
at boot time surfaces the bug before any request lands instead of
denying every request silently.
