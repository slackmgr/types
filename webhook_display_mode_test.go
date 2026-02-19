package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
)

func TestWebhookDisplayMode(t *testing.T) {
	t.Parallel()

	assert.True(t, common.WebhookDisplayModeIsValid(common.WebhookDisplayModeAlways))
	assert.True(t, common.WebhookDisplayModeIsValid(common.WebhookDisplayModeOpenIssue))
	assert.True(t, common.WebhookDisplayModeIsValid(common.WebhookDisplayModeResolvedIssue))
	assert.False(t, common.WebhookDisplayModeIsValid("invalid"))
}

func TestWebhookDisplayModeString(t *testing.T) {
	t.Parallel()

	s := common.ValidWebhookDisplayModes()
	assert.Len(t, s, 3)
	assert.Contains(t, s, "always")
	assert.Contains(t, s, "open_issue")
	assert.Contains(t, s, "resolved_issue")
}
