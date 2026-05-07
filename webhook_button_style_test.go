package types_test

import (
	"testing"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookButtonStyle(t *testing.T) {
	t.Parallel()

	assert.True(t, types.WebhookButtonStyleIsValid(types.WebhookButtonStyleDefault))
	assert.True(t, types.WebhookButtonStyleIsValid(types.WebhookButtonStylePrimary))
	assert.True(t, types.WebhookButtonStyleIsValid(types.WebhookButtonStyleDanger))
	assert.False(t, types.WebhookButtonStyleIsValid("invalid"))
}

func TestWebhookButtonStyleString(t *testing.T) {
	t.Parallel()

	s := types.ValidWebhookButtonStyles()
	assert.Len(t, s, 3)
	assert.Contains(t, s, "default")
	assert.Contains(t, s, "primary")
	assert.Contains(t, s, "danger")
}
