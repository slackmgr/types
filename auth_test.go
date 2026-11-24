package types_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnownScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"alerts:write", "alerts:test", "config:read", "*"}, types.KnownScopes())

	for _, scope := range types.KnownScopes() {
		assert.True(t, types.IsKnownScope(scope), scope)
	}

	assert.False(t, types.IsKnownScope(""))
	assert.False(t, types.IsKnownScope("alerts:read"))
	assert.False(t, types.IsKnownScope("ALERTS:WRITE"))
}

func TestPrincipal_HasScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		principal *types.Principal
		scope     string
		want      bool
	}{
		{name: "nil principal", principal: nil, scope: types.ScopeAlertsWrite, want: false},
		{name: "no scopes", principal: &types.Principal{ID: "a"}, scope: types.ScopeAlertsWrite, want: false},
		{name: "matching scope", principal: &types.Principal{ID: "a", Scopes: []string{types.ScopeConfigRead, types.ScopeAlertsWrite}}, scope: types.ScopeAlertsWrite, want: true},
		{name: "other scope", principal: &types.Principal{ID: "a", Scopes: []string{types.ScopeConfigRead}}, scope: types.ScopeAlertsWrite, want: false},
		{name: "all scopes", principal: &types.Principal{ID: "a", Scopes: []string{types.ScopeAll}}, scope: types.ScopeAlertsTest, want: true},
		{name: "case sensitive", principal: &types.Principal{ID: "a", Scopes: []string{"Alerts:Write"}}, scope: types.ScopeAlertsWrite, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.principal.HasScope(tt.scope))
		})
	}
}

func TestAuthorizerFunc(t *testing.T) {
	t.Parallel()

	var authz types.Authorizer = types.AuthorizerFunc(func(_ context.Context, p *types.Principal, scope string) error {
		if p.HasScope(scope) {
			return nil
		}

		return types.ErrForbidden
	})

	p := &types.Principal{ID: "a", Scopes: []string{types.ScopeAlertsWrite}}

	require.NoError(t, authz.Authorize(t.Context(), p, types.ScopeAlertsWrite))
	require.ErrorIs(t, authz.Authorize(t.Context(), p, types.ScopeConfigRead), types.ErrForbidden)
}

func TestChannelAuthorizerFunc(t *testing.T) {
	t.Parallel()

	var authz types.ChannelAuthorizer = types.ChannelAuthorizerFunc(func(_ context.Context, p *types.Principal, channelID string) error {
		if p.ID == "a" && channelID == "C123" {
			return nil
		}

		return types.ErrForbidden
	})

	p := &types.Principal{ID: "a"}

	require.NoError(t, authz.AuthorizeChannel(t.Context(), p, "C123"))
	require.ErrorIs(t, authz.AuthorizeChannel(t.Context(), p, "C456"), types.ErrForbidden)
}

func TestChainAuthenticators(t *testing.T) {
	t.Parallel()

	alice := &types.Principal{ID: "alice"}
	bob := &types.Principal{ID: "bob"}
	internalErr := errors.New("identity provider unavailable")

	ok := func(p *types.Principal) types.Authenticator {
		return types.AuthenticatorFunc(func(*http.Request) (*types.Principal, error) { return p, nil })
	}
	fail := func(err error) types.Authenticator {
		return types.AuthenticatorFunc(func(*http.Request) (*types.Principal, error) { return nil, err })
	}
	nilNil := types.AuthenticatorFunc(func(*http.Request) (*types.Principal, error) { return nil, nil }) //nolint:nilnil // tests a misbehaving authenticator

	tests := []struct {
		name           string
		authenticators []types.Authenticator
		wantPrincipal  *types.Principal
		wantErr        error
	}{
		{name: "empty chain", wantErr: types.ErrNoCredentials},
		{name: "only nil authenticators", authenticators: []types.Authenticator{nil, nil}, wantErr: types.ErrNoCredentials},
		{name: "first succeeds", authenticators: []types.Authenticator{ok(alice), ok(bob)}, wantPrincipal: alice},
		{name: "skips nil", authenticators: []types.Authenticator{nil, ok(bob)}, wantPrincipal: bob},
		{name: "no credentials then success", authenticators: []types.Authenticator{fail(types.ErrNoCredentials), ok(bob)}, wantPrincipal: bob},
		{name: "invalid then success", authenticators: []types.Authenticator{fail(types.ErrInvalidCredentials), ok(bob)}, wantPrincipal: bob},
		{name: "internal error then success", authenticators: []types.Authenticator{fail(internalErr), ok(bob)}, wantPrincipal: bob},
		{name: "all no credentials", authenticators: []types.Authenticator{fail(types.ErrNoCredentials), fail(types.ErrNoCredentials)}, wantErr: types.ErrNoCredentials},
		{name: "invalid wins over no credentials", authenticators: []types.Authenticator{fail(types.ErrNoCredentials), fail(types.ErrInvalidCredentials)}, wantErr: types.ErrInvalidCredentials},
		{name: "wrapped invalid", authenticators: []types.Authenticator{fail(fmt.Errorf("token expired: %w", types.ErrInvalidCredentials))}, wantErr: types.ErrInvalidCredentials},
		{name: "nil principal without error is invalid", authenticators: []types.Authenticator{nilNil}, wantErr: types.ErrInvalidCredentials},
		{name: "internal error wins", authenticators: []types.Authenticator{fail(types.ErrInvalidCredentials), fail(internalErr), fail(types.ErrNoCredentials)}, wantErr: internalErr},
		{name: "first internal error wins", authenticators: []types.Authenticator{fail(internalErr), fail(errors.New("other"))}, wantErr: internalErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			p, err := types.ChainAuthenticators(tt.authenticators...).Authenticate(req)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, p)

				return
			}

			require.NoError(t, err)
			assert.Same(t, tt.wantPrincipal, p)
		})
	}
}
