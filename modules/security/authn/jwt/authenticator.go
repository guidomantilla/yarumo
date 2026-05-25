package jwt

import (
	"context"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctokens "github.com/guidomantilla/yarumo/crypto/tokens"
	"github.com/guidomantilla/yarumo/security/authn"
)

// jwtAuthenticator is the JWT-backed authn.Authenticator. It delegates
// signature/expiry verification to *tokens.Method and reshapes the
// resulting Payload map into a *Principal according to the configured
// claim keys.
type jwtAuthenticator struct {
	method       *ctokens.Method
	subjectClaim string
	nameClaim    string
	rolesClaim   string
}

// NewJWTAuthenticator returns a stateless authn.Authenticator backed by
// the given *tokens.Method. The Method must have been built with a
// verifying key (WithKey / WithVerifyingKey / WithGeneratedKey) — there
// is no implicit fallback. A nil method panics via cassert.NotNil so
// construction-time misconfiguration surfaces immediately rather than
// at the first Validate call.
//
// Claim mapping defaults to "sub" → Principal.ID, "name" →
// Principal.Name, "roles" → Principal.Roles. All other Payload keys
// flow into Principal.Attributes verbatim. Override via WithSubjectClaim
// / WithNameClaim / WithRolesClaim.
func NewJWTAuthenticator(method *ctokens.Method, options ...Option) authn.Authenticator {
	cassert.NotNil(method, "tokens method is nil")

	opts := NewOptions(options...)

	return &jwtAuthenticator{
		method:       method,
		subjectClaim: opts.subjectClaim,
		nameClaim:    opts.nameClaim,
		rolesClaim:   opts.rolesClaim,
	}
}

// Validate verifies the JWT token via the underlying *tokens.Method and
// reshapes the resulting Payload into a *Principal. The returned error
// is always wrapped through authn.ErrAuthentication so transport
// middleware can translate verification failures to a uniform 401
// without inspecting concrete error types.
func (a *jwtAuthenticator) Validate(_ context.Context, token string) (*authn.Principal, error) {
	cassert.NotNil(a, "jwtAuthenticator is nil")

	if token == "" {
		return nil, authn.ErrAuthentication(authn.ErrTokenEmpty)
	}

	payload, err := a.method.Validate(token)
	if err != nil {
		return nil, authn.ErrAuthentication(authn.ErrTokenInvalid, err)
	}

	principal, err := principalFromPayload(payload, a.subjectClaim, a.nameClaim, a.rolesClaim)
	if err != nil {
		return nil, authn.ErrAuthentication(authn.ErrTokenInvalid, err)
	}

	return principal, nil
}

// principalFromPayload extracts a *Principal from the JWT payload map
// using the configured claim keys. Validation rules:
//
//   - The subject claim MUST be a non-empty string; otherwise the
//     function returns ErrSubjectClaimMissing.
//   - The name claim is optional; non-string values are ignored.
//   - The roles claim accepts []any (typical JSON decoding) or
//     []string. Non-string entries are skipped. Missing / wrong type
//     produces an empty (non-nil) slice.
//   - Every payload entry that is NOT one of the three mapped claims is
//     copied into Principal.Attributes verbatim.
func principalFromPayload(payload ctokens.Payload, subjectClaim, nameClaim, rolesClaim string) (*authn.Principal, error) {
	subject, ok := payload[subjectClaim].(string)
	if !ok || subject == "" {
		return nil, ErrSubjectClaimMissing
	}

	name, _ := payload[nameClaim].(string)
	roles := extractRoles(payload[rolesClaim])

	attributes := make(map[string]any, len(payload))

	for key, value := range payload {
		if key == subjectClaim || key == nameClaim || key == rolesClaim {
			continue
		}

		attributes[key] = value
	}

	return &authn.Principal{
		ID:         subject,
		Name:       name,
		Roles:      roles,
		Attributes: attributes,
	}, nil
}

// extractRoles converts a JWT-decoded roles claim into []string. JSON
// arrays decode as []any in Go, so the common case is a type-switch
// over []any with per-element string assertion. A []string is also
// accepted for callers that hand-craft the Payload in tests. Anything
// else degrades to an empty (non-nil) slice.
func extractRoles(value any) []string {
	roles := []string{}

	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			role, ok := item.(string)
			if !ok {
				continue
			}

			roles = append(roles, role)
		}
	case []string:
		roles = append(roles, typed...)
	}

	return roles
}
