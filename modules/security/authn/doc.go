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
//   - authn/http: net/http middleware extracting the Authorization
//     header, validating the bearer token, and injecting the Principal
//     into the request ctx.
//   - authn/grpc: gRPC unary + stream server interceptors mirroring the
//     HTTP middleware, reading the bearer token from the "authorization"
//     metadata key.
//
// The package owns no lifecycle and spawns no goroutines: Authenticators
// are stateless validators, transport middleware are pure function
// composition.
package authn
