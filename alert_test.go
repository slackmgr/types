package common_test

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertConstructors(t *testing.T) {
	t.Parallel()

	t.Run("panic common.alert", func(t *testing.T) {
		t.Parallel()

		a := common.NewPanicAlert()
		assert.Equal(t, common.AlertPanic, a.Severity)
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1)
	})

	t.Run("error common.alert", func(t *testing.T) {
		t.Parallel()

		a := common.NewErrorAlert()
		assert.Equal(t, common.AlertError, a.Severity)
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1)
	})

	t.Run("warning common.alert", func(t *testing.T) {
		t.Parallel()

		a := common.NewWarningAlert()
		assert.Equal(t, common.AlertWarning, a.Severity)
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1)
	})

	t.Run("resolved common.alert", func(t *testing.T) {
		t.Parallel()

		a := common.NewResolvedAlert()
		assert.Equal(t, common.AlertResolved, a.Severity)
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1)
	})

	t.Run("info common.alert", func(t *testing.T) {
		t.Parallel()

		a := common.NewInfoAlert()
		assert.Equal(t, common.AlertInfo, a.Severity)
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1)
	})
}

func TestAlertDedupID(t *testing.T) {
	t.Parallel()

	t.Run("dedup id", func(t *testing.T) {
		t.Parallel()

		timestamp := time.Now()
		a := common.Alert{
			SlackChannelID: "C12345678",
			RouteKey:       "foo",
			CorrelationID:  "bar",
			Timestamp:      timestamp,
			Header:         "header",
			Text:           "text",
		}
		expected := hash("alert", "C12345678", "foo", "bar", timestamp.UTC().Format(time.RFC3339Nano), "header", "text")
		assert.Equal(t, expected, a.UniqueID())
	})
}

func TestAlertClean(t *testing.T) {
	t.Parallel()

	var randGen *rand.Rand // nil, randString creates its own thread-safe source

	t.Run("timestamp common.newer than 7 days is kept", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		a := common.Alert{
			Timestamp: now.Add(-7 * 24 * time.Hour).Add(10 * time.Second),
		}
		a.Clean()
		assert.Equal(t, now.Add(-7*24*time.Hour).Add(10*time.Second), a.Timestamp, "timestamp should not be updated when it's less than 7 days old")
	})

	t.Run("timestamp older than 7 days is ignored", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		a := common.Alert{
			Timestamp: now.Add(-7 * 24 * time.Hour).Add(-1 * time.Second),
		}
		a.Clean()
		assert.InDelta(t, time.Now().Unix(), a.Timestamp.Unix(), 1, "timestamp should be updated to now when over 7 days old")
	})

	t.Run("type should be trimmed and lowercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Type: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "foo", a.Type)
	})

	t.Run("slackChannelID should be trimmed and uppercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			SlackChannelID: "  c12345678  ",
		}
		a.Clean()
		assert.Equal(t, "C12345678", a.SlackChannelID)
	})

	t.Run("routeKey should be trimmed and lowercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			RouteKey: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "foo", a.RouteKey)
	})

	t.Run("header should be trimmed and common.newline replaced with space", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Header:             "  Foo\nbar  ",
			HeaderWhenResolved: "  Hei\nresolved  ",
		}
		a.Clean()
		assert.Equal(t, "Foo bar", a.Header)
		assert.Equal(t, "Hei resolved", a.HeaderWhenResolved)
	})

	t.Run("text should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Text:             "  Foo\nbar  ",
			TextWhenResolved: "  Hei\nresolved  ",
		}
		a.Clean()
		assert.Equal(t, "Foo\nbar", a.Text)
		assert.Equal(t, "Hei\nresolved", a.TextWhenResolved)
	})

	t.Run("fallbackText should be trimmed and simplified", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			FallbackText: "  Foo\nbar :status: ",
		}
		a.Clean()
		assert.Equal(t, "Foo bar", a.FallbackText)
	})

	t.Run("correlationID should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			CorrelationID: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "FOO", a.CorrelationID)
	})

	t.Run("username should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Username: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "FOO", a.Username)
	})

	t.Run("author should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Author: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "FOO", a.Author)
	})

	t.Run("host should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Host: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "FOO", a.Host)
	})

	t.Run("footer should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Footer: "  FOO  ",
		}
		a.Clean()
		assert.Equal(t, "FOO", a.Footer)
	})

	t.Run("iconEmoji should be trimmed and lowercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			IconEmoji: "  :Foo:  ",
		}
		a.Clean()
		assert.Equal(t, ":foo:", a.IconEmoji)
	})

	t.Run("severity should be trimmed and lowercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Severity: "ERROR",
		}
		a.Clean()
		assert.Equal(t, common.AlertError, a.Severity)
	})

	t.Run("fallbackText should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxFallbackTextLength+1, randGen)
		a := common.Alert{
			FallbackText: s,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxFallbackTextLength-3]+"...", a.FallbackText)
	})

	t.Run("empty severity should default to error", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Severity: "",
		}
		a.Clean()
		assert.Equal(t, common.AlertError, a.Severity)
	})

	t.Run("severity 'critical' should be converted to 'error'", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Severity: "critical",
		}
		a.Clean()
		assert.Equal(t, common.AlertError, a.Severity)
	})

	t.Run("negative archivingDelaySeconds should be set to 0", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			ArchivingDelaySeconds: -1,
		}
		a.Clean()
		assert.Equal(t, 0, a.ArchivingDelaySeconds)
	})

	t.Run("negative notificationDelaySeconds should be set to 0", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			NotificationDelaySeconds: -1,
		}
		a.Clean()
		assert.Equal(t, 0, a.NotificationDelaySeconds)
	})

	t.Run("header should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxHeaderLength+1, randGen)
		s2 := randString(common.MaxHeaderLength+1, randGen)
		a := common.Alert{
			Header:             s,
			HeaderWhenResolved: s2,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxHeaderLength-3]+"...", a.Header)
		assert.Equal(t, s2[:common.MaxHeaderLength-3]+"...", a.HeaderWhenResolved)
	})

	t.Run("text should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxTextLength+1, randGen)
		s2 := randString(common.MaxTextLength+1, randGen) + "```" // Ends with code block
		a := common.Alert{
			Text:             s,
			TextWhenResolved: s2,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxTextLength-3]+"...", a.Text)
		assert.Equal(t, s2[:common.MaxTextLength-6]+"...```", a.TextWhenResolved)
	})

	t.Run("author should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxAuthorLength+1, randGen)
		a := common.Alert{
			Author: s,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxAuthorLength-3]+"...", a.Author)
	})

	t.Run("username should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxUsernameLength+1, randGen)
		a := common.Alert{
			Username: s,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxUsernameLength-3]+"...", a.Username)
	})

	t.Run("host should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxHostLength+1, randGen)
		a := common.Alert{
			Host: s,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxHostLength-3]+"...", a.Host)
	})

	t.Run("footer should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(common.MaxFooterLength+1, randGen)
		a := common.Alert{
			Footer: s,
		}
		a.Clean()
		assert.Equal(t, s[:common.MaxFooterLength-3]+"...", a.Footer)
	})

	t.Run("field titles and values should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		title := randString(common.MaxFieldTitleLength+1, randGen)
		value := randString(common.MaxFieldValueLength+1, randGen)
		a := common.Alert{
			Fields: []*common.Field{
				{Title: title, Value: value},
			},
		}
		a.Clean()
		assert.Equal(t, title[:common.MaxFieldTitleLength-3]+"...", a.Fields[0].Title)
		assert.Equal(t, value[:common.MaxFieldValueLength-3]+"...", a.Fields[0].Value)
	})

	t.Run("webhook fields be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ID:               "	foo  ",
					URL:              "  http://foo.bar  ",
					ConfirmationText: "  some text  ",
					ButtonText:       "  press me  ",
					PlainTextInput: []*common.WebhookPlainTextInput{
						{
							ID:          "  foo  ",
							Description: "  bar  ",
						},
					},
				},
			},
		}
		a.Clean()
		assert.Equal(t, "foo", a.Webhooks[0].ID)
		assert.Equal(t, "http://foo.bar", a.Webhooks[0].URL)
		assert.Equal(t, "some text", a.Webhooks[0].ConfirmationText)
		assert.Equal(t, "press me", a.Webhooks[0].ButtonText)
		assert.Equal(t, "foo", a.Webhooks[0].PlainTextInput[0].ID)
		assert.Equal(t, "bar", a.Webhooks[0].PlainTextInput[0].Description)
	})

	t.Run("webhook button style 'default' should be replaced with empty string", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ButtonStyle: "default",
				},
			},
		}
		a.Clean()
		assert.Equal(t, common.WebhookButtonStyle(""), a.Webhooks[0].ButtonStyle)
	})

	t.Run("alert escalation should be sorted by delay seconds", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Escalation: []*common.Escalation{
				{DelaySeconds: 60},
				{DelaySeconds: 30},
			},
		}
		a.Clean()
		assert.Equal(t, 30, a.Escalation[0].DelaySeconds)
		assert.Equal(t, 60, a.Escalation[1].DelaySeconds)
	})

	t.Run("alert escalation moveToChannel should be trimmed and uppercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Escalation: []*common.Escalation{
				{MoveToChannel: "  c12345678  "},
			},
		}
		a.Clean()
		assert.Equal(t, "C12345678", a.Escalation[0].MoveToChannel)
	})

	t.Run("alert escalation mentions should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Escalation: []*common.Escalation{
				{SlackMentions: []string{"  <@foo>  "}},
			},
		}
		a.Clean()
		assert.Equal(t, "<@foo>", a.Escalation[0].SlackMentions[0])
	})
}

func TestAlertValidation(t *testing.T) {
	t.Parallel()

	var randGen *rand.Rand // nil, randString creates its own thread-safe source

	t.Run("valid minimum common.alert", func(t *testing.T) {
		t.Parallel()

		var a *common.Alert
		require.Error(t, a.Validate())
		a = &common.Alert{
			SlackChannelID: "C12345678",
			Header:         "foo",
		}
		a.Clean()
		require.NoError(t, a.Validate())
	})

	t.Run("alert.slackChannelID and common.alert.routeKey should be on the correct format", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{SlackChannelID: "abcdefghi", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		a = &common.Alert{SlackChannelID: "ABab129cf", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		a = &common.Alert{SlackChannelID: "abcdefghi9238yr", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Channel ID and route key can both be empty
		a = &common.Alert{Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Channel names are allowed
		a = &common.Alert{SlackChannelID: "12345678", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Channel names are allowed
		a = &common.Alert{SlackChannelID: "foo-something", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Invalid characters
		a = &common.Alert{SlackChannelID: "sdkjsdf asdfasdf", Header: "foo"}
		a.Clean()
		require.Error(t, a.Validate())

		// Too long channelID
		a = &common.Alert{SlackChannelID: randString(common.MaxSlackChannelIDLength+1, randGen), Header: "foo"}
		a.Clean()
		require.Error(t, a.Validate())

		// routeKey is OK
		a = &common.Alert{RouteKey: "abcdefghi", Header: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// routeKey too long
		a = &common.Alert{RouteKey: randString(common.MaxRouteKeyLength+1, randGen), Header: "foo"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "routeKey")
	})

	t.Run("alert.header and common.alert.text cannot both be empty", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{
			SlackChannelID: "C12345678",
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "header and text")
	})

	t.Run("alert.iconEmoji should be on the correct format", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", IconEmoji: ":foo:"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Invalid format
		a = &common.Alert{Header: "a", RouteKey: "b", IconEmoji: ":foo:bar:"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "iconEmoji")

		// Invalid format
		a = &common.Alert{Header: "a", RouteKey: "b", IconEmoji: "foo"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "iconEmoji")

		// Invalid format
		a = &common.Alert{Header: "a", RouteKey: "b", IconEmoji: "foo:"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "iconEmoji")

		// Too long
		a = &common.Alert{Header: "a", RouteKey: "b", IconEmoji: ":" + randString(common.MaxIconEmojiLength+1, randGen) + ":"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "iconEmoji")
	})

	t.Run("alert.link should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Invalid format
		a = &common.Alert{Header: "a", RouteKey: "b", Link: "foo"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "link is not a valid absolute URL")

		// Valid format
		a = &common.Alert{Header: "a", RouteKey: "b", Link: "http://foo.bar?foo=bar#sfd"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Relative url is not allowed
		a = &common.Alert{Header: "a", RouteKey: "b", Link: "/foo"}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "link is not a valid absolute URL")
	})

	t.Run("alert.severity should be on the correct format", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError}
		a.Clean()
		require.NoError(t, a.Validate())

		a = &common.Alert{Header: "a", RouteKey: "b", Severity: "foo"} // Invalid severity
		a.Clean()
		require.ErrorContains(t, a.Validate(), "severity")
	})

	t.Run("alert.correlationID should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Valid format
		a = &common.Alert{Header: "a", RouteKey: "b", CorrelationID: "foo"}
		a.Clean()
		require.NoError(t, a.Validate())

		// Too long
		a = &common.Alert{Header: "a", RouteKey: "b", CorrelationID: randString(common.MaxCorrelationIDLength+1, randGen)}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "correlationId")
	})

	t.Run("alert.autoResolveSeconds should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Min value is OK
		a := &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: true, AutoResolveSeconds: common.MinAutoResolveSeconds}
		a.Clean()
		require.NoError(t, a.Validate())

		// Too small
		a = &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: true, AutoResolveSeconds: common.MinAutoResolveSeconds - 1}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "autoResolveSeconds")

		// Negative value
		a = &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: true, AutoResolveSeconds: -1}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "autoResolveSeconds")

		// Ignore invalid value when issueFollowUpEnabled is false
		a = &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: false, AutoResolveSeconds: -1}
		a.Clean()
		require.NoError(t, a.Validate())

		// Too long
		a = &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: true, AutoResolveSeconds: common.MaxAutoResolveSeconds + 1}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "autoResolveSeconds")

		// Maximum value is OK
		a = &common.Alert{Header: "a", RouteKey: "b", IssueFollowUpEnabled: true, AutoResolveSeconds: common.MaxAutoResolveSeconds}
		a.Clean()
		require.NoError(t, a.Validate())
	})

	t.Run("alert.ignoreIfTextContains should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b", IgnoreIfTextContains: []string{}}
		a.Clean()
		require.NoError(t, a.Validate())

		// All good
		a = &common.Alert{Header: "a", RouteKey: "b", IgnoreIfTextContains: []string{"foo", "bar"}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Max length is OK
		a = &common.Alert{Header: "a", RouteKey: "b", IgnoreIfTextContains: []string{randString(common.MaxIgnoreIfTextContainsLength, randGen)}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Too long
		a = &common.Alert{Header: "a", RouteKey: "b", IgnoreIfTextContains: []string{"foo", randString(common.MaxIgnoreIfTextContainsLength+1, randGen)}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "ignoreIfTextContains")
	})

	t.Run("alert.Fields should not have too many items", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b"}
		for i := 1; i <= common.MaxFieldCount+1; i++ {
			a.Fields = append(a.Fields, &common.Field{Title: "foo", Value: "bar"})
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "too many fields")
	})

	t.Run("alert.webhooks should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Max webhooks is MaxWebhookCount
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{}}
		for i := 1; i <= common.MaxWebhookCount+1; i++ {
			a.Webhooks = append(a.Webhooks, &common.Webhook{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"})
		}
		a.Clean()
		require.Error(t, a.Validate(), "too many webhooks")

		// ID is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "", URL: "http://foo.bar", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].id is required")

		// ID must be unique
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{
			{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"},
			{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"},
		}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[1].id must be unique")

		// Url is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is required")

		// Url max length is MaxWebhookURLLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar/" + randString(common.MaxWebhookURLLength-14, randGen), ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is too long")

		// HTTP Url must be valid absolute URL
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is not a valid absolute URL")

		// HTTP Url must be absolute
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "https://", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is not a valid absolute URL")

		// Non-HTTP URL (custom handler) must be valid ASCII
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "custom-handler\x00invalid", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url contains invalid characters")

		// Non-HTTP URL (custom handler) with valid ASCII is accepted
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "my-custom-handler:action", ButtonText: "press me"}}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Button text is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: ""}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].buttonText is required")

		// Button text max length is MaxWebhookButtonTextLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: randString(common.MaxWebhookButtonTextLength+1, randGen)}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].buttonText is too long")

		// Confirmation text max length is MaxWebhookConfirmationTextLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", ConfirmationText: randString(common.MaxWebhookConfirmationTextLength+1, randGen)}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].confirmationText is too long")

		// Button style must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", ButtonStyle: "foo"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].buttonStyle 'foo' is not valid")

		// Access level must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", AccessLevel: "foo"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].accessLevel 'foo' is not valid")

		// Display mode must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", DisplayMode: "foo"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].displayMode 'foo' is not valid")

		// Max payload size is MaxWebhookPayloadCount
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", Payload: map[string]any{}}}}
		for i := 1; i <= common.MaxWebhookPayloadCount+1; i++ {
			a.Webhooks[0].Payload[randString(10, randGen)] = randString(10, randGen)
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].payload item count is too large")

		// Max plain text input size is MaxWebhookPlainTextInputCount
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"}}}
		for i := 1; i <= common.MaxWebhookPlainTextInputCount+1; i++ {
			a.Webhooks[0].PlainTextInput = append(a.Webhooks[0].PlainTextInput, &common.WebhookPlainTextInput{ID: randString(5, randGen), Description: randString(5, randGen)})
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput item count is too large")

		// Max checkbox size is MaxWebhookCheckboxInputCount
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"}}}
		for i := 1; i <= common.MaxWebhookCheckboxInputCount+1; i++ {
			a.Webhooks[0].CheckboxInput = append(a.Webhooks[0].CheckboxInput, &common.WebhookCheckboxInput{ID: randString(5, randGen), Label: randString(5, randGen)})
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput item count is too large")

		// Input ID is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "", Description: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].id is required")

		// Input ID must be unique
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo"}, {ID: "foo", Description: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[1].id must be unique")

		// Input ID max length is MaxWebhookInputIDLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: randString(common.MaxWebhookInputIDLength+1, randGen), Description: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].id is too long")

		// Input description max length is MaxWebhookInputDescriptionLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: randString(common.MaxWebhookInputDescriptionLength+1, randGen)}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].description is too long")

		// Input minLength cannot be negative
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MinLength: -1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].minLength must be >=0")

		// Input minLength cannot be larger than MaxWebhookInputTextLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MinLength: common.MaxWebhookInputTextLength + 1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), fmt.Sprintf("webhook[0].plainTextInput[0].minLength must be <=%d", common.MaxWebhookInputTextLength))

		// Input maxLength cannot be negative
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MaxLength: -1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].maxLength must be >=0")

		// Input maxLength cannot be larger than MaxWebhookInputTextLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MaxLength: common.MaxWebhookInputTextLength + 1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), fmt.Sprintf("webhook[0].plainTextInput[0].maxLength must be <=%d", common.MaxWebhookInputTextLength))

		// Input minLength cannot be larger than maxLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MinLength: 10, MaxLength: 5}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].maxLength cannot be smaller than minLength")

		// Input initialValue cannot be longer than maxLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MaxLength: 5, InitialValue: "123456"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].initialValue cannot be longer than maxLength")

		// Checkbox ID is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "", Label: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].id is required")

		// Checkbox ID must be unique
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo"}, {ID: "foo", Label: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[1].id must be unique")

		// Checkbox ID max length is MaxWebhookInputIDLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: randString(common.MaxWebhookInputIDLength+1, randGen), Label: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].id is too long")

		// Checkbox label max length is MaxWebhookInputLabelLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: randString(common.MaxWebhookInputLabelLength+1, randGen)}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].label is too long")

		// Checkbox options length cannot be larger than MaxWebhookCheckboxOptionCount
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo"}}}}}
		for i := 1; i <= common.MaxWebhookCheckboxOptionCount+1; i++ {
			a.Webhooks[0].CheckboxInput[0].Options = append(a.Webhooks[0].CheckboxInput[0].Options, &common.WebhookCheckboxOption{Value: "foo"})
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].options item count is too large")

		// Checkbox option value is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo", Options: []*common.WebhookCheckboxOption{{Value: ""}}}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].options[0].value is required")

		// Checkbox option value must be unique
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo", Options: []*common.WebhookCheckboxOption{{Value: "foo"}, {Value: "foo"}}}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].options[1].value must be unique")

		// Checkbox option text max length is MaxWebhookCheckboxOptionTextLength
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo", Options: []*common.WebhookCheckboxOption{{Value: "foo", Text: randString(common.MaxWebhookCheckboxOptionTextLength+1, randGen)}}}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].options[0].text is too long")
	})

	t.Run("alert.escalation should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Escalation delay must be at least MinEscalationDelaySeconds
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds - 1, Severity: common.AlertError}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), fmt.Sprintf("escalation[0].delaySeconds '%d' is too low", common.MinEscalationDelaySeconds-1))

		// Escalation delay must be at least MinEscalationDelayDiffSeconds larger than the previous escalation
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError}, {DelaySeconds: common.MinEscalationDelaySeconds + common.MinEscalationDelayDiffSeconds - 1, Severity: common.AlertPanic}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), fmt.Sprintf("escalation[1].delaySeconds '%d' is too small compared to previous escalation", common.MinEscalationDelaySeconds+common.MinEscalationDelayDiffSeconds-1))

		// Escalation severity must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: "foo"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'foo' is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertInfo}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'info' is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertResolved}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'resolved' is not valid")

		// Escalation mentions count must be at most MaxEscalationSlackMentionCount
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError, SlackMentions: []string{}}}}
		for i := 1; i <= common.MaxEscalationSlackMentionCount+1; i++ {
			a.Escalation[0].SlackMentions = append(a.Escalation[0].SlackMentions, "<@foo>")
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions item count is too large")

		// Escalation mentions must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError, SlackMentions: []string{"foo"}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions[0] is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError, SlackMentions: []string{"<@" + randString(common.MaxMentionLength+1, randGen) + ">"}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions[0] is not valid")

		// Escalation moveToChannel must be a valid channel ID or channel name
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError, MoveToChannel: "foo bar"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].moveToChannel is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: common.MinEscalationDelaySeconds, Severity: common.AlertError, MoveToChannel: randString(common.MaxSlackChannelIDLength+1, randGen)}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].moveToChannel is not valid")
	})
}

func TestAlertCleanUnicodeTruncation(t *testing.T) {
	t.Parallel()

	// Use a multi-byte UTF-8 character (Japanese "sun" character, 3 bytes per rune)
	multiByteChar := "\u65e5" // 日

	t.Run("header with unicode should truncate by rune count not bytes", func(t *testing.T) {
		t.Parallel()

		// Create a string of 131 Japanese characters (393 bytes)
		header := strings.Repeat(multiByteChar, 131)
		a := common.Alert{Header: header}
		a.Clean()

		// Should truncate to 127 runes + "..." = 130 runes total
		assert.Len(t, []rune(a.Header), 130)
		assert.True(t, strings.HasSuffix(a.Header, "..."))
	})

	t.Run("text with unicode should truncate safely", func(t *testing.T) {
		t.Parallel()

		// Create text longer than MaxTextLength with multi-byte characters (10002 runes)
		text := strings.Repeat(multiByteChar, 10001)
		a := common.Alert{Text: text}
		a.Clean()

		// Should be exactly MaxTextLength runes
		assert.Len(t, []rune(a.Text), common.MaxTextLength)
		assert.True(t, strings.HasSuffix(a.Text, "..."))
	})

	t.Run("fallbackText with unicode should truncate safely", func(t *testing.T) {
		t.Parallel()

		fallback := strings.Repeat(multiByteChar, 151)
		a := common.Alert{FallbackText: fallback}
		a.Clean()

		assert.Len(t, []rune(a.FallbackText), common.MaxFallbackTextLength)
		assert.True(t, strings.HasSuffix(a.FallbackText, "..."))
	})

	t.Run("field title with unicode should truncate safely", func(t *testing.T) {
		t.Parallel()

		title := strings.Repeat(multiByteChar, 31)
		a := common.Alert{
			Fields: []*common.Field{{Title: title, Value: "test"}},
		}
		a.Clean()

		assert.Len(t, []rune(a.Fields[0].Title), common.MaxFieldTitleLength)
		assert.True(t, strings.HasSuffix(a.Fields[0].Title, "..."))
	})

	t.Run("field value with unicode should truncate safely", func(t *testing.T) {
		t.Parallel()

		value := strings.Repeat(multiByteChar, 201)
		a := common.Alert{
			Fields: []*common.Field{{Title: "test", Value: value}},
		}
		a.Clean()

		assert.Len(t, []rune(a.Fields[0].Value), common.MaxFieldValueLength)
		assert.True(t, strings.HasSuffix(a.Fields[0].Value, "..."))
	})
}

func TestAlertCleanNilElements(t *testing.T) {
	t.Parallel()

	t.Run("nil field element should not panic", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Fields: []*common.Field{nil, {Title: "test", Value: "value"}, nil},
		}
		assert.NotPanics(t, func() { a.Clean() })
		assert.Equal(t, "test", a.Fields[1].Title)
	})

	t.Run("nil webhook element should not panic", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{nil, {ID: "test", URL: "http://test.com", ButtonText: "click"}},
		}
		assert.NotPanics(t, func() { a.Clean() })
		assert.Equal(t, "test", a.Webhooks[1].ID)
	})

	t.Run("nil plainTextInput element should not panic", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ID:         "test",
					URL:        "http://test.com",
					ButtonText: "click",
					PlainTextInput: []*common.WebhookPlainTextInput{
						nil,
						{ID: "input1", Description: "desc"},
					},
				},
			},
		}
		assert.NotPanics(t, func() { a.Clean() })
		assert.Equal(t, "input1", a.Webhooks[0].PlainTextInput[1].ID)
	})

	t.Run("nil checkboxInput element should not panic", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ID:         "test",
					URL:        "http://test.com",
					ButtonText: "click",
					CheckboxInput: []*common.WebhookCheckboxInput{
						nil,
						{ID: "cb1", Label: "My Label"},
					},
				},
			},
		}
		assert.NotPanics(t, func() { a.Clean() })
		assert.Equal(t, "cb1", a.Webhooks[0].CheckboxInput[1].ID)
	})

	t.Run("nil escalation element should not panic", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Escalation: []*common.Escalation{
				nil,
				{DelaySeconds: 60, Severity: common.AlertError, MoveToChannel: "  c12345  "},
				nil,
			},
		}
		assert.NotPanics(t, func() { a.Clean() })
		assert.Equal(t, "C12345", a.Escalation[2].MoveToChannel) // nil elements sorted to front
	})
}

func TestAlertCleanAdditional(t *testing.T) {
	t.Parallel()

	t.Run("checkboxInput ID and Label should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ID:         "test",
					URL:        "http://test.com",
					ButtonText: "click",
					CheckboxInput: []*common.WebhookCheckboxInput{
						{ID: "  checkbox1  ", Label: "  My Label  "},
					},
				},
			},
		}
		a.Clean()
		assert.Equal(t, "checkbox1", a.Webhooks[0].CheckboxInput[0].ID)
		assert.Equal(t, "My Label", a.Webhooks[0].CheckboxInput[0].Label)
	})

	t.Run("plainTextInput InitialValue should be trimmed", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Webhooks: []*common.Webhook{
				{
					ID:         "test",
					URL:        "http://test.com",
					ButtonText: "click",
					PlainTextInput: []*common.WebhookPlainTextInput{
						{ID: "input1", Description: "desc", InitialValue: "  initial  "},
					},
				},
			},
		}
		a.Clean()
		assert.Equal(t, "initial", a.Webhooks[0].PlainTextInput[0].InitialValue)
	})

	t.Run("escalation severity should be trimmed and lowercased", func(t *testing.T) {
		t.Parallel()

		a := common.Alert{
			Escalation: []*common.Escalation{
				{DelaySeconds: 60, Severity: "  ERROR  "},
			},
		}
		a.Clean()
		assert.Equal(t, common.AlertError, a.Escalation[0].Severity)
	})
}

func TestAlertValidationAdditional(t *testing.T) {
	t.Parallel()

	var randGen *rand.Rand

	t.Run("escalation count should not exceed 3", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{
			{DelaySeconds: 30, Severity: common.AlertError},
			{DelaySeconds: 60, Severity: common.AlertError},
			{DelaySeconds: 90, Severity: common.AlertError},
		}}
		a.Clean()
		require.NoError(t, a.Validate())

		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{
			{DelaySeconds: 30, Severity: common.AlertError},
			{DelaySeconds: 60, Severity: common.AlertError},
			{DelaySeconds: 90, Severity: common.AlertError},
			{DelaySeconds: 120, Severity: common.AlertError},
		}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "too many escalation points")
	})

	t.Run("initialValue should not be shorter than minLength", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{
			ID:         "foo",
			URL:        "http://foo.bar",
			ButtonText: "press me",
			PlainTextInput: []*common.WebhookPlainTextInput{
				{ID: "input1", Description: "desc", MinLength: 5, MaxLength: 100, InitialValue: "ab"},
			},
		}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "initialValue cannot be shorter than minLength")
	})

	t.Run("ignoreIfTextContains count should not exceed max", func(t *testing.T) {
		t.Parallel()

		items := make([]string, common.MaxIgnoreIfTextContainsCount+1)
		for i := range items {
			items[i] = "item"
		}
		a := &common.Alert{Header: "a", RouteKey: "b", IgnoreIfTextContains: items}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "too many ignoreIfTextContains items")
	})

	t.Run("webhook ID should not exceed max length", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{
			ID:         randString(common.MaxWebhookIDLength+1, randGen),
			URL:        "http://foo.bar",
			ButtonText: "press me",
		}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].id is too long")
	})

	t.Run("checkbox option value should not exceed max length", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{
			ID:         "foo",
			URL:        "http://foo.bar",
			ButtonText: "press me",
			CheckboxInput: []*common.WebhookCheckboxInput{{
				ID:    "cb1",
				Label: "label",
				Options: []*common.WebhookCheckboxOption{
					{Value: randString(common.MaxCheckboxOptionValueLength+1, randGen), Text: "text"},
				},
			}},
		}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "options[0].value is too long")
	})

	t.Run("nil webhook should return error in validation", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError, Webhooks: []*common.Webhook{nil}}
		require.ErrorContains(t, a.Validate(), "webhook[0] is nil")
	})

	t.Run("nil plainTextInput should return error in validation", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError, Webhooks: []*common.Webhook{{
			ID:             "foo",
			URL:            "http://foo.bar",
			ButtonText:     "press me",
			PlainTextInput: []*common.WebhookPlainTextInput{nil},
		}}}
		require.ErrorContains(t, a.Validate(), "plainTextInput[0] is nil")
	})

	t.Run("nil checkboxInput should return error in validation", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError, Webhooks: []*common.Webhook{{
			ID:            "foo",
			URL:           "http://foo.bar",
			ButtonText:    "press me",
			CheckboxInput: []*common.WebhookCheckboxInput{nil},
		}}}
		require.ErrorContains(t, a.Validate(), "checkboxInput[0] is nil")
	})

	t.Run("nil checkbox option should return error in validation", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError, Webhooks: []*common.Webhook{{
			ID:         "foo",
			URL:        "http://foo.bar",
			ButtonText: "press me",
			CheckboxInput: []*common.WebhookCheckboxInput{{
				ID:      "cb1",
				Label:   "label",
				Options: []*common.WebhookCheckboxOption{nil},
			}},
		}}}
		require.ErrorContains(t, a.Validate(), "options[0] is nil")
	})

	t.Run("nil escalation should return error in validation", func(t *testing.T) {
		t.Parallel()

		a := &common.Alert{Header: "a", RouteKey: "b", Severity: common.AlertError, Escalation: []*common.Escalation{nil}}
		require.ErrorContains(t, a.Validate(), "escalation[0] is nil")
	})
}

var testLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randString generates a random string of n characters.
// Each call creates its own random source to be safe for concurrent use in parallel tests.
func randString(n int, _ *rand.Rand) string {
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = testLetters[localRand.Intn(len(testLetters))]
	}
	return string(b)
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
