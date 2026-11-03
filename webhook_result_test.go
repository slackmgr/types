package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookResult_Clean(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var r *types.WebhookResult
		assert.NotPanics(t, r.Clean)
	})

	t.Run("defaults and trimming", func(t *testing.T) {
		t.Parallel()

		r := &types.WebhookResult{
			Message: "  hello  ",
			AfterClick: &types.WebhookAfterClick{
				StatusText:     "  in progress ",
				HideWebhookIDs: []string{" a "},
			},
		}

		r.Clean()

		assert.Equal(t, "hello", r.Message)
		assert.Equal(t, types.WebhookHideButtonsNone, r.AfterClick.HideButtons)
		assert.Equal(t, "in progress", r.AfterClick.StatusText)
		assert.Equal(t, []string{"a"}, r.AfterClick.HideWebhookIDs)
	})

	t.Run("without after-click", func(t *testing.T) {
		t.Parallel()

		r := &types.WebhookResult{Message: " no "}
		r.Clean()

		assert.Equal(t, "no", r.Message)
		assert.Nil(t, r.AfterClick)
	})
}

func TestWebhookResult_Validate(t *testing.T) {
	t.Parallel()

	tooManyIDs := make([]string, types.MaxWebhookCount+1)
	for i := range tooManyIDs {
		tooManyIDs[i] = "id"
	}

	tests := []struct {
		name    string
		result  *types.WebhookResult
		wantErr string
	}{
		{name: "nil", result: nil, wantErr: "webhook result is nil"},
		{name: "empty", result: &types.WebhookResult{}},
		{name: "message only", result: &types.WebhookResult{Message: "bad input"}},
		{
			name:   "message at max length",
			result: &types.WebhookResult{Message: strings.Repeat("a", types.MaxWebhookResultMessageLength)},
		},
		{
			name:    "message too long",
			result:  &types.WebhookResult{Message: strings.Repeat("a", types.MaxWebhookResultMessageLength+1)},
			wantErr: "message is too long",
		},
		{
			name: "ok with after-click",
			result: &types.WebhookResult{AfterClick: &types.WebhookAfterClick{
				HideButtons:       types.WebhookHideButtonsClicked,
				HideWebhookIDs:    []string{"any-id"},
				StatusText:        "in progress",
				ResetAfterSeconds: 60,
			}},
		},
		{
			name:    "invalid hide buttons",
			result:  &types.WebhookResult{AfterClick: &types.WebhookAfterClick{HideButtons: "some"}},
			wantErr: "afterClick.hideButtons 'some' is not valid",
		},
		{
			name:    "too many hide ids",
			result:  &types.WebhookResult{AfterClick: &types.WebhookAfterClick{HideWebhookIDs: tooManyIDs}},
			wantErr: "afterClick.hideWebhookIds item count is too large",
		},
		{
			name:    "empty hide id",
			result:  &types.WebhookResult{AfterClick: &types.WebhookAfterClick{HideWebhookIDs: []string{""}}},
			wantErr: "afterClick.hideWebhookIds[0] is required",
		},
		{
			name: "hide id too long",
			result: &types.WebhookResult{AfterClick: &types.WebhookAfterClick{
				HideWebhookIDs: []string{strings.Repeat("a", types.MaxWebhookIDLength+1)},
			}},
			wantErr: "afterClick.hideWebhookIds[0] is too long",
		},
		{
			name: "status text too long",
			result: &types.WebhookResult{AfterClick: &types.WebhookAfterClick{
				StatusText: strings.Repeat("a", types.MaxWebhookAfterClickStatusTextLength+1),
			}},
			wantErr: "afterClick.statusText is too long",
		},
		{
			name: "reset too large",
			result: &types.WebhookResult{AfterClick: &types.WebhookAfterClick{
				ResetAfterSeconds: types.MaxWebhookAfterClickResetSeconds + 1,
			}},
			wantErr: "afterClick.resetAfterSeconds must be between 0 and 86400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.result.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestWebhookResponse_JSON(t *testing.T) {
	t.Parallel()

	t.Run("with result", func(t *testing.T) {
		t.Parallel()

		body := `{"slackManagerResult": {"message": "Restarting", "afterClick": {"hideButtons": "all", "statusText": "Working"}}}`

		var resp types.WebhookResponse
		require.NoError(t, json.Unmarshal([]byte(body), &resp))
		require.NotNil(t, resp.Result)

		assert.Equal(t, "Restarting", resp.Result.Message)
		require.NotNil(t, resp.Result.AfterClick)
		assert.Equal(t, types.WebhookHideButtonsAll, resp.Result.AfterClick.HideButtons)
		assert.Equal(t, "Working", resp.Result.AfterClick.StatusText)
	})

	t.Run("without result key", func(t *testing.T) {
		t.Parallel()

		var resp types.WebhookResponse
		require.NoError(t, json.Unmarshal([]byte(`{"message": "ok"}`), &resp))
		assert.Nil(t, resp.Result)
	})

	t.Run("marshal uses result key", func(t *testing.T) {
		t.Parallel()

		resp := types.WebhookResponse{Result: &types.WebhookResult{Message: "no"}}

		b, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.JSONEq(t, `{"`+types.WebhookResultKey+`": {"message": "no"}}`, string(b))
	})
}
