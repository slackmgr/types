package types

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// SlackChannelIDOrNameRegex matches valid Slack channel IDs and channel names.
	// Channel names are mapped to channel IDs by the API.
	SlackChannelIDOrNameRegex = regexp.MustCompile(fmt.Sprintf(`^[0-9a-zA-Z\-_]{1,%d}$`, MaxSlackChannelIDLength))

	// IconRegex matches valid Slack icon emojis, on the format ':emoji:'.
	IconRegex = regexp.MustCompile(fmt.Sprintf(`^:[^:]{1,%d}:$`, MaxIconEmojiLength))

	// SlackMentionRegex matches valid Slack mentions, such as <!here>, <!channel> and <@U12345678>.
	SlackMentionRegex = regexp.MustCompile(fmt.Sprintf(`^((<!here>)|(<!channel>)|(<@[^>\s]{1,%d}>))$`, MaxMentionLength))
)

const (
	// MaxTimestampAge is the maximum age of an alert timestamp.
	// If the timestamp is older than this, it will be replaced with the current time.
	MaxTimestampAge = 7 * 24 * time.Hour

	// Alert field length limits.
	// These constants define maximum lengths for various alert fields to ensure
	// compatibility with Slack's API limits and prevent excessive data storage.

	// MaxSlackChannelIDLength is the maximum length of a Slack channel ID or name.
	MaxSlackChannelIDLength = 80
	// MaxRouteKeyLength is the maximum length of a routing key for channel routing.
	MaxRouteKeyLength = 1000
	// MaxHeaderLength is the maximum length of the alert header (title).
	// Slack's header block limit is 150 characters; we reserve space for status emoji replacement.
	MaxHeaderLength = 130
	// MaxFallbackTextLength is the maximum length of the fallback notification text.
	MaxFallbackTextLength = 150
	// MaxTextLength is the maximum length of the alert text (body).
	MaxTextLength = 10000
	// MaxAuthorLength is the maximum length of the author field.
	MaxAuthorLength = 100
	// MaxHostLength is the maximum length of the host field.
	MaxHostLength = 100
	// MaxFooterLength is the maximum length of the footer field.
	MaxFooterLength = 300
	// MaxUsernameLength is the maximum length of the bot username.
	MaxUsernameLength = 100
	// MaxFieldTitleLength is the maximum length of a field title.
	MaxFieldTitleLength = 30
	// MaxFieldValueLength is the maximum length of a field value.
	MaxFieldValueLength = 200
	// MaxIconEmojiLength is the maximum length of the icon emoji (excluding colons).
	MaxIconEmojiLength = 50
	// MaxMentionLength is the maximum length of a Slack mention (excluding angle brackets).
	MaxMentionLength = 20
	// MaxCorrelationIDLength is the maximum length of the correlation ID.
	MaxCorrelationIDLength = 500

	// Auto-resolve timing limits.

	// MinAutoResolveSeconds is the minimum seconds before auto-resolving an issue.
	MinAutoResolveSeconds = 30
	// MaxAutoResolveSeconds is the maximum seconds before auto-resolving an issue (approximately 2 years).
	MaxAutoResolveSeconds = 63113851

	// IgnoreIfTextContains limits.

	// MaxIgnoreIfTextContainsLength is the maximum length of each ignore pattern.
	MaxIgnoreIfTextContainsLength = 1000
	// MaxIgnoreIfTextContainsCount is the maximum number of ignore patterns per alert.
	MaxIgnoreIfTextContainsCount = 20

	// MaxFieldCount is the maximum number of fields per alert.
	MaxFieldCount = 20

	// Webhook limits.
	// These constants define limits for webhook configurations.

	// MaxWebhookCount is the maximum number of webhooks per alert.
	MaxWebhookCount = 5
	// MaxWebhookIDLength is the maximum length of a webhook ID.
	MaxWebhookIDLength = 100
	// MaxWebhookURLLength is the maximum length of a webhook URL.
	MaxWebhookURLLength = 1000
	// MaxWebhookButtonTextLength is the maximum length of button text (Slack limit: 25 characters).
	MaxWebhookButtonTextLength = 25
	// MaxWebhookConfirmationTextLength is the maximum length of confirmation dialog text.
	MaxWebhookConfirmationTextLength = 1000
	// MaxWebhookPayloadCount is the maximum number of key-value pairs in webhook payload.
	MaxWebhookPayloadCount = 50
	// MaxWebhookPlainTextInputCount is the maximum number of text inputs per webhook.
	MaxWebhookPlainTextInputCount = 10
	// MaxWebhookCheckboxInputCount is the maximum number of checkbox groups per webhook.
	MaxWebhookCheckboxInputCount = 10
	// MaxWebhookInputIDLength is the maximum length of an input field ID.
	MaxWebhookInputIDLength = 200
	// MaxWebhookInputDescriptionLength is the maximum length of an input field description/placeholder.
	MaxWebhookInputDescriptionLength = 200
	// MaxWebhookInputLabelLength is the maximum length of a checkbox group label.
	MaxWebhookInputLabelLength = 200
	// MaxWebhookInputTextLength is the maximum length of text input content.
	MaxWebhookInputTextLength = 3000
	// MaxWebhookCheckboxOptionCount is the maximum number of options per checkbox group.
	MaxWebhookCheckboxOptionCount = 5
	// MaxWebhookCheckboxOptionTextLength is the maximum length of checkbox option text.
	MaxWebhookCheckboxOptionTextLength = 50
	// MaxCheckboxOptionValueLength is the maximum length of a checkbox option value.
	MaxCheckboxOptionValueLength = 100

	// Escalation limits.
	// These constants define limits for escalation configurations.

	// MaxEscalationCount is the maximum number of escalation points per alert.
	MaxEscalationCount = 3
	// MinEscalationDelaySeconds is the minimum delay before the first escalation triggers.
	MinEscalationDelaySeconds = 30
	// MinEscalationDelayDiffSeconds is the minimum time between consecutive escalations.
	MinEscalationDelayDiffSeconds = 30
	// MaxEscalationSlackMentionCount is the maximum number of Slack mentions per escalation.
	MaxEscalationSlackMentionCount = 10
)

// Alert represents a single alert that can be sent to the Slack Manager.
// Alerts with the same CorrelationID are grouped together in issues by the Slack Manager.
type Alert struct {
	// Timestamp is the time when the alert was created. If the timestamp is empty (or older than 7 days), it will be replaced with the current time.
	Timestamp time.Time `json:"timestamp"`

	// CorrelationID is an optional field used to group related alerts together in issues.
	// If unset, the correlation ID is constructed by hashing [Header, Text, Author, Host, SlackChannelID].
	// It is strongly recommended to set this to an explicit value, which makes sense in your context, rather than relying on the default hash value.
	// With a custom correlation ID, you can update both header and text without creating a new issue.
	CorrelationID string `json:"correlationId"`

	// Type is the type of alert, such as 'compliance', 'security' or 'metrics'.
	// It is primarily used for routing, when the alert RouteKey field is used (rather than SlackChannelID).
	// This field is optional, and case-insensitive.
	Type string `json:"type"`

	// Header is the main header (title) of the alert.
	// It is automatically truncated at MaxHeaderLength characters.
	// Include :status: in the header (or text) to have it replaced with the appropriate emoji for the issue severity.
	// This field is optional, but Header and Text cannot both be empty.
	Header string `json:"header"`

	// HeaderWhenResolved is the main header (title) of the issue when in the *resolved* state.
	// It is automatically truncated at MaxHeaderLength characters.
	// This field is optional. If unset, the Header field is used for all issue states.
	HeaderWhenResolved string `json:"headerWhenResolved"`

	// Text is the main text (body) of the alert.
	// It is automatically truncated at MaxTextLength characters.
	// Include :status: in the text (or header) to have it replaced with the appropriate emoji for the issue severity.
	// This field is optional, but Header and Text cannot both be empty.
	Text string `json:"text"`

	// TextWhenResolved is the main text (body) of the alert when in the *resolved* state.
	// It is automatically truncated at MaxTextLength characters.
	// This field is optional. If unset, the Text field is used for all issue states.
	TextWhenResolved string `json:"textWhenResolved"`

	// FallbackText is the text displayed in Slack notifications.
	// It should be a short, human-readable summary of the alert, without markdown or line breaks.
	// It is automatically truncated at MaxFallbackTextLength characters.
	// This field is optional. If unset, Slack decides what to display in notifications (which may not always be ideal).
	FallbackText string `json:"fallbackText"`

	// Author is the 'author' of the alert (if relevant), displayed as a context block in the Slack post.
	// It is automatically truncated at MaxAuthorLength characters.
	// This field is optional.
	Author string `json:"author"`

	// Host is the 'host' on which the alert originated (if any), displayed as a context block in the Slack post.
	// It is automatically truncated at MaxHostLength characters.
	// This field is optional.
	Host string `json:"host"`

	// Footer is the 'footer' of the alert, displayed as a context block at the bottom of the Slack post.
	// It is automatically truncated at MaxFooterLength characters.
	// This field is optional.
	Footer string `json:"footer"`

	// Link is an optional link (url) to more information about the alert, displayed as a context block in the Slack post.
	// This field is optional, but if set, it must be a valid absolute URL, starting with http:// or https://
	Link string `json:"link"`

	// IssueFollowUpEnabled is a flag that determines if the issue should be automatically resolved after a certain time.
	// If set to true, the issue will be resolved after AutoResolveSeconds seconds.
	// Set to false for fire-and-forget alerts, where no follow-up is needed (i.e. no issue tracking).
	IssueFollowUpEnabled bool `json:"issueFollowUpEnabled"`

	// AutoResolveSeconds is the number of seconds after which the issue should be automatically resolved, if IssueFollowUpEnabled is true.
	// The value must be between MinAutoResolveSeconds and MaxAutoResolveSeconds.
	AutoResolveSeconds int `json:"autoResolveSeconds"`

	// AutoResolveAsInconclusive is a flag that determines if the issue should be automatically resolved as 'inconclusive' instead of 'resolved'.
	// This affects the which emoji is used in the Slack post.
	// The default value is false, which means the issue is resolved with status 'resolved'.
	AutoResolveAsInconclusive bool `json:"autoResolveAsInconclusive"`

	// Severity is the severity of the alert, such as 'panic', 'error', 'warning', 'resolved' or 'info'.
	// This value determines the emoji used in the Slack post (for the :status: placeholder in header or text).
	// The value must be one of the predefined AlertSeverity constants.
	// If unset, the severity is automatically set to 'error'.
	Severity AlertSeverity `json:"severity"`

	// SlackChannelID is the ID of the Slack channel where the alert should be posted.
	// Slack channel names are also accepted, and are automatically converted to channel IDs by the API.
	// The value must be an existing channel ID or name, and the Slack Manager integration must have been added to the channel.
	// This field is optional.
	// If both SlackChannelID and RouteKey are set, SlackChannelID takes precedence.
	// If both SlackChannelID and RouteKey are empty, the API will still accept the alert IF a fallback mapping exists. Otherwise, it will return an error.
	SlackChannelID string `json:"slackChannelId"`

	// RouteKey is the case-insensitive route key of the alert, used for routing to the correct Slack channel by the API.
	// The API will return an error if the route key does not match any configured route.
	// This field is optional.
	// If both SlackChannelID and RouteKey are set, SlackChannelID takes precedence.
	// If both SlackChannelID and RouteKey are empty, the API will still accept the alert IF a fallback mapping exists. Otherwise, it will return an error.
	RouteKey string `json:"routeKey"`

	// Username is the username that the alert should be posted as in Slack.
	// This field is optional. If omitted, the alert is posted as the default bot user.
	Username string `json:"username"`

	// IconEmoji is the emoji that the alert should be posted with in Slack, on the format ':emoji:'.
	IconEmoji string `json:"iconEmoji"`

	// Fields are rendered in a compact format that allows for 2 columns of side-by-side text.
	Fields []*Field `json:"fields"`

	// NotificationDelaySeconds is the number of seconds to wait before creating an actual Slack post.
	// If the issue is resolved before the delay is over, no Slack post is created for the issue.
	// This is useful for issues that may be resolved quickly, to avoid unnecessary notifications.
	NotificationDelaySeconds int `json:"notificationDelaySeconds"`

	// ArchivingDelaySeconds is the number of seconds to wait before archiving the issue, after it is resolved.
	// A non-archived issue is re-opened if a new alert with the same CorrelationID is received, and the same Slack post is updated.
	// An archived issue can never be re-opened, and any new alerts with the same CorrelationID will generate a new issue and new Slack post.
	ArchivingDelaySeconds int `json:"archivingDelaySeconds"`

	// Escalation defines a list of escalation points for this alert's issue.
	// Each escalation can increase severity, add Slack mentions, or move the issue to a different channel after a specified delay.
	// Escalations are sorted by DelaySeconds and triggered in order if the issue remains unresolved.
	// Maximum of MaxEscalationCount escalations allowed.
	Escalation []*Escalation `json:"escalation"`

	// IgnoreIfTextContains is a list of substrings that, if found in the alert text, will cause the alert to be ignored.
	// This is useful for filtering out known noise or false positives.
	// Maximum of MaxIgnoreIfTextContainsCount items, each up to MaxIgnoreIfTextContainsLength characters.
	IgnoreIfTextContains []string `json:"ignoreIfTextContains"`

	// Webhooks defines interactive buttons that appear on the Slack post.
	// Each webhook triggers an HTTP POST to the specified URL when clicked.
	// Webhooks can include confirmation dialogs, input forms, and access level restrictions.
	// Maximum of MaxWebhookCount webhooks allowed.
	Webhooks []*Webhook `json:"webhooks"`

	// Metadata is an arbitrary key-value map for storing custom data with the alert.
	// This data is passed through to webhook payloads and can be used for tracking or correlation purposes.
	// The Slack Manager does not interpret this data.
	Metadata map[string]any `json:"metadata"`

	// Deprecated: FailOnRateLimitError is no longer in use.
	FailOnRateLimitError bool `json:"failOnRateLimitError"`
}

// Field is an alert field.
type Field struct {
	// Title is the title of the field. It is automatically truncated at MaxFieldTitleLength characters.
	Title string `json:"title"`

	// Value is the value of the field. It is automatically truncated at MaxFieldValueLength characters.
	Value string `json:"value"`
}

// Escalation represents an escalation point for an issue.
type Escalation struct {
	// Severity is the new severity of the issue, when the escalation is triggered.
	Severity AlertSeverity `json:"severity"`

	// DelaySeconds is the number of seconds since the issue was created (first alert received),
	// before the escalation is triggered.
	DelaySeconds int `json:"delaySeconds"`

	// SlackMentions is a list of Slack mentions that should be added to the Slack post when the escalation is triggered.
	SlackMentions []string `json:"slackMentions"`

	// MoveToChannel is the ID or name of the Slack channel where the alert should be moved when the escalation is triggered.
	MoveToChannel string `json:"moveToChannel"`
}

// Webhook represents an interactive button that appears on the Slack post.
// When clicked, it triggers an HTTP POST request to the specified URL (for http/https URLs),
// or invokes a custom webhook handler registered in the Slack Manager app.
type Webhook struct {
	// ID is the unique identifier for this webhook within the alert.
	// It must be unique among all webhooks in the same alert.
	// Maximum length: MaxWebhookIDLength characters.
	ID string `json:"id"`

	// URL specifies the target for the webhook when the button is clicked.
	// For HTTP webhooks, this must be a valid absolute URL starting with http:// or https://.
	// For custom webhook handlers registered in the Slack Manager app, this can be an arbitrary
	// ASCII string identifier that the handler recognizes.
	// The field name "URL" is retained for backwards compatibility.
	// Maximum length: MaxWebhookURLLength characters.
	URL string `json:"url"`

	// ConfirmationText is the text displayed in a confirmation dialog before triggering the webhook.
	// If empty, no confirmation dialog is shown and the webhook is triggered immediately.
	// Maximum length: MaxWebhookConfirmationTextLength characters.
	ConfirmationText string `json:"confirmationText"`

	// ButtonText is the label displayed on the button in Slack.
	// This field is required.
	// Maximum length: MaxWebhookButtonTextLength characters.
	ButtonText string `json:"buttonText"`

	// ButtonStyle determines the visual appearance of the button in Slack.
	// Valid values are defined by WebhookButtonStyle constants.
	// If empty, the default Slack button style is used.
	ButtonStyle WebhookButtonStyle `json:"buttonStyle"`

	// AccessLevel controls who can click this webhook button.
	// Valid values are defined by WebhookAccessLevel constants.
	// If empty, anyone in the channel can trigger the webhook.
	AccessLevel WebhookAccessLevel `json:"accessLevel"`

	// DisplayMode controls when the webhook button is visible.
	// Valid values are defined by WebhookDisplayMode constants.
	// If empty, the button is always visible.
	DisplayMode WebhookDisplayMode `json:"displayMode"`

	// Payload is a map of key-value pairs sent in the HTTP POST body when the webhook is triggered.
	// Alert metadata and input values are merged into this payload.
	// Maximum of MaxWebhookPayloadCount items.
	Payload map[string]any `json:"payload"`

	// PlainTextInput defines text input fields shown in the webhook's modal dialog.
	// User-entered values are included in the webhook payload.
	// Maximum of MaxWebhookPlainTextInputCount inputs.
	PlainTextInput []*WebhookPlainTextInput `json:"plainTextInput"`

	// CheckboxInput defines checkbox groups shown in the webhook's modal dialog.
	// Selected values are included in the webhook payload.
	// Maximum of MaxWebhookCheckboxInputCount inputs.
	CheckboxInput []*WebhookCheckboxInput `json:"checkboxInput"`
}

// WebhookPlainTextInput represents a text input field in a webhook's modal dialog.
// The user's input is included in the webhook payload with the field ID as the key.
type WebhookPlainTextInput struct {
	// ID is the unique identifier for this input field.
	// It must be unique among all inputs (both text and checkbox) in the same webhook.
	// The ID is used as the key in the webhook payload.
	// Maximum length: MaxWebhookInputIDLength characters.
	ID string `json:"id"`

	// Description is the placeholder text shown in the input field before the user types.
	// Maximum length: MaxWebhookInputDescriptionLength characters.
	Description string `json:"description"`

	// MinLength is the minimum number of characters required for the input.
	// Must be >= 0 and <= MaxLength.
	MinLength int `json:"minLength"`

	// MaxLength is the maximum number of characters allowed for the input.
	// Must be >= MinLength and <= MaxWebhookInputTextLength.
	MaxLength int `json:"maxLength"`

	// Multiline determines whether the input field allows multiple lines of text.
	// If true, a larger textarea is shown instead of a single-line input.
	Multiline bool `json:"multiline"`

	// InitialValue is the default text pre-filled in the input field.
	// Must satisfy the MinLength and MaxLength constraints.
	InitialValue string `json:"initialValue"`
}

// WebhookCheckboxInput represents a group of checkboxes in a webhook's modal dialog.
// Selected option values are included in the webhook payload as an array with the field ID as the key.
type WebhookCheckboxInput struct {
	// ID is the unique identifier for this checkbox group.
	// It must be unique among all inputs (both text and checkbox) in the same webhook.
	// The ID is used as the key in the webhook payload.
	// Maximum length: MaxWebhookInputIDLength characters.
	ID string `json:"id"`

	// Label is the text displayed above the checkbox group.
	// Maximum length: MaxWebhookInputLabelLength characters.
	Label string `json:"label"`

	// Options is the list of checkbox options available in this group.
	// Maximum of MaxWebhookCheckboxOptionCount options.
	Options []*WebhookCheckboxOption `json:"options"`
}

// WebhookCheckboxOption represents a single checkbox option within a WebhookCheckboxInput.
type WebhookCheckboxOption struct {
	// Value is the value included in the webhook payload when this option is selected.
	// Must be unique among all options in the same checkbox group.
	// Maximum length: MaxCheckboxOptionValueLength characters.
	Value string `json:"value"`

	// Text is the label displayed next to the checkbox.
	// Maximum length: MaxWebhookCheckboxOptionTextLength characters.
	Text string `json:"text"`

	// Selected determines whether this checkbox is pre-selected when the modal opens.
	Selected bool `json:"selected"`
}

// NewPanicAlert returns an alert with the severity set to 'panic'
func NewPanicAlert() *Alert {
	return NewAlert(AlertPanic)
}

// NewErrorAlert returns an alert with the severity set to 'error'
func NewErrorAlert() *Alert {
	return NewAlert(AlertError)
}

// NewWarningAlert returns an alert with the severity set to 'warning'
func NewWarningAlert() *Alert {
	return NewAlert(AlertWarning)
}

// NewResolvedAlert returns an alert with the severity set to 'resolved'
func NewResolvedAlert() *Alert {
	return NewAlert(AlertResolved)
}

// NewInfoAlert returns an alert with the severity set to 'info'
func NewInfoAlert() *Alert {
	return NewAlert(AlertInfo)
}

// NewAlert returns an alert with the specified severity
func NewAlert(severity AlertSeverity) *Alert {
	return &Alert{
		Timestamp: time.Now().UTC(),
		Severity:  severity,
		Metadata:  make(map[string]any),
	}
}

// UniqueID returns a unique and deterministic ID for this alert, for database/storage purposes.
// The ID is based on certain fields of the alert, and is base64 encoded to ensure it is safe for use in URLs and as a database key.
func (a *Alert) UniqueID() string {
	return hash("alert", a.SlackChannelID, a.RouteKey, a.CorrelationID, a.Timestamp.UTC().Format(time.RFC3339Nano), a.Header, a.Text)
}

// Clean normalizes and sanitizes all alert fields.
// It trims whitespace, normalizes case where appropriate, truncates certain fields that exceed maximum lengths,
// and applies default values for empty or invalid fields (e.g., sets Severity to 'error' if empty).
// This method should be called before validation to ensure consistent data.
func (a *Alert) Clean() {
	if time.Since(a.Timestamp) > MaxTimestampAge {
		a.Timestamp = time.Now()
	}

	a.CorrelationID = strings.TrimSpace(a.CorrelationID)

	a.Type = strings.ToLower(strings.TrimSpace(a.Type))

	a.SlackChannelID = strings.ToUpper(strings.TrimSpace(a.SlackChannelID))

	a.RouteKey = strings.ToLower(strings.TrimSpace(a.RouteKey))

	a.Header = strings.ReplaceAll(strings.TrimSpace(a.Header), "\n", " ")
	a.Header = truncateStringIfNeeded(a.Header, MaxHeaderLength)

	a.HeaderWhenResolved = strings.ReplaceAll(strings.TrimSpace(a.HeaderWhenResolved), "\n", " ")
	a.HeaderWhenResolved = truncateStringIfNeeded(a.HeaderWhenResolved, MaxHeaderLength)

	a.Text = strings.TrimSpace(a.Text)
	a.Text = shortenAlertTextIfNeeded(a.Text)

	a.TextWhenResolved = strings.TrimSpace(a.TextWhenResolved)
	a.TextWhenResolved = shortenAlertTextIfNeeded(a.TextWhenResolved)

	a.FallbackText = strings.TrimSpace(strings.ReplaceAll(a.FallbackText, ":status:", ""))
	a.FallbackText = strings.ReplaceAll(a.FallbackText, "\n", " ")
	a.FallbackText = truncateStringIfNeeded(a.FallbackText, MaxFallbackTextLength)

	a.Username = strings.TrimSpace(a.Username)
	a.Username = truncateStringIfNeeded(a.Username, MaxUsernameLength)

	a.Author = strings.TrimSpace(a.Author)
	a.Author = truncateStringIfNeeded(a.Author, MaxAuthorLength)

	a.Host = strings.TrimSpace(a.Host)
	a.Host = truncateStringIfNeeded(a.Host, MaxHostLength)

	a.Link = strings.TrimSpace(a.Link)

	a.Footer = strings.TrimSpace(a.Footer)
	a.Footer = truncateStringIfNeeded(a.Footer, MaxFooterLength)

	a.IconEmoji = strings.ToLower(strings.TrimSpace(a.IconEmoji))

	a.Severity = AlertSeverity(strings.ToLower(strings.TrimSpace(string(a.Severity))))

	if a.Severity == "" || a.Severity == "critical" {
		a.Severity = AlertError
	}

	if a.Severity == "resolve" || a.Severity == "recovered" || a.Severity == "recover" {
		a.Severity = AlertResolved
	}

	if a.ArchivingDelaySeconds < 0 {
		a.ArchivingDelaySeconds = 0
	}

	if a.NotificationDelaySeconds < 0 {
		a.NotificationDelaySeconds = 0
	}

	for _, field := range a.Fields {
		if field == nil {
			continue
		}

		field.Title = strings.TrimSpace(field.Title)
		field.Title = truncateStringIfNeeded(field.Title, MaxFieldTitleLength)

		field.Value = strings.TrimSpace(field.Value)
		field.Value = truncateStringIfNeeded(field.Value, MaxFieldValueLength)
	}

	for _, hook := range a.Webhooks {
		if hook == nil {
			continue
		}

		hook.ID = strings.TrimSpace(hook.ID)
		hook.ButtonText = strings.TrimSpace(hook.ButtonText)
		hook.URL = strings.TrimSpace(hook.URL)
		hook.ConfirmationText = strings.TrimSpace(hook.ConfirmationText)

		if hook.ButtonStyle == "default" {
			hook.ButtonStyle = ""
		}

		for _, input := range hook.PlainTextInput {
			if input == nil {
				continue
			}

			input.ID = strings.TrimSpace(input.ID)
			input.Description = strings.TrimSpace(input.Description)
			input.InitialValue = strings.TrimSpace(input.InitialValue)
		}

		for _, input := range hook.CheckboxInput {
			if input == nil {
				continue
			}

			input.ID = strings.TrimSpace(input.ID)
			input.Label = strings.TrimSpace(input.Label)
		}
	}

	if len(a.Escalation) > 0 {
		sort.Slice(a.Escalation, func(i, j int) bool {
			if a.Escalation[i] == nil {
				return true
			}
			if a.Escalation[j] == nil {
				return false
			}
			return a.Escalation[i].DelaySeconds < a.Escalation[j].DelaySeconds
		})

		for _, e := range a.Escalation {
			if e == nil {
				continue
			}

			e.Severity = AlertSeverity(strings.ToLower(strings.TrimSpace(string(e.Severity))))
			e.MoveToChannel = strings.ToUpper(strings.TrimSpace(e.MoveToChannel))

			for i, mention := range e.SlackMentions {
				e.SlackMentions[i] = strings.TrimSpace(mention)
			}
		}
	}
}

// Validate returns an error if one or more of the required fields are empty or invalid
func (a *Alert) Validate() error {
	if a == nil {
		return errors.New("alert is nil")
	}

	if err := a.ValidateSlackChannelIDAndRouteKey(); err != nil {
		return err
	}

	if err := a.ValidateHeaderAndText(); err != nil {
		return err
	}

	if err := a.ValidateIcon(); err != nil {
		return err
	}

	if err := a.ValidateLink(); err != nil {
		return err
	}

	if err := a.ValidateSeverity(); err != nil {
		return err
	}

	if err := a.ValidateCorrelationID(); err != nil {
		return err
	}

	if err := a.ValidateAutoResolve(); err != nil {
		return err
	}

	if err := a.ValidateFields(); err != nil {
		return err
	}

	if err := a.ValidateWebhooks(); err != nil {
		return err
	}

	if err := a.ValidateEscalation(); err != nil {
		return err
	}

	return a.ValidateIgnoreIfTextContains()
}

// ValidateSlackChannelIDAndRouteKey validates that SlackChannelID and RouteKey are valid, if set.
// Both values are allowed to be empty (in which case a fallback mapping must exist in the API).
func (a *Alert) ValidateSlackChannelIDAndRouteKey() error {
	if a.SlackChannelID != "" {
		if !SlackChannelIDOrNameRegex.MatchString(a.SlackChannelID) {
			return fmt.Errorf("slackChannelId '%s' is not valid", a.SlackChannelID)
		}

		return nil
	}

	if len(a.RouteKey) > MaxRouteKeyLength {
		return fmt.Errorf("routeKey is too long, expected length <=%d", MaxRouteKeyLength)
	}

	return nil
}

// ValidateHeaderAndText validates that at least one of Header or Text is non-empty.
// An alert must have either a header or text content to be meaningful.
func (a *Alert) ValidateHeaderAndText() error {
	if a.Header == "" && a.Text == "" {
		return errors.New("header and text cannot both be empty")
	}

	return nil
}

// ValidateIcon validates that IconEmoji, if set, matches the expected Slack emoji format ':emoji:'.
func (a *Alert) ValidateIcon() error {
	if a.IconEmoji == "" {
		return nil
	}

	if !IconRegex.MatchString(a.IconEmoji) {
		return fmt.Errorf("iconEmoji '%s' is not valid", a.IconEmoji)
	}

	return nil
}

// ValidateLink validates that Link, if set, is a valid absolute URL with a scheme.
func (a *Alert) ValidateLink() error {
	if a.Link == "" {
		return nil
	}

	url, err := url.ParseRequestURI(a.Link)
	if err != nil {
		return errors.New("link is not a valid absolute URL")
	}

	if url.Scheme == "" {
		return errors.New("link is not a valid absolute URL")
	}

	return nil
}

// ValidateSeverity validates that Severity is one of the allowed AlertSeverity values.
func (a *Alert) ValidateSeverity() error {
	if !SeverityIsValid(a.Severity) {
		return fmt.Errorf("severity '%s' is not valid, expected one of [%s]", a.Severity, strings.Join(ValidSeverities(), ", "))
	}

	return nil
}

// ValidateCorrelationID validates that CorrelationID, if set, does not exceed MaxCorrelationIDLength.
func (a *Alert) ValidateCorrelationID() error {
	if a.CorrelationID == "" {
		return nil
	}

	if len(a.CorrelationID) > MaxCorrelationIDLength {
		return fmt.Errorf("correlationId is too long, expected length <=%d", MaxCorrelationIDLength)
	}

	return nil
}

// ValidateAutoResolve validates that AutoResolveSeconds is within the allowed range
// when IssueFollowUpEnabled is true.
func (a *Alert) ValidateAutoResolve() error {
	if !a.IssueFollowUpEnabled {
		return nil
	}

	if a.AutoResolveSeconds < MinAutoResolveSeconds {
		return fmt.Errorf("autoResolveSeconds %d is too low, expected value >=%d", a.AutoResolveSeconds, MinAutoResolveSeconds)
	}

	if a.AutoResolveSeconds > MaxAutoResolveSeconds {
		return fmt.Errorf("autoResolveSeconds %d is too high, expected value <=%d", a.AutoResolveSeconds, MaxAutoResolveSeconds)
	}

	return nil
}

// ValidateIgnoreIfTextContains validates that the IgnoreIfTextContains slice
// does not exceed the maximum count and that each item does not exceed the maximum length.
func (a *Alert) ValidateIgnoreIfTextContains() error {
	if len(a.IgnoreIfTextContains) == 0 {
		return nil
	}

	if len(a.IgnoreIfTextContains) > MaxIgnoreIfTextContainsCount {
		return fmt.Errorf("too many ignoreIfTextContains items, expected <=%d", MaxIgnoreIfTextContainsCount)
	}

	for index, s := range a.IgnoreIfTextContains {
		if len(s) > MaxIgnoreIfTextContainsLength {
			return fmt.Errorf("ignoreIfTextContains[%d] is too long, expected length <=%d", index, MaxIgnoreIfTextContainsLength)
		}
	}

	return nil
}

// ValidateFields validates that the number of fields does not exceed MaxFieldCount.
func (a *Alert) ValidateFields() error {
	if len(a.Fields) > MaxFieldCount {
		return fmt.Errorf("too many fields, expected <=%d", MaxFieldCount)
	}

	return nil
}

// ValidateWebhooks validates all webhooks in the alert.
// It checks that the webhook count is within limits, all required fields are present,
// URLs are valid, IDs are unique, and all nested inputs are properly configured.
func (a *Alert) ValidateWebhooks() error {
	if a.Webhooks == nil {
		return nil
	}

	if len(a.Webhooks) > MaxWebhookCount {
		return fmt.Errorf("too many webhooks, expected <=%d", MaxWebhookCount)
	}

	webhookIDs := make(map[string]struct{})

	for index, hook := range a.Webhooks {
		if hook == nil {
			return fmt.Errorf("webhook[%d] is nil", index)
		}

		if hook.ID == "" {
			return fmt.Errorf("webhook[%d].id is required", index)
		}

		if len(hook.ID) > MaxWebhookIDLength {
			return fmt.Errorf("webhook[%d].id is too long, expected length <=%d", index, MaxWebhookIDLength)
		}

		if _, ok := webhookIDs[hook.ID]; ok {
			return fmt.Errorf("webhook[%d].id must be unique", index)
		}

		webhookIDs[hook.ID] = struct{}{}

		if hook.URL == "" {
			return fmt.Errorf("webhook[%d].url is required", index)
		}

		if len(hook.URL) > MaxWebhookURLLength {
			return fmt.Errorf("webhook[%d].url is too long, expected length <=%d", index, MaxWebhookURLLength)
		}

		// For HTTP URLs, validate as absolute URL. For custom handler identifiers, validate as ASCII.
		if strings.HasPrefix(strings.ToLower(hook.URL), "http") {
			parsedURL, err := url.ParseRequestURI(hook.URL)
			if err != nil {
				return fmt.Errorf("webhook[%d].url is not a valid absolute URL", index)
			}

			if parsedURL.Scheme == "" || parsedURL.Host == "" {
				return fmt.Errorf("webhook[%d].url is not a valid absolute URL", index)
			}
		} else if !isValidASCII(hook.URL) {
			return fmt.Errorf("webhook[%d].url contains invalid characters, expected printable ASCII", index)
		}

		if hook.ButtonText == "" {
			return fmt.Errorf("webhook[%d].buttonText is required", index)
		}

		if len(hook.ButtonText) > MaxWebhookButtonTextLength {
			return fmt.Errorf("webhook[%d].buttonText is too long, expected length <=%d", index, MaxWebhookButtonTextLength)
		}

		if len(hook.ConfirmationText) > MaxWebhookConfirmationTextLength {
			return fmt.Errorf("webhook[%d].confirmationText is too long, expected length <=%d", index, MaxWebhookConfirmationTextLength)
		}

		if hook.ButtonStyle != "" && !WebhookButtonStyleIsValid(hook.ButtonStyle) {
			return fmt.Errorf("webhook[%d].buttonStyle '%s' is not valid, expected empty or one of [%s]", index, hook.ButtonStyle, strings.Join(ValidWebhookButtonStyles(), ", "))
		}

		if hook.AccessLevel != "" && !WebhookAccessLevelIsValid(hook.AccessLevel) {
			return fmt.Errorf("webhook[%d].accessLevel '%s' is not valid, expected empty or one of [%s]", index, hook.AccessLevel, strings.Join(ValidWebhookAccessLevels(), ", "))
		}

		if hook.DisplayMode != "" && !WebhookDisplayModeIsValid(hook.DisplayMode) {
			return fmt.Errorf("webhook[%d].displayMode '%s' is not valid, expected empty or one of [%s]", index, hook.DisplayMode, strings.Join(ValidWebhookDisplayModes(), ", "))
		}

		if len(hook.Payload) > MaxWebhookPayloadCount {
			return fmt.Errorf("webhook[%d].payload item count is too large, expected <=%d", index, MaxWebhookPayloadCount)
		}

		if len(hook.PlainTextInput) > MaxWebhookPlainTextInputCount {
			return fmt.Errorf("webhook[%d].plainTextInput item count is too large, expected <=%d", index, MaxWebhookPlainTextInputCount)
		}

		if len(hook.CheckboxInput) > MaxWebhookCheckboxInputCount {
			return fmt.Errorf("webhook[%d].checkboxInput item count is too large, expected <=%d", index, MaxWebhookCheckboxInputCount)
		}

		inputIDs := make(map[string]struct{})

		for inputIndex, input := range hook.PlainTextInput {
			if input == nil {
				return fmt.Errorf("webhook[%d].plainTextInput[%d] is nil", index, inputIndex)
			}

			if input.ID == "" {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id is required", index, inputIndex)
			}

			if _, ok := inputIDs[input.ID]; ok {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id must be unique among all inputs", index, inputIndex)
			}

			inputIDs[input.ID] = struct{}{}

			if len(input.ID) > MaxWebhookInputIDLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id is too long, expected <=%d", index, inputIndex, MaxWebhookInputIDLength)
			}

			if len(input.Description) > MaxWebhookInputDescriptionLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].description is too long, expected <=%d", index, inputIndex, MaxWebhookInputDescriptionLength)
			}

			if input.MinLength < 0 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].minLength must be >=0", index, inputIndex)
			}

			if input.MinLength > MaxWebhookInputTextLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].minLength must be <=%d", index, inputIndex, MaxWebhookInputTextLength)
			}

			if input.MaxLength < 0 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength must be >=0", index, inputIndex)
			}

			if input.MaxLength > MaxWebhookInputTextLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength must be <=%d", index, inputIndex, MaxWebhookInputTextLength)
			}

			if input.MaxLength < input.MinLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength cannot be smaller than minLength", index, inputIndex)
			}

			if len(input.InitialValue) > input.MaxLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].initialValue cannot be longer than maxLength", index, inputIndex)
			}

			if len(input.InitialValue) < input.MinLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].initialValue cannot be shorter than minLength", index, inputIndex)
			}
		}

		for inputIndex, input := range hook.CheckboxInput {
			if input == nil {
				return fmt.Errorf("webhook[%d].checkboxInput[%d] is nil", index, inputIndex)
			}

			if input.ID == "" {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id is required", index, inputIndex)
			}

			if _, ok := inputIDs[input.ID]; ok {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id must be unique among all inputs", index, inputIndex)
			}

			inputIDs[input.ID] = struct{}{}

			if len(input.ID) > MaxWebhookInputIDLength {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id is too long, expected <=%d", index, inputIndex, MaxWebhookInputIDLength)
			}

			if len(input.Label) > MaxWebhookInputLabelLength {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].label is too long, expected <=%d", index, inputIndex, MaxWebhookInputLabelLength)
			}

			if len(input.Options) > MaxWebhookCheckboxOptionCount {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].options item count is too large, expected <=%d", index, inputIndex, MaxWebhookCheckboxOptionCount)
			}

			values := make(map[string]struct{})

			for optionIndex, option := range input.Options {
				if option == nil {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d] is nil", index, inputIndex, optionIndex)
				}

				if option.Value == "" {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value is required", index, inputIndex, optionIndex)
				}

				if len(option.Value) > MaxCheckboxOptionValueLength {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value is too long, expected <=%d", index, inputIndex, optionIndex, MaxCheckboxOptionValueLength)
				}

				if _, ok := values[option.Value]; ok {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value must be unique", index, inputIndex, optionIndex)
				}

				values[option.Value] = struct{}{}

				if len(option.Text) > MaxWebhookCheckboxOptionTextLength {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].text is too long, expected <=%d", index, inputIndex, optionIndex, MaxWebhookCheckboxOptionTextLength)
				}
			}
		}
	}

	return nil
}

// ValidateEscalation validates all escalation points in the alert.
// It checks that the escalation count is within limits, delays are properly spaced,
// severities are valid for escalation, and Slack mentions and channels are valid.
func (a *Alert) ValidateEscalation() error {
	if a.Escalation == nil {
		return nil
	}

	if len(a.Escalation) > MaxEscalationCount {
		return fmt.Errorf("too many escalation points, expected <=%d", MaxEscalationCount)
	}

	previousDelay := 0

	for index, e := range a.Escalation {
		if e == nil {
			return fmt.Errorf("escalation[%d] is nil", index)
		}

		if e.DelaySeconds < MinEscalationDelaySeconds {
			return fmt.Errorf("escalation[%d].delaySeconds '%d' is too low, expected value >=%d", index, e.DelaySeconds, MinEscalationDelaySeconds)
		}

		if previousDelay > 0 && e.DelaySeconds-previousDelay < MinEscalationDelayDiffSeconds {
			return fmt.Errorf("escalation[%d].delaySeconds '%d' is too small compared to previous escalation, expected diff >=%d", index, e.DelaySeconds, MinEscalationDelayDiffSeconds)
		}

		previousDelay = e.DelaySeconds

		if e.Severity != AlertPanic && e.Severity != AlertError && e.Severity != AlertWarning {
			return fmt.Errorf("escalation[%d].severity '%s' is not valid, expected one of [panic, error, warning]", index, e.Severity)
		}

		if len(e.SlackMentions) > MaxEscalationSlackMentionCount {
			return fmt.Errorf("escalation[%d].slackMentions item count is too large, expected <=%d", index, MaxEscalationSlackMentionCount)
		}

		for j, mention := range e.SlackMentions {
			if !SlackMentionRegex.MatchString(mention) {
				return fmt.Errorf("escalation[%d].slackMentions[%d] is not valid", index, j)
			}
		}

		if e.MoveToChannel != "" {
			if !SlackChannelIDOrNameRegex.MatchString(e.MoveToChannel) {
				return fmt.Errorf("escalation[%d].moveToChannel is not valid", index)
			}
		}
	}

	return nil
}

func shortenAlertTextIfNeeded(text string) string {
	if utf8.RuneCountInString(text) <= MaxTextLength {
		return text
	}

	endsWithCodeBlock := strings.HasSuffix(text, "```")

	if endsWithCodeBlock {
		return strings.TrimSpace(truncateString(text, MaxTextLength-6)) + "...```"
	}

	return strings.TrimSpace(truncateString(text, MaxTextLength-3)) + "..."
}

// truncateStringIfNeeded truncates s to maxLen runes if it exceeds that limit, appending "...".
// Trailing whitespace is trimmed from the truncated portion before appending.
// If s is within maxLen runes, it is returned unchanged.
func truncateStringIfNeeded(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	return strings.TrimSpace(truncateString(s, maxLen-3)) + "..."
}

// truncateString truncates a string to maxRunes runes, safely handling multi-byte UTF-8 characters.
func truncateString(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}

	runes := []rune(s)

	return string(runes[:maxRunes])
}

// isValidASCII returns true if the string contains only printable ASCII characters (0x20-0x7E).
func isValidASCII(s string) bool {
	for i := range len(s) {
		if s[i] < 0x20 || s[i] > 0x7E {
			return false
		}
	}
	return true
}

func hash(input ...string) string {
	h := sha256.New()

	delimiter := []byte{0}

	for _, s := range input {
		h.Write([]byte(s))
		h.Write(delimiter)
	}

	bs := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(bs)
}
