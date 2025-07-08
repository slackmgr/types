package common_test

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"testing"
	"time"

	common "github.com/peteraglen/slack-manager-common"
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
		expected := hash("alert", "C12345678", "foo", "bar", timestamp.Format(time.RFC3339Nano), "header", "text")
		assert.Equal(t, expected, a.UniqueID())
	})
}

func TestAlertClean(t *testing.T) {
	t.Parallel()

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

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

		s := randString(151, randGen)
		a := common.Alert{
			FallbackText: s,
		}
		a.Clean()
		assert.Equal(t, s[:147]+"...", a.FallbackText)
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

		s := randString(131, randGen)
		s2 := randString(131, randGen)
		a := common.Alert{
			Header:             s,
			HeaderWhenResolved: s2,
		}
		a.Clean()
		assert.Equal(t, s[:127]+"...", a.Header)
		assert.Equal(t, s2[:127]+"...", a.HeaderWhenResolved)
	})

	t.Run("text should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(10001, randGen)
		s2 := randString(10001, randGen) + "```" // Ends with code block
		a := common.Alert{
			Text:             s,
			TextWhenResolved: s2,
		}
		a.Clean()
		assert.Equal(t, s[:9997]+"...", a.Text)
		assert.Equal(t, s2[:9994]+"...```", a.TextWhenResolved)
	})

	t.Run("author should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(101, randGen)
		a := common.Alert{
			Author: s,
		}
		a.Clean()
		assert.Equal(t, s[:97]+"...", a.Author)
	})

	t.Run("username should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(101, randGen)
		a := common.Alert{
			Username: s,
		}
		a.Clean()
		assert.Equal(t, s[:97]+"...", a.Username)
	})

	t.Run("host should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(101, randGen)
		a := common.Alert{
			Host: s,
		}
		a.Clean()
		assert.Equal(t, s[:97]+"...", a.Host)
	})

	t.Run("footer should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		s := randString(301, randGen)
		a := common.Alert{
			Footer: s,
		}
		a.Clean()
		assert.Equal(t, s[:297]+"...", a.Footer)
	})

	t.Run("field titles and values should be truncated when too long", func(t *testing.T) {
		t.Parallel()

		title := randString(31, randGen)
		value := randString(201, randGen)
		a := common.Alert{
			Fields: []*common.Field{
				{Title: title, Value: value},
			},
		}
		a.Clean()
		assert.Equal(t, title[:27]+"...", a.Fields[0].Title)
		assert.Equal(t, value[:197]+"...", a.Fields[0].Value)
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

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

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

		// Max webhooks is 5
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{}}
		for i := 1; i <= 6; i++ {
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

		// Url max length is 1000
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar/" + randString(986, randGen), ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is too long")

		// Url must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "foo", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is not a valid absolute URL")

		// Url must be absolute
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "/foo", ButtonText: "press me"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].url is not a valid absolute URL")

		// Button text is required
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: ""}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].buttonText is required")

		// Button text max length is 25
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: randString(26, randGen)}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].buttonText is too long")

		// Confirmation text max length is 1000
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", ConfirmationText: randString(1001, randGen)}}}
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

		// Max payload size is 50
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", Payload: map[string]interface{}{}}}}
		for i := 1; i <= 51; i++ {
			a.Webhooks[0].Payload[randString(10, randGen)] = randString(10, randGen)
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].payload item count is too large")

		// Max plain text input size is 10
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"}}}
		for i := 1; i <= 11; i++ {
			a.Webhooks[0].PlainTextInput = append(a.Webhooks[0].PlainTextInput, &common.WebhookPlainTextInput{ID: randString(5, randGen), Description: randString(5, randGen)})
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput item count is too large")

		// Max checkbox size is 10
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me"}}}
		for i := 1; i <= 11; i++ {
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

		// Input ID max length is 200
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: randString(201, randGen), Description: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].id is too long")

		// Input description max length is 200
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: randString(201, randGen)}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].description is too long")

		// Input minLength cannot be negative
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MinLength: -1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].minLength must be >=0")

		// Input minstLength cannt be larger than 3000
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MinLength: 3001}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].minLength must be <=3000")

		// Input maxLength cannot be negative
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MaxLength: -1}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].maxLength must be >=0")

		// Input maxLength cannot be larger than 3000
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", PlainTextInput: []*common.WebhookPlainTextInput{{ID: "foo", Description: "foo", MaxLength: 3001}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].plainTextInput[0].maxLength must be <=3000")

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

		// Checkbox ID max length is 200
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: randString(201, randGen), Label: "foo"}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].id is too long")

		// Checkbox label max length is 200
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: randString(201, randGen)}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].label is too long")

		// Checkbox options length cannot be larger than 5
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo"}}}}}
		for i := 1; i <= 6; i++ {
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

		// Checkbox option text max length is 50
		a = &common.Alert{Header: "a", RouteKey: "b", Webhooks: []*common.Webhook{{ID: "foo", URL: "http://foo.bar", ButtonText: "press me", CheckboxInput: []*common.WebhookCheckboxInput{{ID: "foo", Label: "foo", Options: []*common.WebhookCheckboxOption{{Value: "foo", Text: randString(51, randGen)}}}}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "webhook[0].checkboxInput[0].options[0].text is too long")
	})

	t.Run("alert.escalation should be on the correct format", func(t *testing.T) {
		t.Parallel()

		// Empty is OK
		a := &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{}}
		a.Clean()
		require.NoError(t, a.Validate())

		// Escalation delay must be at least 30 seconds
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 29, Severity: common.AlertError}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].delaySeconds '29' is too low")

		// Escalation delay must be at least 30 seconds larger than the previous escalation
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError}, {DelaySeconds: 59, Severity: common.AlertPanic}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[1].delaySeconds '59' is too small compared to previous escalation")

		// Escalation severity must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: "foo"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'foo' is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertInfo}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'info' is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertResolved}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].severity 'resolved' is not valid")

		// Escalation mentions count must be at most 10
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError, SlackMentions: []string{}}}}
		for i := 1; i <= 11; i++ {
			a.Escalation[0].SlackMentions = append(a.Escalation[0].SlackMentions, "<@foo>")
		}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions item count is too large")

		// Escalation mentions must be valid
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError, SlackMentions: []string{"foo"}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions[0] is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError, SlackMentions: []string{"<@" + randString(common.MaxMentionLength+1, randGen) + ">"}}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].slackMentions[0] is not valid")

		// Escalation moveToChannel must be a valid channel ID or channel name
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError, MoveToChannel: "foo bar"}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].moveToChannel is not valid")
		a = &common.Alert{Header: "a", RouteKey: "b", Escalation: []*common.Escalation{{DelaySeconds: 30, Severity: common.AlertError, MoveToChannel: randString(81, randGen)}}}
		a.Clean()
		require.ErrorContains(t, a.Validate(), "escalation[0].moveToChannel is not valid")
	})
}

var testLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int, randGen *rand.Rand) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = testLetters[randGen.Intn(len(testLetters))]
	}
	return string(b)
}

func hash(input ...string) string {
	h := sha256.New()

	for _, s := range input {
		h.Write([]byte(s))
	}

	bs := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(bs)
}
