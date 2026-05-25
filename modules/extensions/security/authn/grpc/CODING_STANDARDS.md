# Coding Standards — modules/extensions/security/authn/grpc/

Transport adapter for `modules/security/authn`. Owns the gRPC unary +
stream server interceptors. Lives in `modules/extensions/` so
`google.golang.org/grpc` does not bleed into a consumer that does not
serve gRPC.

Inherits the workspace-wide standards from
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).
PACKAGES.md classification: **Shape B** — interface-returning
constructors (`NewUnaryInterceptor`, `NewStreamInterceptor`) + private
helpers.

See [`modules/security/authn/CODING_STANDARDS.md`](../../../../security/authn/CODING_STANDARDS.md)
for the failure contract. The stream interceptor wraps the
`grpc.ServerStream` with a `Context()` override so the upstream handler
sees the ctx augmented with the Principal — this is the canonical
google.golang.org/grpc pattern and the `containedctx` linter is
explicitly waived in `.golangci.yml` for `interceptor.go`.
