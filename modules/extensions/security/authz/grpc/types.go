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

// Package grpc provides gRPC unary and stream server interceptors that
// terminate authorization against the security/authz contract.
//
// The interceptors read the authenticated principal from ctx via a
// caller-provided PrincipalReader (typically wired to authn's
// FromContext) and call Policy.Evaluate on every inbound RPC. Allow
// forwards to the handler; Deny / Abstain short-circuits with
// codes.PermissionDenied and Decision.Reason as the status message.
//
// Configuration mirrors the HTTP adapter: a PrincipalReader is
// mandatory (missing reader denies every call), an audit hook is
// optional (default logs via common/log), and an optional resource
// resolver translates the inbound RPC into an authz.Resource for
// policy evaluation.
package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/guidomantilla/yarumo/security/authz"
)

var (
	_ RequireUnaryFn        = RequireUnary
	_ RequireStreamFn       = RequireStream
	_ GRPCResourceResolverFn = (GRPCResourceResolverFn)(nil)
)

// RequireUnaryFn is the function type for RequireUnary.
type RequireUnaryFn func(policy authz.Policy, action string, opts ...Option) grpc.UnaryServerInterceptor

// RequireStreamFn is the function type for RequireStream.
type RequireStreamFn func(policy authz.Policy, action string, opts ...Option) grpc.StreamServerInterceptor

// GRPCResourceResolverFn is the function type for resolving the
// Resource a gRPC RPC targets. The interceptor invokes the resolver
// once per inbound RPC before evaluating the policy. method is the
// gRPC FullMethod ("/pkg.Service/Method"), req is the typed request
// message (any) for unary calls or nil for stream calls.
//
// Returning the zero Resource is valid for action-only checks where
// resource type/id are irrelevant.
type GRPCResourceResolverFn func(ctx context.Context, method string, req any) authz.Resource
