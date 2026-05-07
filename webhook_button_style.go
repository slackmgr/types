package types

// WebhookButtonStyle represents a webhook button style.
type WebhookButtonStyle string

const (
	// WebhookButtonStyleDefault represents the default Slack button style (no special styling).
	WebhookButtonStyleDefault WebhookButtonStyle = "default"

	// WebhookButtonStylePrimary represents Slack button style 'primary'.
	// Green outline and text, used for affirmative or key confirmation actions (e.g., "Approve", "Confirm").
	WebhookButtonStylePrimary WebhookButtonStyle = "primary"

	// WebhookButtonStyleDanger represents Slack button style 'danger'.
	// Red outline and text, used for actions with significant consequences, such as deletion (e.g., "Delete", "Remove").
	WebhookButtonStyleDanger WebhookButtonStyle = "danger"
)

// WebhookButtonStyleIsValid returns true if the provided WebhookButtonStyle is valid.
func WebhookButtonStyleIsValid(s WebhookButtonStyle) bool {
	switch s {
	case WebhookButtonStyleDefault, WebhookButtonStylePrimary, WebhookButtonStyleDanger:
		return true
	}
	return false
}

// ValidWebhookButtonStyles returns a slice of valid WebhookButtonStyle values.
func ValidWebhookButtonStyles() []string {
	return []string{
		string(WebhookButtonStyleDefault),
		string(WebhookButtonStylePrimary),
		string(WebhookButtonStyleDanger),
	}
}
