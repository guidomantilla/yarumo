# Coding Standards — modules/extensions/security/authz/grpc/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md)
and the parent contract conventions in
[`modules/security/authz/CODING_STANDARDS.md`](../../../../security/authz/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `_ RequireUnaryFn = RequireUnary`, `_ RequireStreamFn = RequireStream`, `_ GRPCResourceResolverFn = (GRPCResourceResolverFn)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `grpc.UnaryServerInterceptor` / `grpc.StreamServerInterceptor`; helpers (`readPrincipal`, `buildRequest`, `methodAttr`, `peerIP`, `denyStatus`) are private |
| 4 | Constructor returns interface | Yes | `RequireUnary(policy, action, opts...) grpc.UnaryServerInterceptor`, `RequireStream(policy, action, opts...) grpc.StreamServerInterceptor` |
| 5 | Options | Yes | `Options` + `WithPrincipalReader`, `WithAuditHook`, `WithResourceResolver`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | Callers construct interceptors directly via `RequireUnary` / `RequireStream`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Thin adapter on security/authz

This module is a **thin adapter**: it accepts a pre-configured
`authz.Policy` and evaluates it once per inbound RPC (unary or stream).
The Decision contract (Allow/Deny/Abstain), the Request envelope
(Principal/Resource/Action/Environment), the PrincipalReader interface,
and the audit hook all live in `modules/security/authz/`. This module
owns only:

- Reading the Principal from `ctx` via the caller-provided
  `PrincipalReader` (denying when missing).
- Resolving the `authz.Resource` from the gRPC `FullMethod` + typed
  request message via an optional `GRPCResourceResolverFn`.
- Populating `Environment.IP` from the gRPC peer and stashing the
  `grpc_method` attribute into `Environment.Attrs`.
- Translating Deny / Abstain into `codes.PermissionDenied` with
  `Decision.Reason` as the status message.

### Override: Sibling to `extensions/security/authz/http/`

Same module layout, same shape: a Require* constructor per transport,
HTTP-specific or gRPC-specific options, the same PrincipalReader /
AuditHook / ResourceResolver pattern. The contract lives in
`security/authz`; each transport adapter is its own Go module so a
consumer of HTTP-only code never pulls `google.golang.org/grpc`.

### Override: Fail closed at construction time

Per the parent module's convention, `RequireUnary` and `RequireStream`
panic at construction when `policy` is `nil` or `action` is empty. The
rationale matches the `lifecycle.Build` family: a wiring bug is loud at
boot, not silent at first RPC.

### Override: WithResourceResolver naming

The Option is `WithResourceResolver`, not `WithGRPCResourceResolver`,
because the module name (`extensions/security/authz/grpc`) already
conveys the transport. The resolver function type
(`GRPCResourceResolverFn`) keeps the `GRPC` prefix to disambiguate from
the HTTP resolver type living in the sibling module.
