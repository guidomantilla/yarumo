// Package authz provides authorization primitives: a Policy interface,
// a Decision type carrying Allow/Deny/Abstain with a reason, a Request
// envelope (Principal/Resource/Action/Environment), and Require
// middleware for HTTP and gRPC.
//
// Sub-package rbac implements role-based access control with role
// inheritance and permission wildcards on top of the Policy interface.
//
// Design notes
//
// Principal is typed as any inside Request so authz does NOT take a
// dependency on authn. Consumers cast inside the policy to whatever
// shape their authentication layer produces (typically an
// authn.Principal). This keeps the two modules independent and lets
// the Policy be reused with custom principal shapes (machine identities,
// API keys, service accounts, etc.).
//
// The Require HTTP and gRPC middleware do not pull the Principal from
// a hardcoded context key. Consumers wire a PrincipalReader option
// (WithPrincipalReader) which knows how their authn layer stashes the
// principal in ctx. This keeps coupling explicit and lets the same
// middleware run on top of any authn implementation.
//
// Concurrency: all public types in this package are safe for concurrent
// use by multiple goroutines.
package authz
