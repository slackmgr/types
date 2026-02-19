package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
)

func TestWebhookAccessLevelValidation(t *testing.T) {
	t.Parallel()

	assert.True(t, common.WebhookAccessLevelIsValid(common.WebhookAccessLevelGlobalAdmins))
	assert.True(t, common.WebhookAccessLevelIsValid(common.WebhookAccessLevelChannelAdmins))
	assert.True(t, common.WebhookAccessLevelIsValid(common.WebhookAccessLevelChannelMembers))
	assert.False(t, common.WebhookAccessLevelIsValid("invalid"))
}

func TestWebhookAccessLevelString(t *testing.T) {
	t.Parallel()

	s := common.ValidWebhookAccessLevels()
	assert.Len(t, s, 3)
	assert.Contains(t, s, "global_admins")
	assert.Contains(t, s, "channel_admins")
	assert.Contains(t, s, "channel_members")
}
