// Copyright 2026 Guido Mauricio Mantilla Tarazona
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package authn provides authentication primitives for the workspace.
//
// The root package defines the public abstractions only:
//
//   - Principal: the immutable identity of an authenticated caller
//     (id, name, roles, attributes), produced by an Authenticator and
//     propagated through ctx by transport-specific middleware.
//   - Authenticator: a token-string-to-Principal validator. Concrete
//     implementations live in subpackages so consumers pull only the
//     dependencies they actually use.
//   - WithPrincipal / FromContext: context propagation helpers used by
//     every transport middleware (HTTP, gRPC, future ones).
//
// Subpackages:
//
//   - authn/token: TokenAuthenticator that delegates verification to
//     modules/crypto/tokens. Works with every algorithm crypto/tokens
//     supports — the JWT family (HS/RS/PS/ES/EdDSA) and the opaque
//     AEAD family (OPAQUE_AES_GCM, OPAQUE_XCHACHA20_POLY1305) — since
//     dispatch is owned by *tokens.Method. Pulls golang-jwt/v5
//     transitively only when imported.
//
// Transport adapters live in their own top-level modules under
// modules/extensions/security/authn/ so a consumer of this contract
// never pulls google.golang.org/grpc unless it imports the grpc
// adapter explicitly:
//
//   - extensions/security/authn/http: net/http server middleware
//     extracting the Authorization header, validating the bearer
//     token, and injecting the Principal into the request ctx.
//   - extensions/security/authn/grpc: gRPC unary + stream server
//     interceptors mirroring the HTTP middleware, reading the bearer
//     token from the "authorization" metadata key.
//
// The package owns no lifecycle and spawns no goroutines: Authenticators
// are stateless validators, transport middleware are pure function
// composition.
package authn

import (
	"context"
)

var (
	_ WithPrincipalFn     = WithPrincipal
	_ FromContextFn       = FromContext
	_ ErrAuthenticationFn = ErrAuthentication
)

// WithPrincipalFn is the function type for WithPrincipal.
type WithPrincipalFn func(ctx context.Context, principal *Principal) context.Context

// FromContextFn is the function type for FromContext.
type FromContextFn func(ctx context.Context) (*Principal, bool)

// ErrAuthenticationFn is the function type for ErrAuthentication.
type ErrAuthenticationFn func(causes ...error) error

// Authenticator defines the contract for validating an authentication
// token and extracting the resulting Principal.
//
// Implementations are stateless validators: they take an opaque token
// string (typically a JWT, an opaque AEAD-encrypted token, an API key,
// etc.) and produce the *Principal it represents on success. The token
// shape and verification semantics are implementation-defined.
//
// Implementations must be safe for concurrent use by multiple goroutines.
// On failure they MUST return an error wrapping a sentinel from the
// authn package (ErrTokenEmpty, ErrTokenInvalid) so transport middleware
// can translate verification failures to a uniform 401 response without
// inspecting concrete impl-specific error types.
type Authenticator interface {
	// Validate verifies token and returns the *Principal it carries.
	// Implementations must return ErrTokenEmpty when token is the empty
	// string and an ErrTokenInvalid-wrapped error for any verification
	// failure (signature, expiry, issuer mismatch, malformed claims).
	// ctx is propagated so implementations that talk to an external
	// service can honor caller cancellation.
	Validate(ctx context.Context, token string) (*Principal, error)
}

// Principal is the immutable identity of an authenticated caller.
//
// It carries the minimal data needed by downstream authorization layers:
// an opaque ID (typically the JWT "sub"), a human-readable name (the
// "name" claim or equivalent), the roles assigned to the caller, and a
// free-form attributes bag for impl-specific claims (tenant, scopes,
// permissions, custom JWT claims, etc.).
//
// Principal is constructed once by an Authenticator and propagated
// read-only through ctx via WithPrincipal / FromContext. Mutating a
// Principal after construction is undefined behavior; treat the value as
// frozen.
type Principal struct {
	// ID is the stable identifier of the caller (e.g. the JWT "sub"
	// claim, a user ID, a service account name). Never empty for a
	// successfully validated Principal.
	ID string
	// Name is the human-readable display name of the caller, when the
	// underlying token carries one. May be empty.
	Name string
	// Roles is the list of roles granted to the caller. The slice may
	// be empty but is never nil for a Principal produced by a
	// well-formed Authenticator.
	Roles []string
	// Attributes is a free-form bag for impl-specific claims (tenant,
	// scopes, permissions, arbitrary custom JWT claims). The map may
	// be empty but is never nil for a Principal produced by a
	// well-formed Authenticator. Values are typed as any to mirror the
	// JSON shape of JWT claims.
	Attributes map[string]any
}
