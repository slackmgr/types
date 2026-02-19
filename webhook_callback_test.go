package common_test

import (
	"testing"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
)

func TestWebhookGetPayloadValue(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetPayloadValue("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{}
	val = w.GetPayloadValue("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{
		Payload: map[string]any{
			"key": "value",
		},
	}
	val = w.GetPayloadValue("key")
	assert.Equal(t, "value", val)

	val = w.GetPayloadValue("invalid")
	assert.Empty(t, val)
}

func TestWebhookGetPayloadString(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetPayloadString("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{}
	val = w.GetPayloadString("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{
		Payload: map[string]any{
			"key": "value",
		},
	}
	val = w.GetPayloadString("key")
	assert.Equal(t, "value", val)

	val = w.GetPayloadString("invalid")
	assert.Empty(t, val)
}

func TestWebhookGetPayloadInt(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetPayloadInt("key", 42)
	assert.Equal(t, 42, val)

	w = &common.WebhookCallback{}
	val = w.GetPayloadInt("key", 42)
	assert.Equal(t, 42, val)

	w = &common.WebhookCallback{
		Payload: map[string]any{
			"key": 123,
		},
	}
	val = w.GetPayloadInt("key", 42)
	assert.Equal(t, 123, val)

	val = w.GetPayloadInt("invalid", 42)
	assert.Equal(t, 42, val)
}

func TestWebhookGetPayloadBool(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetPayloadBool("key", true)
	assert.True(t, val)

	w = &common.WebhookCallback{}
	val = w.GetPayloadBool("key", true)
	assert.True(t, val)

	w = &common.WebhookCallback{
		Payload: map[string]any{
			"key": true,
		},
	}
	val = w.GetPayloadBool("key", false)
	assert.True(t, val)

	val = w.GetPayloadBool("invalid", false)
	assert.False(t, val)
}

func TestWebhookGetInputValue(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetInputValue("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{}
	val = w.GetInputValue("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{
		Input: map[string]string{
			"key": "value",
		},
	}
	val = w.GetInputValue("key")
	assert.Equal(t, "value", val)

	val = w.GetInputValue("invalid")
	assert.Empty(t, val)
}

func TestGetCheckboxInputSelectedValues(t *testing.T) {
	t.Parallel()

	var w *common.WebhookCallback
	val := w.GetCheckboxInputSelectedValues("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{}
	val = w.GetCheckboxInputSelectedValues("key")
	assert.Empty(t, val)

	w = &common.WebhookCallback{
		CheckboxInput: map[string][]string{
			"key": {"value1", "value2"},
		},
	}
	val = w.GetCheckboxInputSelectedValues("key")
	assert.Equal(t, []string{"value1", "value2"}, val)

	val = w.GetCheckboxInputSelectedValues("invalid")
	assert.Empty(t, val)
}
