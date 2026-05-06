package types

import "time"

// WebhookCallback is the payload sent via HTTP POST to the webhook URL when a user clicks a webhook button.
type WebhookCallback struct {
	// ID is the webhook ID from the originating Webhook definition.
	ID string `json:"id"`

	// UserID is the Slack user ID of the user who clicked the button.
	UserID string `json:"userId"`

	// UserRealName is the display name of the user who clicked the button.
	UserRealName string `json:"userRealName"`

	// ChannelID is the Slack channel ID where the button was clicked.
	ChannelID string `json:"channelId"`

	// MessageID is the Slack message timestamp (ts) of the post that contained the button.
	MessageID string `json:"messageId"`

	// Timestamp is the time when the webhook was triggered.
	Timestamp time.Time `json:"timestamp"`

	// PlainTextInput contains the values entered by the user in the webhook's PlainTextInput fields,
	// keyed by each input's ID.
	PlainTextInput map[string]string `json:"plainTextInput"`

	// CheckboxInput contains the selected option values from the webhook's CheckboxInput groups,
	// keyed by each input's ID. Each value is the list of selected option values.
	CheckboxInput map[string][]string `json:"checkboxInput"`

	// Payload contains the static key-value pairs from the originating Webhook.Payload definition.
	// These values are set at alert creation time and are not modified by user interaction.
	Payload map[string]any `json:"payload"`
}

func (w *WebhookCallback) GetPayloadValue(key string) any {
	if w == nil || w.Payload == nil {
		return ""
	}

	if v, ok := w.Payload[key]; ok {
		return v
	}

	return nil
}

func (w *WebhookCallback) GetPayloadString(key string) string {
	if w == nil || w.Payload == nil {
		return ""
	}

	if s, ok := w.Payload[key]; ok {
		if val, ok := s.(string); ok {
			return val
		}
	}

	return ""
}

func (w *WebhookCallback) GetPayloadInt(key string, defaultValue int) int {
	if w == nil || w.Payload == nil {
		return defaultValue
	}

	if s, ok := w.Payload[key]; ok {
		if val, ok := s.(int); ok {
			return val
		}
	}

	return defaultValue
}

func (w *WebhookCallback) GetPayloadBool(key string, defaultValue bool) bool {
	if w == nil || w.Payload == nil {
		return defaultValue
	}

	if b, ok := w.Payload[key]; ok {
		if val, ok := b.(bool); ok {
			return val
		}
	}

	return defaultValue
}

func (w *WebhookCallback) GetInputValue(key string) string {
	if w == nil || w.PlainTextInput == nil {
		return ""
	}

	if s, ok := w.PlainTextInput[key]; ok {
		return s
	}

	return ""
}

func (w *WebhookCallback) GetCheckboxInputSelectedValues(key string) []string {
	if w == nil || w.CheckboxInput == nil {
		return []string{}
	}

	if v, ok := w.CheckboxInput[key]; ok {
		return v
	}

	return []string{}
}
