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

// Package token provides a token-backed Authenticator that delegates
// verification to modules/crypto/tokens.
//
// crypto/tokens supports two algorithm families behind a single
// *Method type: the JWT family (HS/RS/PS/ES/EdDSA) and the opaque AEAD
// family (OPAQUE_AES_GCM, OPAQUE_XCHACHA20_POLY1305). Method.Validate
// dispatches internally by algorithm, so this Authenticator works with
// every supported algorithm — the family is determined entirely by the
// *Method the caller wires.
//
// The package is a sibling of authn/http and authn/grpc and lives in
// its own subpackage so consumers that integrate with a non-token
// authentication backend never pull github.com/golang-jwt/jwt/v5 (a
// transitive dependency of crypto/tokens) into their build graph.
//
// # Claim mapping
//
// modules/crypto/tokens validates the token envelope (signature/expiry
// for JWT, AEAD authentication tag for opaque) and returns the inner
// Payload map — the JSON object under the "payload" key. The standard
// JWT registered claims (iss, sub, exp, ...) are consumed during
// verification and are NOT surfaced by Method.Validate. As a result
// this Authenticator extracts Principal fields exclusively from the
// Payload map; token producers must duplicate the subject (and
// optionally the name/roles) inside "payload" rather than relying on
// the registered "sub" claim being available downstream.
//
// The default claim keys are "sub" (Principal.ID), "name"
// (Principal.Name), "roles" (Principal.Roles). All non-mapped Payload
// entries are copied into Principal.Attributes verbatim. Callers can
// rename the source keys via WithSubjectClaim, WithNameClaim, and
// WithRolesClaim.
package token
