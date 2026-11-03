package types

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Webhook after-click limits.
const (
	// MaxWebhookAfterClickStatusTextLength is the maximum length of [WebhookAfterClick.StatusText].
	MaxWebhookAfterClickStatusTextLength = 200

	// DefaultWebhookAfterClickResetSeconds is the reset delay used when [WebhookAfterClick.ResetAfterSeconds] is 0.
	DefaultWebhookAfterClickResetSeconds = 300

	// MaxWebhookAfterClickResetSeconds is the maximum value of [WebhookAfterClick.ResetAfterSeconds] (24 hours).
	MaxWebhookAfterClickResetSeconds = 86400
)

// WebhookHideButtons controls which webhook buttons are hidden after a webhook was clicked and delivered.
// See [WebhookAfterClick.HideButtons].
type WebhookHideButtons string

const (
	// WebhookHideButtonsNone hides no buttons, except those listed in [WebhookAfterClick.HideWebhookIDs].
	WebhookHideButtonsNone WebhookHideButtons = "none"

	// WebhookHideButtonsClicked hides the clicked button, plus those listed in [WebhookAfterClick.HideWebhookIDs].
	WebhookHideButtonsClicked WebhookHideButtons = "clicked"

	// WebhookHideButtonsAll hides all webhook buttons on the alert.
	WebhookHideButtonsAll WebhookHideButtons = "all"
)

// WebhookHideButtonsIsValid returns true if the provided [WebhookHideButtons] is valid.
func WebhookHideButtonsIsValid(s WebhookHideButtons) bool {
	switch s {
	case WebhookHideButtonsNone, WebhookHideButtonsClicked, WebhookHideButtonsAll:
		return true
	}
	return false
}

// ValidWebhookHideButtons returns a slice of valid [WebhookHideButtons] values.
func ValidWebhookHideButtons() []string {
	return []string{
		string(WebhookHideButtonsNone),
		string(WebhookHideButtonsClicked),
		string(WebhookHideButtonsAll),
	}
}

// WebhookAfterClick describes how the Slack post changes after a webhook button was clicked and the webhook was
// delivered, while the receiving system is working on the request. It gives everyone in the channel immediate
// feedback, and prevents duplicate clicks.
//
// The post-click state is temporary. It is cleared when the next alert for the issue is received (even if the alert
// is otherwise unchanged or ignored), or when ResetAfterSeconds has passed without an alert. On expiry, the buttons
// are shown again with a note that no update was received. Nothing changes if the webhook could not be delivered;
// the user who clicked is informed instead.
type WebhookAfterClick struct {
	// HideButtons controls which webhook buttons are hidden.
	// Valid values are defined by WebhookHideButtons constants.
	// If empty, defaults to WebhookHideButtonsNone.
	HideButtons WebhookHideButtons `json:"hideButtons"`

	// HideWebhookIDs lists additional webhooks (by ID) whose buttons are hidden.
	// Each ID must match a webhook in the same alert. Ignored when HideButtons is WebhookHideButtonsAll.
	// Maximum of MaxWebhookCount items.
	HideWebhookIDs []string `json:"hideWebhookIds"`

	// StatusText is an optional text shown on the Slack post, e.g. "Restart in progress, the alert will be updated
	// when done". The clicking user and time are added by Slack Manager.
	// Maximum length: MaxWebhookAfterClickStatusTextLength characters.
	StatusText string `json:"statusText"`

	// ResetAfterSeconds is the maximum time the post-click state is kept if no new alert is received.
	// If 0, defaults to DefaultWebhookAfterClickResetSeconds.
	// Must be >=0 and <=MaxWebhookAfterClickResetSeconds.
	ResetAfterSeconds int `json:"resetAfterSeconds"`
}

// ResetAfter returns the reset delay, i.e. ResetAfterSeconds, or [DefaultWebhookAfterClickResetSeconds] if it is 0.
func (w *WebhookAfterClick) ResetAfter() time.Duration {
	if w.ResetAfterSeconds <= 0 {
		return DefaultWebhookAfterClickResetSeconds * time.Second
	}

	return time.Duration(w.ResetAfterSeconds) * time.Second
}

// HidesWebhook returns true if the button of the webhook with ID webhookID is hidden after the webhook with ID
// clickedWebhookID was clicked.
func (w *WebhookAfterClick) HidesWebhook(clickedWebhookID, webhookID string) bool {
	if w.HideButtons == WebhookHideButtonsAll {
		return true
	}

	if w.HideButtons == WebhookHideButtonsClicked && webhookID == clickedWebhookID {
		return true
	}

	return slices.Contains(w.HideWebhookIDs, webhookID)
}

// clean trims whitespace and applies the default HideButtons value. It is a no-op on a nil receiver.
func (w *WebhookAfterClick) clean() {
	if w == nil {
		return
	}

	if w.HideButtons == "" {
		w.HideButtons = WebhookHideButtonsNone
	}

	w.StatusText = strings.TrimSpace(w.StatusText)

	for i, id := range w.HideWebhookIDs {
		w.HideWebhookIDs[i] = strings.TrimSpace(id)
	}
}

// validate validates the after-click settings. field is the field path used in error messages.
// If webhookIDs is non-nil, each of HideWebhookIDs must be a key in it. Otherwise, each ID must be non-empty and
// at most MaxWebhookIDLength characters, since the webhooks of the alert are not known (e.g. in a [WebhookResult]).
// It is a no-op on a nil receiver.
func (w *WebhookAfterClick) validate(field string, webhookIDs map[string]struct{}) error {
	if w == nil {
		return nil
	}

	if w.HideButtons != "" && !WebhookHideButtonsIsValid(w.HideButtons) {
		return fmt.Errorf("%s.hideButtons '%s' is not valid, expected empty or one of [%s]", field, w.HideButtons, strings.Join(ValidWebhookHideButtons(), ", "))
	}

	if len(w.HideWebhookIDs) > MaxWebhookCount {
		return fmt.Errorf("%s.hideWebhookIds item count is too large, expected <=%d", field, MaxWebhookCount)
	}

	for idIndex, id := range w.HideWebhookIDs {
		if webhookIDs != nil {
			if _, ok := webhookIDs[id]; !ok {
				return fmt.Errorf("%s.hideWebhookIds[%d] '%s' does not match any webhook id in the alert", field, idIndex, id)
			}

			continue
		}

		if id == "" {
			return fmt.Errorf("%s.hideWebhookIds[%d] is required", field, idIndex)
		}

		if len(id) > MaxWebhookIDLength {
			return fmt.Errorf("%s.hideWebhookIds[%d] is too long, expected length <=%d", field, idIndex, MaxWebhookIDLength)
		}
	}

	if len(w.StatusText) > MaxWebhookAfterClickStatusTextLength {
		return fmt.Errorf("%s.statusText is too long, expected length <=%d", field, MaxWebhookAfterClickStatusTextLength)
	}

	if w.ResetAfterSeconds < 0 || w.ResetAfterSeconds > MaxWebhookAfterClickResetSeconds {
		return fmt.Errorf("%s.resetAfterSeconds must be between 0 and %d", field, MaxWebhookAfterClickResetSeconds)
	}

	return nil
}
