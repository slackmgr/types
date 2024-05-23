package common

import (
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
	SlackChannelIDOrNameRegex = regexp.MustCompile(`^[0-9a-zA-Z\-_]{1,80}$`)

	// IconRegex matches valid Slack icon emojis, on the format ':emoji:'.
	IconRegex = regexp.MustCompile(`^:[^:]{1,50}:$`)

	// SlackMentionRegex matches valid Slack mentions, such as <!here>, <!channel> and <@U12345678>.
	SlackMentionRegex = regexp.MustCompile(`^((<!here>)|(<!channel>)|(<@[^>\s]{1,100}>))$`)

	// MaxTimestampAge is the maximum age of an alert timestamp. If the timestamp is older than this, it will be replaced with the current time.
	MaxTimestampAge = 7 * 24 * time.Hour

	MaxHeaderLength       = 130
	MaxFallbackTextLength = 150
	MaxTextLength         = 10000
	MaxAuthorLength       = 100
	MaxHostLength         = 100
	MaxFooterLength       = 300
)

// Alert represents a single alert that can be sent to the Slack Manager.
type Alert struct {
	// Timestamp is the time when the alert was created. If the timestamp is empty (or older than 7 days), it will be replaced with the current time.
	Timestamp time.Time `json:"timestamp"`

	// CorrelationID is an optional field used to group related alerts together.
	// If unset, the correlation ID is constructed by hashing [Header, Text, Author, Host, SlackChannelID].
	// It is strongly recommended to set this to an explicit value, which makes sense in your context, rather than relying on the default hash value.
	CorrelationID string `json:"correlationId"`

	// Header is the main header (title) of the alert.
	// It is automatically truncated to 130 characters.
	// This field is optional, but Header and Text cannot both be empty.
	Header string `json:"header"`

	// HeaderWhenResolved is the main header (title) of the alert when in the *resolved* state.
	// It is automatically truncated to 130 characters.
	// This field is optional. If unset, the Header field is used for all alert states.
	HeaderWhenResolved string `json:"headerWhenResolved"`

	// Text is the main text (body) of the alert.
	// It is automatically truncated to 10000 characters.
	// This field is optional, but Header and Text cannot both be empty.
	Text string `json:"text"`

	// TextWhenResolved is the main text (body) of the alert when in the *resolved* state.
	// It is automatically truncated to 10000 characters.
	// This field is optional. If unset, the Text field is used for all alert states.
	TextWhenResolved string `json:"textWhenResolved"`

	// FallbackText is the text displayed in Slack notifications. It should be a short, human-readable summary of the alert, without markdown or line breaks.
	// It is automatically truncated to 150 characters.
	// This field is optional. If unset, Slack decides what to display in notifications (which may not always be ideal).
	FallbackText string `json:"fallbackText"`

	// Author is the 'author' of the alert (if relevant), displayed as a context block in the Slack post.
	// It is automatically truncated to 100 characters.
	// This field is optional.
	Author string `json:"author"`

	// Host is the 'host' on which the alert originated (if any), displayed as a context block in the Slack post.
	// It is automatically truncated to 100 characters.
	// This field is optional.
	Host                      string        `json:"host"`
	Footer                    string        `json:"footer"`
	Link                      string        `json:"link"`
	AutoResolveSeconds        int           `json:"autoResolveSeconds"`
	AutoResolveAsInconclusive bool          `json:"autoResolveAsInconclusive"`
	Severity                  AlertSeverity `json:"severity"`
	SlackChannelID            string        `json:"slackChannelId"`
	RouteKey                  string        `json:"routeKey"`
	IssueFollowUpEnabled      bool          `json:"issueFollowUpEnabled"`
	Username                  string        `json:"username"`
	IconEmoji                 string        `json:"iconEmoji"`

	// Fields are rendered in a compact format that allows for 2 columns of side-by-side text.
	Fields                   []*Field               `json:"fields"`
	NotificationDelaySeconds int                    `json:"notificationDelaySeconds"`
	ArchivingDelaySeconds    int                    `json:"archivingDelaySeconds"`
	Escalation               []*Escalation          `json:"escalation"`
	IgnoreIfTextContains     []string               `json:"ignoreIfTextContains"`
	FailOnRateLimitError     bool                   `json:"failOnRateLimitError"`
	Webhooks                 []*Webhook             `json:"webhooks"`
	Metadata                 map[string]interface{} `json:"metadata"`
}

// Field is an alert field.
type Field struct {
	// Title is the title of the field. It is automatically truncated to 30 characters.
	Title string `json:"title"`

	// Value is the value of the field. It is automatically truncated to 200 characters.
	Value string `json:"value"`
}

// Escalation represents an escalation point for an alert.
type Escalation struct {
	Severity      AlertSeverity `json:"severity"`
	DelaySeconds  int           `json:"delaySeconds"`
	SlackMentions []string      `json:"slackMentions"`
	MoveToChannel string        `json:"moveToChannel"`
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
		Metadata:  make(map[string]interface{}),
	}
}

func (a *Alert) Clean() {
	if time.Since(a.Timestamp) > MaxTimestampAge {
		a.Timestamp = time.Now()
	}

	a.SlackChannelID = strings.TrimSpace(a.SlackChannelID)
	a.RouteKey = strings.TrimSpace(a.RouteKey)
	a.Header = strings.ReplaceAll(strings.TrimSpace(a.Header), "\n", " ")
	a.HeaderWhenResolved = strings.ReplaceAll(strings.TrimSpace(a.HeaderWhenResolved), "\n", " ")
	a.Text = strings.TrimSpace(a.Text)
	a.TextWhenResolved = strings.TrimSpace(a.TextWhenResolved)
	a.FallbackText = strings.ReplaceAll(strings.TrimSpace(a.FallbackText), ":status:", "")
	a.FallbackText = strings.ReplaceAll(a.FallbackText, "\n", " ")
	a.CorrelationID = strings.TrimSpace(a.CorrelationID)
	a.Username = strings.TrimSpace(a.Username)
	a.Author = strings.TrimSpace(a.Author)
	a.Host = strings.TrimSpace(a.Host)
	a.Footer = strings.TrimSpace(a.Footer)
	a.IconEmoji = strings.TrimSpace(a.IconEmoji)
	a.Severity = AlertSeverity(strings.ToLower(strings.TrimSpace(string(a.Severity))))

	if len(a.FallbackText) > MaxFallbackTextLength {
		a.FallbackText = a.FallbackText[:MaxFallbackTextLength-3] + "..."
	}

	if a.Severity == "critical" {
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

	if len(a.Username) > 100 {
		a.Username = strings.TrimSpace(a.Username[:97]) + "..."
	}

	if len(a.Footer) > MaxFooterLength {
		a.Footer = strings.TrimSpace(a.Footer[:MaxFooterLength-3]) + "..."
	}

	for _, field := range a.Fields {
		if len(field.Title) > 30 {
			field.Title = strings.TrimSpace(field.Title[:27]) + "..."
		}

		if len(field.Value) > 200 {
			field.Value = strings.TrimSpace(field.Value[:197]) + "..."
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
			e.MoveToChannel = strings.TrimSpace(e.MoveToChannel)

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

	if err := a.ValidateSlackChannelID(); err != nil {
		return err
	}

	if err := a.ValidateHeaderAndText(); err != nil {
		return err
	}

	if err := a.ValidateIcon(); err != nil {
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

func (a *Alert) ValidateSlackChannelID() error {
	if a.SlackChannelID != "" {
		if !SlackChannelIDOrNameRegex.MatchString(a.SlackChannelID) {
			return fmt.Errorf("slackChannelId '%s' is invalid", a.SlackChannelID)
		}

		return nil
	}

	if a.RouteKey == "" {
		return errors.New("slackChannelId and routeKey cannot both be empty")
	}

	if len(a.RouteKey) > 1000 {
		return errors.New("routeKey is too long, expected length <=1000")
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
		return fmt.Errorf("iconEmoji '%s' is invalid", a.IconEmoji)
	}

	return nil
}

func (a *Alert) ValidateSeverity() error {
	if a.Severity == "" {
		return errors.New("severity cannot be empty")
	}

	if !SeverityIsValid(a.Severity) {
		return fmt.Errorf("severity '%s' is invalid, expected one of [%s]", a.Severity, strings.Join(ValidSeverities(), ", "))
	}

	return nil
}

func (a *Alert) ValidateCorrelationID() error {
	if a.CorrelationID == "" {
		return nil
	}

	if len(a.CorrelationID) > 500 {
		return errors.New("correlationId is too long, expected length <=500")
	}

	return nil
}

func (a *Alert) ValidateAutoResolve() error {
	if !a.IssueFollowUpEnabled {
		return nil
	}

	if a.AutoResolveSeconds < 30 {
		return fmt.Errorf("autoResolveSeconds %d is too low, expected value >=30", a.AutoResolveSeconds)
	}

	// 2 years
	if a.AutoResolveSeconds > 63113851 {
		return fmt.Errorf("autoResolveSeconds %d is too high, expected value <=63113851 (2 years)", a.AutoResolveSeconds)
	}

	return nil
}

func (a *Alert) ValidateIgnoreIfTextContains() error {
	if len(a.IgnoreIfTextContains) == 0 {
		return nil
	}

	for index, s := range a.IgnoreIfTextContains {
		if len(s) > 1000 {
			return fmt.Errorf("ignoreIfTextContains[%d] is too long, expected length <=1000", index)
		}
	}

	return nil
}

func (a *Alert) ValidateFields() error {
	if len(a.Fields) > 20 {
		return errors.New("too many fields, expected <=20")
	}

	return nil
}

func (a *Alert) ValidateWebhooks() error {
	if a.Webhooks == nil {
		return nil
	}

	if len(a.Webhooks) > 5 {
		return fmt.Errorf("too many webhooks, expected <=5")
	}

	webhookIDs := make(map[string]struct{})

	for index, hook := range a.Webhooks {
		if hook.ID == "" {
			return fmt.Errorf("webhook[%d].id is empty", index)
		}

		if _, ok := webhookIDs[hook.ID]; ok {
			return fmt.Errorf("webhook[%d].id must be unique", index)
		}

		webhookIDs[hook.ID] = struct{}{}

		if hook.URL == "" {
			return fmt.Errorf("webhook[%d].url is empty", index)
		}

		if len(hook.URL) > 1000 {
			return fmt.Errorf("webhook[%d].url is too long, expected length <=1000", index)
		}

		_, err := url.ParseRequestURI(hook.URL)
		if err != nil {
			return fmt.Errorf("webhook[%d].url '%s' is not a valid absolute URL", index, hook.URL)
		}

		if hook.ButtonText == "" {
			return fmt.Errorf("webhook[%d].buttonText is empty", index)
		}

		if len(hook.ButtonText) > 25 {
			return fmt.Errorf("webhook[%d].buttonText is too long, expected length <=25", index)
		}

		if len(hook.ConfirmationText) > 1000 {
			return fmt.Errorf("webhook[%d].displayText is too long, expected length <=1000", index)
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
				return fmt.Errorf("webhook[%d].plainTextInput[%d].id is empty", index, inputIndex)
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
				return fmt.Errorf("webhook[%d].checkboxInput[%d].id is empty", index, inputIndex)
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
					return fmt.Errorf("webhook[%d].checkboxInput[%d].options[%d].value is empty", index, inputIndex, optionIndex)
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
			return fmt.Errorf("escalation[%d].severity '%s' is invalid, expected one of [panic, error, warning]", index, e.Severity)
		}

		if len(e.SlackMentions) > 10 {
			return fmt.Errorf("escalation[%d].slackMentions have too many mentions, expected <=10", index)
		}

		for j, mention := range e.SlackMentions {
			if !SlackMentionRegex.MatchString(mention) {
				return fmt.Errorf("escalation[%d].slackMentions[%d] '%s' is not valid", index, j, mention)
			}
		}

		if e.MoveToChannel != "" {
			if !SlackChannelIDOrNameRegex.MatchString(e.MoveToChannel) {
				return fmt.Errorf("escalation[%d].moveToChannel '%s' is invalid", index, e.MoveToChannel)
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
