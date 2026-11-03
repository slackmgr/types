package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// WebhookResultKey is the top-level JSON key of a [WebhookResponse]. A webhook response body is only treated as
	// a [WebhookResult] if it is a JSON object with this key.
	WebhookResultKey = "slackManagerResult"

	// MaxWebhookResultMessageLength is the maximum length of [WebhookResult.Message].
	MaxWebhookResultMessageLength = 1000
)

// WebhookResponse is the body a synchronous webhook receiver (e.g. an HTTP endpoint) may respond with, to give
// feedback on a webhook button click. It wraps the [WebhookResult] in the [WebhookResultKey] key, so that existing
// response bodies are not mistaken for results.
//
// Example body:
//
//	{"slackManagerResult": {"message": "Restarting 3 pods", "afterClick": {"hideButtons": "all"}}}
//
// Whether the request was accepted is signalled by the response itself, e.g. the HTTP status code: 2xx means
// accepted, 4xx means rejected.
type WebhookResponse struct {
	// Result is the webhook result. Nil if the body has no [WebhookResultKey] key.
	Result *WebhookResult `json:"slackManagerResult"`
}

// WebhookResult is the feedback from a synchronous webhook receiver on a webhook button click.
// See [WebhookResponse] for the response body format.
//
// The Message is shown to the user who clicked the button, both when the request was accepted and when it was
// rejected. The AfterClick state is applied to the Slack post (seen by everyone in the channel) only when the request
// was accepted, and overrides [Webhook.AfterClick] for this click.
type WebhookResult struct {
	// Message is an optional message for the user who clicked the button, e.g. a validation error or
	// "3 pods will be restarted, the alert is updated in about 2 minutes". It is shown as plain text.
	// Maximum length: MaxWebhookResultMessageLength characters.
	Message string `json:"message"`

	// AfterClick optionally overrides [Webhook.AfterClick] for this click. Ignored when the request was rejected.
	// Each of AfterClick.HideWebhookIDs should match a webhook in the alert; unknown IDs are ignored, since the alert
	// is not known when the result is validated.
	AfterClick *WebhookAfterClick `json:"afterClick,omitempty"`
}

// Clean trims whitespace and applies defaults (AfterClick.HideButtons). It is a no-op on a nil receiver.
func (r *WebhookResult) Clean() {
	if r == nil {
		return
	}

	r.Message = strings.TrimSpace(r.Message)
	r.AfterClick.clean()
}

// Validate validates the result. Call [WebhookResult.Clean] first to trim whitespace and apply defaults.
func (r *WebhookResult) Validate() error {
	if r == nil {
		return errors.New("webhook result is nil")
	}

	if len(r.Message) > MaxWebhookResultMessageLength {
		return fmt.Errorf("message is too long, expected length <=%d", MaxWebhookResultMessageLength)
	}

	return r.AfterClick.validate("afterClick", nil)
}
