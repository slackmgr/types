package types

import (
	"context"
	"errors"
	"net/http"
	"slices"
)

// Authorization scopes of the Slack Manager REST API. Each protected endpoint requires one scope.
// See [Principal.HasScope].
const (
	// ScopeAlertsWrite allows sending alerts: POST /alerts (and its deprecated aliases) and POST /prometheus-alert.
	ScopeAlertsWrite = "alerts:write"

	// ScopeAlertsTest allows testing alerts with POST /alerts-test. Nothing is sent to Slack.
	ScopeAlertsTest = "alerts:test"

	// ScopeConfigRead allows reading the routing rules (GET /mappings) and the managed channels (GET /channels).
	ScopeConfigRead = "config:read"

	// ScopeAll grants all scopes.
	ScopeAll = "*"
)

// KnownScopes returns all scopes of the REST API, including [ScopeAll].
func KnownScopes() []string {
	return []string{ScopeAlertsWrite, ScopeAlertsTest, ScopeConfigRead, ScopeAll}
}

// IsKnownScope reports whether scope is one of [KnownScopes].
func IsKnownScope(scope string) bool {
	return slices.Contains(KnownScopes(), scope)
}

// Errors returned by [Authenticator] and [Authorizer] implementations. Return them (or wrap them) so that the
// API responds with the matching status code. Any other error is an internal error (HTTP 500): the request is
// rejected (fail closed).
var (
	// ErrNoCredentials means that the request has no credentials that the authenticator understands. The API
	// responds with HTTP 401, or lets the request through anonymously in permissive mode.
	// [ChainAuthenticators] tries the next authenticator.
	ErrNoCredentials = errors.New("no credentials")

	// ErrInvalidCredentials means that the request has credentials, but they are invalid (unknown, expired,
	// malformed). The API responds with HTTP 401, also in permissive mode.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrForbidden means that the principal is authenticated, but not allowed to perform the action. The API
	// responds with HTTP 403.
	ErrForbidden = errors.New("forbidden")
)

// Principal is an authenticated caller of the REST API, e.g. a monitoring system or integration.
type Principal struct {
	// ID is the stable identity of the caller, e.g. "alertmanager-prod". It is logged with each request.
	ID string `json:"id"`

	// Scopes are the scopes granted to the caller, e.g. [ScopeAlertsWrite]. See [Principal.HasScope].
	Scopes []string `json:"scopes,omitempty"`

	// Attributes holds optional, authenticator specific information, e.g. token claims.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// HasScope reports whether the principal has the given scope, or [ScopeAll]. A nil principal has no scopes.
func (p *Principal) HasScope(scope string) bool {
	if p == nil {
		return false
	}

	return slices.Contains(p.Scopes, scope) || slices.Contains(p.Scopes, ScopeAll)
}

// Authenticator identifies the caller of an HTTP request.
//
// Authenticate returns the [Principal] of the request, or an error: [ErrNoCredentials] if the request has no
// credentials of the kind the authenticator handles, [ErrInvalidCredentials] if they are invalid, or any other
// error for internal failures (e.g. an unreachable identity provider). It must not read the request body.
// It is called concurrently and must be safe for concurrent use.
type Authenticator interface {
	Authenticate(r *http.Request) (*Principal, error)
}

// AuthenticatorFunc is a function that implements [Authenticator].
type AuthenticatorFunc func(r *http.Request) (*Principal, error)

// Authenticate calls f(r).
func (f AuthenticatorFunc) Authenticate(r *http.Request) (*Principal, error) {
	return f(r)
}

// Authorizer decides whether an authenticated [Principal] may perform an action that requires the given scope
// (e.g. [ScopeAlertsWrite]).
//
// Authorize returns nil if the action is allowed, [ErrForbidden] if it is not, or any other error for internal
// failures (e.g. an unreachable policy service). It is called concurrently and must be safe for concurrent use.
type Authorizer interface {
	Authorize(ctx context.Context, principal *Principal, scope string) error
}

// AuthorizerFunc is a function that implements [Authorizer].
type AuthorizerFunc func(ctx context.Context, principal *Principal, scope string) error

// Authorize calls f(ctx, principal, scope).
func (f AuthorizerFunc) Authorize(ctx context.Context, principal *Principal, scope string) error {
	return f(ctx, principal, scope)
}

// ChainAuthenticators returns an [Authenticator] (as an [AuthenticatorFunc]) that tries the given authenticators in order, and returns the
// first principal found. Nil authenticators are skipped.
//
// If no authenticator succeeds, it returns the most significant error: an internal error (the first one) over
// [ErrInvalidCredentials] over [ErrNoCredentials]. This lets authenticators of different kinds (e.g. API keys
// and OIDC tokens, which are both bearer tokens) share a request header.
func ChainAuthenticators(authenticators ...Authenticator) AuthenticatorFunc {
	chain := make([]Authenticator, 0, len(authenticators))

	for _, a := range authenticators {
		if a != nil {
			chain = append(chain, a)
		}
	}

	return AuthenticatorFunc(func(r *http.Request) (*Principal, error) {
		var internalErr error

		invalid := false

		for _, a := range chain {
			principal, err := a.Authenticate(r)

			switch {
			case err == nil && principal != nil:
				return principal, nil
			case err == nil, errors.Is(err, ErrInvalidCredentials):
				invalid = true
			case errors.Is(err, ErrNoCredentials):
			default:
				if internalErr == nil {
					internalErr = err
				}
			}
		}

		switch {
		case internalErr != nil:
			return nil, internalErr
		case invalid:
			return nil, ErrInvalidCredentials
		default:
			return nil, ErrNoCredentials
		}
	})
}
