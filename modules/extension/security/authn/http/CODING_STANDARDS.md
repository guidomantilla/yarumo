# Coding Standards — modules/extension/security/authn/http/

Transport adapter for `modules/security/authn`. Owns the server-side
`net/http` Bearer middleware. Lives in `modules/extension/` so the
`google.golang.org/grpc` adapter does not bleed into a consumer that
serves HTTP only.

Inherits the workspace-wide standards from
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).
PACKAGES.md classification: **Shape B** — interface-returning
constructor + private struct, single canonical package.

See [`modules/security/authn/CODING_STANDARDS.md`](../../../../security/authn/CODING_STANDARDS.md)
for the failure contract: every error returned by the underlying
`Authenticator` is already wrapped through `authn.ErrAuthentication`,
so this middleware just translates to HTTP 401 (default) or whatever
the configured `ErrorHandler` decides.
