package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
)

func TestAlertSeverityValidation(t *testing.T) {
	t.Parallel()

	assert.True(t, common.SeverityIsValid(common.AlertPanic))
	assert.True(t, common.SeverityIsValid(common.AlertError))
	assert.True(t, common.SeverityIsValid(common.AlertWarning))
	assert.True(t, common.SeverityIsValid(common.AlertResolved))
	assert.True(t, common.SeverityIsValid(common.AlertInfo))
	assert.False(t, common.SeverityIsValid("invalid"))
}

func TestAlertPriority(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 3, common.SeverityPriority(common.AlertPanic))
	assert.Equal(t, 2, common.SeverityPriority(common.AlertError))
	assert.Equal(t, 1, common.SeverityPriority(common.AlertWarning))
	assert.Equal(t, 0, common.SeverityPriority(common.AlertResolved))
	assert.Equal(t, 0, common.SeverityPriority(common.AlertInfo))
	assert.Equal(t, -1, common.SeverityPriority("invalid"))
}

func TestValidSeverities(t *testing.T) {
	t.Parallel()

	s := common.ValidSeverities()
	assert.Len(t, s, 5)
	assert.Contains(t, s, "panic")
	assert.Contains(t, s, "error")
	assert.Contains(t, s, "warning")
	assert.Contains(t, s, "resolved")
	assert.Contains(t, s, "info")
}
