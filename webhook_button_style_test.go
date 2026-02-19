package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
)

func TestWebhookButtonStyle(t *testing.T) {
	t.Parallel()

	assert.True(t, common.WebhookButtonStyleIsValid(common.WebhookButtonStylePrimary))
	assert.True(t, common.WebhookButtonStyleIsValid(common.WebhookButtonStyleDanger))
	assert.False(t, common.WebhookButtonStyleIsValid("invalid"))
}

func TestWebhookButtonStyleString(t *testing.T) {
	t.Parallel()

	s := common.ValidWebhookButtonStyles()
	assert.Len(t, s, 2)
	assert.Contains(t, s, "primary")
	assert.Contains(t, s, "danger")
}
