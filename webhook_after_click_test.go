package types_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookHideButtons(t *testing.T) {
	t.Parallel()

	assert.True(t, types.WebhookHideButtonsIsValid(types.WebhookHideButtonsNone))
	assert.True(t, types.WebhookHideButtonsIsValid(types.WebhookHideButtonsClicked))
	assert.True(t, types.WebhookHideButtonsIsValid(types.WebhookHideButtonsAll))
	assert.False(t, types.WebhookHideButtonsIsValid(""))
	assert.False(t, types.WebhookHideButtonsIsValid("invalid"))

	assert.Equal(t, []string{"none", "clicked", "all"}, types.ValidWebhookHideButtons())
}

func TestWebhookAfterClick_ResetAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seconds int
		want    time.Duration
	}{
		{name: "zero uses default", seconds: 0, want: types.DefaultWebhookAfterClickResetSeconds * time.Second},
		{name: "negative uses default", seconds: -1, want: types.DefaultWebhookAfterClickResetSeconds * time.Second},
		{name: "explicit", seconds: 60, want: time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &types.WebhookAfterClick{ResetAfterSeconds: tt.seconds}
			assert.Equal(t, tt.want, w.ResetAfter())
		})
	}
}

func TestWebhookAfterClick_HidesWebhook(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		afterClick types.WebhookAfterClick
		webhookID  string
		want       bool
	}{
		{name: "empty hides nothing", afterClick: types.WebhookAfterClick{}, webhookID: "a", want: false},
		{name: "none hides nothing", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsNone}, webhookID: "a", want: false},
		{name: "none with listed id", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsNone, HideWebhookIDs: []string{"b"}}, webhookID: "b", want: true},
		{name: "none with unlisted id", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsNone, HideWebhookIDs: []string{"b"}}, webhookID: "c", want: false},
		{name: "clicked hides clicked", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsClicked}, webhookID: "a", want: true},
		{name: "clicked keeps others", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsClicked}, webhookID: "b", want: false},
		{name: "clicked with listed id", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsClicked, HideWebhookIDs: []string{"b"}}, webhookID: "b", want: true},
		{name: "all hides any", afterClick: types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsAll}, webhookID: "z", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.afterClick.HidesWebhook("a", tt.webhookID))
		})
	}
}

func TestWebhookAfterClick_JSON(t *testing.T) {
	t.Parallel()

	body := `{"id":"a","url":"http://foo.bar","buttonText":"Restart","afterClick":{"hideButtons":"clicked","hideWebhookIds":["b"],"statusText":"Restart in progress","resetAfterSeconds":120}}`

	var hook types.Webhook
	require.NoError(t, json.Unmarshal([]byte(body), &hook))
	require.NotNil(t, hook.AfterClick)
	assert.Equal(t, types.WebhookAfterClick{
		HideButtons:       types.WebhookHideButtonsClicked,
		HideWebhookIDs:    []string{"b"},
		StatusText:        "Restart in progress",
		ResetAfterSeconds: 120,
	}, *hook.AfterClick)

	// A nil AfterClick is omitted
	out, err := json.Marshal(&types.Webhook{ID: "a"})
	require.NoError(t, err)
	assert.NotContains(t, string(out), "afterClick")
}

func newAfterClickAlert(afterClick *types.WebhookAfterClick) *types.Alert {
	return &types.Alert{Header: "a", RouteKey: "b", Webhooks: []*types.Webhook{
		{ID: "a", URL: "http://foo.bar", ButtonText: "A", AfterClick: afterClick},
		{ID: "b", URL: "http://foo.bar", ButtonText: "B"},
	}}
}

func TestAlertCleanWebhookAfterClick(t *testing.T) {
	t.Parallel()

	a := newAfterClickAlert(&types.WebhookAfterClick{StatusText: "  in progress  ", HideWebhookIDs: []string{" b "}})
	a.Clean()

	require.NoError(t, a.Validate())
	assert.Equal(t, types.WebhookHideButtonsNone, a.Webhooks[0].AfterClick.HideButtons)
	assert.Equal(t, "in progress", a.Webhooks[0].AfterClick.StatusText)
	assert.Equal(t, []string{"b"}, a.Webhooks[0].AfterClick.HideWebhookIDs)
	assert.Nil(t, a.Webhooks[1].AfterClick)

	// Explicit values are kept
	a = newAfterClickAlert(&types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsAll})
	a.Clean()
	assert.Equal(t, types.WebhookHideButtonsAll, a.Webhooks[0].AfterClick.HideButtons)
}

func TestAlertValidationWebhookAfterClick(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		afterClick *types.WebhookAfterClick
		wantErr    string
	}{
		{name: "nil", afterClick: nil},
		{name: "empty", afterClick: &types.WebhookAfterClick{}},
		{name: "full", afterClick: &types.WebhookAfterClick{HideButtons: types.WebhookHideButtonsClicked, HideWebhookIDs: []string{"a", "b"}, StatusText: "x", ResetAfterSeconds: types.MaxWebhookAfterClickResetSeconds}},
		{name: "max status text", afterClick: &types.WebhookAfterClick{StatusText: strings.Repeat("x", types.MaxWebhookAfterClickStatusTextLength)}},
		{
			name:       "invalid hide buttons",
			afterClick: &types.WebhookAfterClick{HideButtons: "foo"},
			wantErr:    "webhook[0].afterClick.hideButtons 'foo' is not valid, expected empty or one of [none, clicked, all]",
		},
		{
			name:       "unknown webhook id",
			afterClick: &types.WebhookAfterClick{HideWebhookIDs: []string{"b", "c"}},
			wantErr:    "webhook[0].afterClick.hideWebhookIds[1] 'c' does not match any webhook id in the alert",
		},
		{
			name:       "empty webhook id",
			afterClick: &types.WebhookAfterClick{HideWebhookIDs: []string{" "}},
			wantErr:    "webhook[0].afterClick.hideWebhookIds[0] '' does not match any webhook id in the alert",
		},
		{
			name:       "too many webhook ids",
			afterClick: &types.WebhookAfterClick{HideWebhookIDs: []string{"a", "b", "a", "b", "a", "b"}},
			wantErr:    "webhook[0].afterClick.hideWebhookIds item count is too large",
		},
		{
			name:       "status text too long",
			afterClick: &types.WebhookAfterClick{StatusText: strings.Repeat("x", types.MaxWebhookAfterClickStatusTextLength+1)},
			wantErr:    "webhook[0].afterClick.statusText is too long",
		},
		{
			name:       "negative reset",
			afterClick: &types.WebhookAfterClick{ResetAfterSeconds: -1},
			wantErr:    "webhook[0].afterClick.resetAfterSeconds must be between 0 and 86400",
		},
		{
			name:       "reset too large",
			afterClick: &types.WebhookAfterClick{ResetAfterSeconds: types.MaxWebhookAfterClickResetSeconds + 1},
			wantErr:    "webhook[0].afterClick.resetAfterSeconds must be between 0 and 86400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := newAfterClickAlert(tt.afterClick)
			a.Clean()

			err := a.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}

	// A later webhook can refer to an earlier one, and vice versa
	a := newAfterClickAlert(nil)
	a.Webhooks[1].AfterClick = &types.WebhookAfterClick{HideWebhookIDs: []string{"a"}}
	a.Clean()
	require.NoError(t, a.Validate())
}
