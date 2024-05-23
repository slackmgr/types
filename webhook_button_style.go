package common

// WebhookButtonStyle represents a webhook button style.
type WebhookButtonStyle string

const (
	// WebhookButtonStylePrimary represents Slack button style 'primary'.
	WebhookButtonStylePrimary WebhookButtonStyle = "primary"

	// WebhookButtonStyleDanger represents Slack button style 'danger'.
	WebhookButtonStyleDanger WebhookButtonStyle = "danger"
)

var validWebhookButtonStyles map[WebhookButtonStyle]struct{}

func init() {
	validWebhookButtonStyles = map[WebhookButtonStyle]struct{}{
		WebhookButtonStylePrimary: {},
		WebhookButtonStyleDanger:  {},
	}
}

// WebhookButtonStyleIsValid returns true if the provided WebhookButtonStyle is valid.
func WebhookButtonStyleIsValid(s WebhookButtonStyle) bool {
	_, ok := validWebhookButtonStyles[s]
	return ok
}

// ValidWebhookButtonStyles returns a slice of valid WebhookButtonStyle values.
func ValidWebhookButtonStyles() []string {
	r := make([]string, len(validWebhookButtonStyles))
	i := 0

	for s := range validWebhookButtonStyles {
		r[i] = string(s)
		i++
	}

	return r
}
