package common

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
	// MaxTimestampAge is the maximum age of an alert timestamp. If the timestamp is older than this, it will be replaced with the current time.
	MaxTimestampAge = 7 * 24 * time.Hour

	MaxSlackChannelIDLength       = 80
	MaxRouteKeyLength             = 1000
	MaxHeaderLength               = 130
	MaxFallbackTextLength         = 150
	MaxTextLength                 = 10000
	MaxAuthorLength               = 100
	MaxHostLength                 = 100
	MaxFooterLength               = 300
	MaxUsernameLength             = 100
	MaxFieldTitleLength           = 30
	MaxFieldValueLength           = 200
	MaxIconEmojiLength            = 50
	MaxMentionLength              = 20
	MaxCorrelationIDLength        = 500
	MinAutoResolveSeconds         = 30
	MaxAutoResolveSeconds         = 63113851 // 2 years
	MaxIgnoreIfTextContainsLength = 1000
	MaxFieldCount                 = 20
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
	// This field is optional, but SlackChannelID and RouteKey cannot both be empty.
	SlackChannelID string `json:"slackChannelId"`

	// RouteKey is the case-insensitive route key of the alert, used for routing to the correct Slack channel by the API.
	// The API will return an error if the route key does not match any configured route.
	// This field is ignored if SlackChannelID is set.
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
	// An archived can never be re-opened, and any new alerts with the same CorrelationID will generate a new issue and new Slack post.
	ArchivingDelaySeconds int                    `json:"archivingDelaySeconds"`
	Escalation            []*Escalation          `json:"escalation"`
	IgnoreIfTextContains  []string               `json:"ignoreIfTextContains"`
	FailOnRateLimitError  bool                   `json:"failOnRateLimitError"`
	Webhooks              []*Webhook             `json:"webhooks"`
	Metadata              map[string]interface{} `json:"metadata"`
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

type Webhook struct {
	ID               string                   `json:"id"`
	URL              string                   `json:"url"`
	ConfirmationText string                   `json:"confirmationText"`
	ButtonText       string                   `json:"buttonText"`
	ButtonStyle      WebhookButtonStyle       `json:"buttonStyle"`
	AccessLevel      WebhookAccessLevel       `json:"accessLevel"`
	DisplayMode      WebhookDisplayMode       `json:"displayMode"`
	Payload          map[string]interface{}   `json:"payload"`
	PlainTextInput   []*WebhookPlainTextInput `json:"plainTextInput"`
	CheckboxInput    []*WebhookCheckboxInput  `json:"checkboxInput"`
}

type WebhookPlainTextInput struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	MinLength    int    `json:"minLength"`
	MaxLength    int    `json:"maxLength"`
	Multiline    bool   `json:"multiline"`
	InitialValue string `json:"initialValue"`
}

type WebhookCheckboxInput struct {
	ID      string                   `json:"id"`
	Label   string                   `json:"label"`
	Options []*WebhookCheckboxOption `json:"options"`
}

type WebhookCheckboxOption struct {
	Value    string `json:"value"`
	Text     string `json:"text"`
	Selected bool   `json:"selected"`
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
	return hash("alert", a.SlackChannelID, a.RouteKey, a.CorrelationID, a.Timestamp.Format(time.RFC3339Nano), a.Header, a.Text)
}

func (a *Alert) Clean() {
	if time.Since(a.Timestamp) > MaxTimestampAge {
		a.Timestamp = time.Now()
	}

	a.Type = strings.ToLower(strings.TrimSpace(a.Type))
	a.SlackChannelID = strings.ToUpper(strings.TrimSpace(a.SlackChannelID))
	a.RouteKey = strings.ToLower(strings.TrimSpace(a.RouteKey))
	a.Header = strings.ReplaceAll(strings.TrimSpace(a.Header), "\n", " ")
	a.HeaderWhenResolved = strings.ReplaceAll(strings.TrimSpace(a.HeaderWhenResolved), "\n", " ")
	a.Text = strings.TrimSpace(a.Text)
	a.TextWhenResolved = strings.TrimSpace(a.TextWhenResolved)
	a.FallbackText = strings.TrimSpace(strings.ReplaceAll(a.FallbackText, ":status:", ""))
	a.FallbackText = strings.ReplaceAll(a.FallbackText, "\n", " ")
	a.CorrelationID = strings.TrimSpace(a.CorrelationID)
	a.Username = strings.TrimSpace(a.Username)
	a.Author = strings.TrimSpace(a.Author)
	a.Host = strings.TrimSpace(a.Host)
	a.Link = strings.TrimSpace(a.Link)
	a.Footer = strings.TrimSpace(a.Footer)
	a.IconEmoji = strings.ToLower(strings.TrimSpace(a.IconEmoji))
	a.Severity = AlertSeverity(strings.ToLower(strings.TrimSpace(string(a.Severity))))

	if len(a.FallbackText) > MaxFallbackTextLength {
		a.FallbackText = a.FallbackText[:MaxFallbackTextLength-3] + "..."
	}

	if a.Severity == "" || a.Severity == "critical" {
		a.Severity = AlertError
	}

	if a.ArchivingDelaySeconds < 0 {
		a.ArchivingDelaySeconds = 0
	}

	if a.NotificationDelaySeconds < 0 {
		a.NotificationDelaySeconds = 0
	}

	// Max length in the Slack API is 150, see https://api.slack.com/reference/block-kit/blocks#header
	// We also need to leave some space for the :status: emoji to be replaced with something a bit longer by the Slack Manager
	if len(a.Header) > MaxHeaderLength {
		a.Header = strings.TrimSpace(a.Header[:MaxHeaderLength-3]) + "..."
	}

	if len(a.HeaderWhenResolved) > MaxHeaderLength {
		a.HeaderWhenResolved = strings.TrimSpace(a.HeaderWhenResolved[:MaxHeaderLength-3]) + "..."
	}

	a.Text = shortenAlertTextIfNeeded(a.Text)
	a.TextWhenResolved = shortenAlertTextIfNeeded(a.TextWhenResolved)

	if len(a.Author) > MaxAuthorLength {
		a.Author = strings.TrimSpace(a.Author[:MaxAuthorLength-3]) + "..."
	}

	if len(a.Host) > MaxHostLength {
		a.Host = strings.TrimSpace(a.Host[:MaxHostLength-3]) + "..."
	}

	if len(a.Username) > MaxUsernameLength {
		a.Username = strings.TrimSpace(a.Username[:MaxUsernameLength-3]) + "..."
	}

	if len(a.Footer) > MaxFooterLength {
		a.Footer = strings.TrimSpace(a.Footer[:MaxFooterLength-3]) + "..."
	}

	for _, field := range a.Fields {
		field.Title = strings.TrimSpace(field.Title)
		field.Value = strings.TrimSpace(field.Value)

		if len(field.Title) > MaxFieldTitleLength {
			field.Title = strings.TrimSpace(field.Title[:MaxFieldTitleLength-3]) + "..."
		}

		if len(field.Value) > MaxFieldValueLength {
			field.Value = strings.TrimSpace(field.Value[:MaxFieldValueLength-3]) + "..."
		}
	}

	for _, hook := range a.Webhooks {
		hook.ID = strings.TrimSpace(hook.ID)
		hook.ButtonText = strings.TrimSpace(hook.ButtonText)
		hook.URL = strings.TrimSpace(hook.URL)
		hook.ConfirmationText = strings.TrimSpace(hook.ConfirmationText)

		if hook.ButtonStyle == "default" {
			hook.ButtonStyle = ""
		}

		for _, input := range hook.PlainTextInput {
			input.ID = strings.TrimSpace(input.ID)
			input.Description = strings.TrimSpace(input.Description)
		}
	}

	if len(a.Escalation) > 0 {
		sort.Slice(a.Escalation, func(i, j int) bool {
			return a.Escalation[i].DelaySeconds < a.Escalation[j].DelaySeconds
		})

		for _, e := range a.Escalation {
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

func (a *Alert) ValidateHeaderAndText() error {
	if a.Header == "" && a.Text == "" {
		return errors.New("header and text cannot both be empty")
	}

	return nil
}

func (a *Alert) ValidateIcon() error {
	if a.IconEmoji == "" {
		return nil
	}

	if !IconRegex.MatchString(a.IconEmoji) {
		return fmt.Errorf("iconEmoji '%s' is not valid", a.IconEmoji)
	}

	return nil
}

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

func (a *Alert) ValidateSeverity() error {
	if !SeverityIsValid(a.Severity) {
		return fmt.Errorf("severity '%s' is not valid, expected one of [%s]", a.Severity, strings.Join(ValidSeverities(), ", "))
	}

	return nil
}

func (a *Alert) ValidateCorrelationID() error {
	if a.CorrelationID == "" {
		return nil
	}

	if len(a.CorrelationID) > MaxCorrelationIDLength {
		return fmt.Errorf("correlationId is too long, expected length <=%d", MaxCorrelationIDLength)
	}

	return nil
}

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

func (a *Alert) ValidateIgnoreIfTextContains() error {
	if len(a.IgnoreIfTextContains) == 0 {
		return nil
	}

	for index, s := range a.IgnoreIfTextContains {
		if len(s) > MaxIgnoreIfTextContainsLength {
			return fmt.Errorf("ignoreIfTextContains[%d] is too long, expected length <=%d", index, MaxIgnoreIfTextContainsLength)
		}
	}

	return nil
}

func (a *Alert) ValidateFields() error {
	if len(a.Fields) > MaxFieldCount {
		return fmt.Errorf("too many fields, expected <=%d", MaxFieldCount)
	}

	return nil
}

func (a *Alert) ValidateWebhooks() error {
	if a.Webhooks == nil {
		return nil
	}

	if len(a.Webhooks) > 5 {
		return errors.New("too many webhooks, expected <=5")
	}

	webhookIDs := make(map[string]struct{})

	for index, hook := range a.Webhooks {
		if hook.ID == "" {
			return fmt.Errorf("webhook[%d].id is required", index)
		}

		if _, ok := webhookIDs[hook.ID]; ok {
			return fmt.Errorf("webhook[%d].id must be unique", index)
		}

		webhookIDs[hook.ID] = struct{}{}

		if hook.URL == "" {
			return fmt.Errorf("webhook[%d].url is required", index)
		}

		if len(hook.URL) > 1000 {
			return fmt.Errorf("webhook[%d].url is too long, expected length <=1000", index)
		}

		url, err := url.ParseRequestURI(hook.URL)
		if err != nil {
			return fmt.Errorf("webhook[%d].url is not a valid absolute URL", index)
		}

		if url.Scheme == "" {
			return fmt.Errorf("webhook[%d].url is not a valid absolute URL", index)
		}

		if hook.ButtonText == "" {
			return fmt.Errorf("webhook[%d].buttonText is required", index)
		}

		if len(hook.ButtonText) > 25 {
			return fmt.Errorf("webhook[%d].buttonText is too long, expected length <=25", index)
		}

		if len(hook.ConfirmationText) > 1000 {
			return fmt.Errorf("webhook[%d].confirmationText is too long, expected length <=1000", index)
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

		if len(hook.Payload) > 50 {
			return fmt.Errorf("webhook[%d].payload item count is too large, expected <=50", index)
		}

		if len(hook.PlainTextInput) > 10 {
			return fmt.Errorf("webhook[%d].plainTextInput item count is too large, expected <=10", index)
		}

		if len(hook.CheckboxInput) > 10 {
			return fmt.Errorf("webhook[%d].checkboxInput item count is too large, expected <=10", index)
		}

		inputIDs := make(map[string]struct{})

		for inputIndex, input := range hook.PlainTextInput {
			if input.ID == "" {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id is required", index, inputIndex)
			}

			if _, ok := inputIDs[input.ID]; ok {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id must be unique among all inputs", index, inputIndex)
			}

			inputIDs[input.ID] = struct{}{}

			if len(input.ID) > 200 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id is too long, expected <=200", index, inputIndex)
			}

			if len(input.Description) > 200 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].description is too long, expected <=200", index, inputIndex)
			}

			if input.MinLength < 0 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].minLength must be >=0", index, inputIndex)
			}

			if input.MinLength > 3000 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].minLength must be <=3000", index, inputIndex)
			}

			if input.MaxLength < 0 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength must be >=0", index, inputIndex)
			}

			if input.MaxLength > 3000 {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength must be <=3000", index, inputIndex)
			}

			if input.MaxLength < input.MinLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].maxLength cannot be smaller than minLength", index, inputIndex)
			}

			if len(input.InitialValue) > input.MaxLength {
				return fmt.Errorf("webhook[%d].plainTextInput[%d].initialValue cannot be longer than maxLength", index, inputIndex)
			}
		}

		for inputIndex, input := range hook.CheckboxInput {
			if input.ID == "" {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id is required", index, inputIndex)
			}

			if _, ok := inputIDs[input.ID]; ok {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id must be unique among all inputs", index, inputIndex)
			}

			inputIDs[input.ID] = struct{}{}

			if len(input.ID) > 200 {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id is too long, expected <=200", index, inputIndex)
			}

			if len(input.Label) > 200 {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].label is too long, expected <=200", index, inputIndex)
			}

			if len(input.Options) > 5 {
				return fmt.Errorf("webhook[%d].checkboxInput[%d].options item count is too large, expected <=5", index, inputIndex)
			}

			values := make(map[string]struct{})

			for optionIndex, option := range input.Options {
				if option.Value == "" {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value is required", index, inputIndex, optionIndex)
				}

				if _, ok := values[option.Value]; ok {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value must be unique", index, inputIndex, optionIndex)
				}

				values[option.Value] = struct{}{}

				if len(option.Text) > 50 {
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].text is too long, expected <=50", index, inputIndex, optionIndex)
				}
			}
		}
	}

	return nil
}

func (a *Alert) ValidateEscalation() error {
	if a.Escalation == nil {
		return nil
	}

	previousDelay := 0

	for index, e := range a.Escalation {
		if e.DelaySeconds < 30 {
			return fmt.Errorf("escalation[%d].delaySeconds '%d' is too low, expected value >=30", index, e.DelaySeconds)
		}

		if previousDelay > 0 && e.DelaySeconds-previousDelay < 30 {
			return fmt.Errorf("escalation[%d].delaySeconds '%d' is too small compared to previous escalation, expected diff >=30", index, e.DelaySeconds)
		}

		previousDelay = e.DelaySeconds

		if e.Severity != AlertPanic && e.Severity != AlertError && e.Severity != AlertWarning {
			return fmt.Errorf("escalation[%d].severity '%s' is not valid, expected one of [panic, error, warning]", index, e.Severity)
		}

		if len(e.SlackMentions) > 10 {
			return fmt.Errorf("escalation[%d].slackMentions item count is too large, expected <=10", index)
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
	if len(text) <= MaxTextLength {
		return text
	}

	endsWithCodeBlock := strings.HasSuffix(text, "```")

	if endsWithCodeBlock {
		return strings.TrimSpace(text[:MaxTextLength-6]) + "...```"
	}

	return strings.TrimSpace(text[:MaxTextLength-3]) + "..."
}

func hash(input ...string) string {
	h := sha256.New()

	for _, s := range input {
		h.Write([]byte(s))
	}

	bs := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(bs)
}
