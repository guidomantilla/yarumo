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

// Package http provides a server-side net/http middleware that
// terminates authorization against the security/authz contract.
//
// The middleware reads the authenticated principal from ctx via a
// caller-provided PrincipalReader (typically wired to authn's
// FromContext) and calls Policy.Evaluate on every inbound request.
// Allow forwards to the next handler; Deny / Abstain short-circuits
// with 403 Forbidden, surfacing Decision.Reason both as a JSON body
// envelope ({"error": "forbidden", "reason": "..."}) and as the
// X-Authz-Reason response header for plain-text consumers.
//
// Configuration mirrors the gRPC adapter: a PrincipalReader is
// mandatory (missing reader denies every call), an audit hook is
// optional (default logs via common/log), and an optional resource
// resolver translates the inbound HTTP request into an authz.Resource
// for policy evaluation.
package http

import (
	"net/http"

	"github.com/guidomantilla/yarumo/security/authz"
)

var (
	_ RequireHTTPFn = RequireHTTP
)

// RequireHTTPFn is the function type for RequireHTTP.
type RequireHTTPFn func(policy authz.Policy, action string, opts ...Option) func(http.Handler) http.Handler

// HTTPResourceResolverFn is the function type for resolving the
// Resource an HTTP request targets. The middleware invokes the
// resolver once per inbound request before evaluating the policy.
//
// Returning the zero Resource is valid for action-only checks where
// resource type/id are irrelevant.
type HTTPResourceResolverFn func(r *http.Request) authz.Resource
