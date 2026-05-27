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

// Package grpc provides server-side gRPC unary and stream interceptors
// that terminate Bearer authentication.
//
// Both interceptors read the bearer token from the gRPC "authorization"
// metadata key (lowercased per HTTP/2 conventions), delegate
// verification to an authn.Authenticator, and on success inject the
// resulting *Principal into the per-RPC ctx via authn.WithPrincipal.
//
// Failure modes map onto codes.Unauthenticated:
//   - missing metadata or missing authorization key.
//   - malformed authorization value (wrong scheme or empty token).
//   - Authenticator.Validate returns an error.
//   - Authenticator returns a nil *Principal with no error.
//
// The interceptors do NOT log; the parent gRPC server's logging
// interceptor is responsible for surface visibility. They simply
// short-circuit the chain with a status.Error(codes.Unauthenticated,
// ...).
package grpc

import (
	"google.golang.org/grpc"

	"github.com/guidomantilla/yarumo/security/authn"
)

var (
	_ UnaryInterceptorFactoryFn  = NewUnaryInterceptor
	_ StreamInterceptorFactoryFn = NewStreamInterceptor
)

// UnaryInterceptorFactoryFn is the function type for
// NewUnaryInterceptor.
type UnaryInterceptorFactoryFn func(authenticator authn.Authenticator, options ...Option) grpc.UnaryServerInterceptor

// StreamInterceptorFactoryFn is the function type for
// NewStreamInterceptor.
type StreamInterceptorFactoryFn func(authenticator authn.Authenticator, options ...Option) grpc.StreamServerInterceptor
