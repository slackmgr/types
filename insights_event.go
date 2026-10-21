package types

import "time"

// InsightsEventType identifies what an [InsightsEvent] describes.
//
// Types are namespaced by subject ("alert.", "issue.", "command.", "webhook.", "slack.", "channel.", "insights.").
// New types may be added in minor releases, so consumers should tolerate unknown types.
type InsightsEventType string

// Alert outcome events, recorded by the manager. Each processing attempt of a queued alert produces exactly one of these.
const (
	// InsightsEventAlertNewIssue means the alert finished processing (acked as successful) and opened a new issue.
	InsightsEventAlertNewIssue InsightsEventType = "alert.new_issue"

	// InsightsEventAlertGrouped means the alert finished processing and changed an existing open issue.
	InsightsEventAlertGrouped InsightsEventType = "alert.grouped"

	// InsightsEventAlertIgnored means the alert finished processing and was saved, but intentionally had no effect.
	// The reason is one of [InsightsReasonNotNewer], [InsightsReasonAlreadyResolved], [InsightsReasonFollowUpDisabled]
	// or [InsightsReasonResolvedWithoutIssue].
	InsightsEventAlertIgnored InsightsEventType = "alert.ignored"

	// InsightsEventAlertDropped means the alert was queued, but permanently discarded by the manager without being processed or saved.
	// The reason is one of [InsightsReasonInvalid], [InsightsReasonMissingEncryptionKey], [InsightsReasonCorruptIssue]
	// or [InsightsReasonUnreadable].
	InsightsEventAlertDropped InsightsEventType = "alert.dropped"

	// InsightsEventAlertRetried means a processing attempt was aborted and the alert was returned to the queue for redelivery.
	// This is not a terminal outcome. The reason is one of [InsightsReasonLockTimeout], [InsightsReasonGateTimeout],
	// [InsightsReasonSlackError], [InsightsReasonDBError] or [InsightsReasonMoveRace].
	InsightsEventAlertRetried InsightsEventType = "alert.retried"
)

// Alert intake events, recorded by the API server (when configured with an insights sink).
const (
	// InsightsEventAlertQueued means the API validated the alert and put it on the alert queue.
	InsightsEventAlertQueued InsightsEventType = "alert.queued"

	// InsightsEventAlertFiltered means the API discarded a valid alert because of a client-defined rule
	// (reason [InsightsReasonIgnoreTerm]). The client received a successful response.
	InsightsEventAlertFiltered InsightsEventType = "alert.filtered"

	// InsightsEventAlertRejected means the API refused the alert and the client received an error response.
	// The reason describes why, e.g. "validation", "no_route" or "rate_limited".
	InsightsEventAlertRejected InsightsEventType = "alert.rejected"
)

// Issue lifecycle events, recorded by the manager.
const (
	// InsightsEventIssueCreated means a new issue was created and saved.
	InsightsEventIssueCreated InsightsEventType = "issue.created"

	// InsightsEventIssueUpdated means an alert changed an existing open issue (one per [InsightsEventAlertGrouped]).
	InsightsEventIssueUpdated InsightsEventType = "issue.updated"

	// InsightsEventIssueResolved means an issue explicitly transitioned to resolved, by a resolved alert
	// ([InsightsReasonAlert]) or a user command ([InsightsReasonUser]). Resolution by auto-resolve timeout is
	// implicit and not recorded as a separate event; see [InsightsEventIssueArchived].
	InsightsEventIssueResolved InsightsEventType = "issue.resolved"

	// InsightsEventIssueReopened means a resolved issue (including auto-resolved) became unresolved again, by a
	// new alert ([InsightsReasonAlert]) or a user command ([InsightsReasonUser]).
	InsightsEventIssueReopened InsightsEventType = "issue.reopened"

	// InsightsEventIssueEscalated means an escalation rule raised the issue severity, or became active for the first time.
	InsightsEventIssueEscalated InsightsEventType = "issue.escalated"

	// InsightsEventIssueMoved means an issue was moved from ChannelID to RelatedChannelID,
	// by escalation ([InsightsReasonEscalation]) or a user command ([InsightsReasonUser]).
	InsightsEventIssueMoved InsightsEventType = "issue.moved"

	// InsightsEventIssueArchived means an issue was archived. The reason describes how the issue ended, e.g.
	// [InsightsReasonUserResolved] or [InsightsReasonAutoResolved]. The event carries lifecycle values such as
	// [InsightsValueOpenSeconds] and [InsightsValueTimeToResolveSeconds].
	InsightsEventIssueArchived InsightsEventType = "issue.archived"
)

// Command and webhook events, recorded by the manager. For command events, the reason is the command action
// (e.g. "resolve_issue", "mute_issue", "webhook") and the actor fields identify the Slack user.
const (
	// InsightsEventCommandExecuted means a user command was executed without error.
	InsightsEventCommandExecuted InsightsEventType = "command.executed"

	// InsightsEventCommandFailed means a user command failed. Commands are attempted exactly once.
	InsightsEventCommandFailed InsightsEventType = "command.failed"

	// InsightsEventCommandSkipped means a user command was ignored because no matching issue was found.
	InsightsEventCommandSkipped InsightsEventType = "command.skipped"

	// InsightsEventWebhookSucceeded means a webhook triggered by a user was delivered successfully. Detail holds the webhook ID.
	InsightsEventWebhookSucceeded InsightsEventType = "webhook.succeeded"

	// InsightsEventWebhookFailed means a webhook triggered by a user failed. Detail holds the webhook ID.
	InsightsEventWebhookFailed InsightsEventType = "webhook.failed"
)

// Slack post events, recorded by the manager.
const (
	// InsightsEventSlackPostCreated means a Slack post was created for an issue. The reason is the post action ("alert" or "resolve").
	InsightsEventSlackPostCreated InsightsEventType = "slack.post_created"

	// InsightsEventSlackPostUpdated means an existing Slack post was updated. The reason is the post action ("alert" or "resolve").
	InsightsEventSlackPostUpdated InsightsEventType = "slack.post_updated"

	// InsightsEventSlackPostDeleted means a Slack post was deleted (or replaced by a tombstone when it has replies).
	InsightsEventSlackPostDeleted InsightsEventType = "slack.post_deleted"

	// InsightsEventSlackPostThrottled means a Slack post update was deferred by throttling. Repeated deferrals each count.
	InsightsEventSlackPostThrottled InsightsEventType = "slack.post_throttled"

	// InsightsEventSlackPostFailed means a Slack post could not be created, updated or deleted.
	// The reason is one of [InsightsReasonChannelUnavailable], [InsightsReasonInvalidBlocks], [InsightsReasonNotInChannel],
	// [InsightsReasonMessageNotFound] or [InsightsReasonAPIError].
	InsightsEventSlackPostFailed InsightsEventType = "slack.post_failed"
)

// Channel and recorder events.
const (
	// InsightsEventChannelStats is a periodic snapshot of the open issues in a channel, recorded while the channel
	// has open issues. Values hold counts such as [InsightsValueOpen]. Detail holds the channel name at the time.
	InsightsEventChannelStats InsightsEventType = "channel.stats"

	// InsightsEventGap is recorded by an insights recorder after events were lost (e.g. buffer overflow or write
	// errors). Count holds the number of lost events since the previous gap event.
	InsightsEventGap InsightsEventType = "insights.gap"
)

// Reasons used by [InsightsEventAlertIgnored].
const (
	// InsightsReasonNotNewer means the alert timestamp was not newer than the last alert of the issue (including redeliveries).
	InsightsReasonNotNewer = "not_newer"

	// InsightsReasonAlreadyResolved means a resolved alert arrived for an issue that was already resolved.
	InsightsReasonAlreadyResolved = "already_resolved"

	// InsightsReasonFollowUpDisabled means the matching issue has issue follow-up disabled, so new alerts are ignored.
	InsightsReasonFollowUpDisabled = "follow_up_disabled"

	// InsightsReasonResolvedWithoutIssue means a resolved alert arrived while no open issue existed.
	InsightsReasonResolvedWithoutIssue = "resolved_without_issue"
)

// Reasons used by [InsightsEventAlertDropped].
const (
	// InsightsReasonInvalid means the alert failed validation in the manager.
	InsightsReasonInvalid = "invalid"

	// InsightsReasonMissingEncryptionKey means the alert has webhooks but no encryption key is configured.
	InsightsReasonMissingEncryptionKey = "missing_encryption_key"

	// InsightsReasonCorruptIssue means the matching issue in the database could not be read.
	InsightsReasonCorruptIssue = "corrupt_issue"

	// InsightsReasonUnreadable means the queue message could not be unmarshalled.
	InsightsReasonUnreadable = "unreadable"
)

// Reasons used by [InsightsEventAlertRetried].
const (
	// InsightsReasonLockTimeout means the channel lock could not be obtained in time.
	InsightsReasonLockTimeout = "lock_timeout"

	// InsightsReasonGateTimeout means processing was interrupted while waiting for the Slack rate limit gate.
	InsightsReasonGateTimeout = "gate_timeout"

	// InsightsReasonSlackError means a Slack API call failed.
	InsightsReasonSlackError = "slack_error"

	// InsightsReasonDBError means a database call failed.
	InsightsReasonDBError = "db_error"

	// InsightsReasonMoveRace means the alert was routed to the wrong channel due to a concurrent issue move.
	InsightsReasonMoveRace = "move_race"
)

// Reasons used by [InsightsEventAlertFiltered].
const (
	// InsightsReasonIgnoreTerm means the alert text contained one of the alert's ignoreIfTextContains terms.
	InsightsReasonIgnoreTerm = "ignore_term"
)

// Reasons used by [InsightsEventIssueResolved], [InsightsEventIssueReopened] and [InsightsEventIssueMoved].
const (
	// InsightsReasonAlert means the transition was caused by an alert.
	InsightsReasonAlert = "alert"

	// InsightsReasonUser means the transition was caused by a user command.
	InsightsReasonUser = "user"

	// InsightsReasonEscalation means the issue was moved by an escalation rule.
	InsightsReasonEscalation = "escalation"
)

// Reasons used by [InsightsEventIssueArchived], describing how the issue ended.
const (
	// InsightsReasonUserResolved means the issue was resolved by a user.
	InsightsReasonUserResolved = "user_resolved"

	// InsightsReasonAlertResolved means the issue was resolved by a resolved alert.
	InsightsReasonAlertResolved = "alert_resolved"

	// InsightsReasonAutoResolved means the issue was resolved by the auto-resolve timeout.
	InsightsReasonAutoResolved = "auto_resolved"

	// InsightsReasonAutoResolvedInconclusive means the issue was auto-resolved and flagged as inconclusive.
	InsightsReasonAutoResolvedInconclusive = "auto_resolved_inconclusive"

	// InsightsReasonTerminated means the issue was terminated by a user.
	InsightsReasonTerminated = "terminated"

	// InsightsReasonFireAndForget means the issue had follow-up disabled and was archived after its archiving delay.
	InsightsReasonFireAndForget = "fire_and_forget"
)

// Reasons used by [InsightsEventSlackPostFailed].
const (
	// InsightsReasonChannelUnavailable means the Slack channel was not found or is archived.
	InsightsReasonChannelUnavailable = "channel_unavailable"

	// InsightsReasonInvalidBlocks means Slack rejected the message formatting.
	InsightsReasonInvalidBlocks = "invalid_blocks"

	// InsightsReasonNotInChannel means the Slack app is not a member of the channel.
	InsightsReasonNotInChannel = "not_in_channel"

	// InsightsReasonMessageNotFound means the Slack post to update no longer exists, or cannot be updated.
	InsightsReasonMessageNotFound = "message_not_found"

	// InsightsReasonAPIError means any other Slack API error.
	InsightsReasonAPIError = "api_error"
)

// Value keys used in [InsightsEvent.Values].
const (
	// InsightsValueOpenSeconds is the time from issue creation to archiving, in seconds ([InsightsEventIssueArchived]).
	InsightsValueOpenSeconds = "open_s"

	// InsightsValueTimeToResolveSeconds is the time from issue creation to (last) resolution, in seconds ([InsightsEventIssueArchived]).
	InsightsValueTimeToResolveSeconds = "time_to_resolve_s"

	// InsightsValueTimeToFirstInvestigationSeconds is the time from issue creation to the first investigate command,
	// in seconds ([InsightsEventIssueArchived]). Absent if the issue was never investigated.
	InsightsValueTimeToFirstInvestigationSeconds = "time_to_first_investigation_s"

	// InsightsValueAlertCount is the total number of alerts in the issue ([InsightsEventIssueArchived]).
	InsightsValueAlertCount = "alert_count"

	// InsightsValueReopenCount is the number of times the issue was reopened ([InsightsEventIssueArchived]).
	InsightsValueReopenCount = "reopen_count"

	// InsightsValueNotified is 1 if an alert Slack post was ever sent for the issue, otherwise 0 ([InsightsEventIssueArchived]).
	InsightsValueNotified = "notified"

	// InsightsValueNotificationDelaySeconds is the configured notification delay of a new issue, in seconds ([InsightsEventIssueCreated]).
	InsightsValueNotificationDelaySeconds = "notification_delay_s"

	// InsightsValueOpen is the number of open (non-archived) issues ([InsightsEventChannelStats]).
	InsightsValueOpen = "open"

	// InsightsValueAlerting is the number of open issues that are neither resolved nor info ([InsightsEventChannelStats]).
	InsightsValueAlerting = "alerting"

	// InsightsValueResolved is the number of open issues that are resolved but not yet archived ([InsightsEventChannelStats]).
	InsightsValueResolved = "resolved"

	// InsightsValueMuted is the number of muted open issues ([InsightsEventChannelStats]).
	InsightsValueMuted = "muted"

	// InsightsValueInvestigated is the number of open issues under investigation ([InsightsEventChannelStats]).
	InsightsValueInvestigated = "investigated"

	// InsightsValueEscalated is the number of escalated open issues ([InsightsEventChannelStats]).
	InsightsValueEscalated = "escalated"
)

// InsightsMaxDetailLen is the maximum length of [InsightsEvent.Detail]. Longer values are truncated by [InsightsEvent.Normalize].
const InsightsMaxDetailLen = 200

// InsightsEvent is a structured record of something the Slack Manager did, used for dashboards and reporting.
//
// Events carry identifiers and classifications only, never alert content such as headers, texts or fields.
// Only Time and Type are required; all other fields are optional and empty when not applicable.
type InsightsEvent struct {
	// Time is when the event happened. Recorders set it to the current time if it is zero.
	Time time.Time `json:"time"`

	// Type identifies what the event describes.
	Type InsightsEventType `json:"type"`

	// Reason is an optional sub-classification. The possible values depend on Type.
	Reason string `json:"reason,omitempty"`

	// ChannelID is the Slack channel where the event happened. It may be empty for API rejections before routing.
	ChannelID string `json:"channelId,omitempty"`

	// RelatedChannelID is a second channel involved in the event, e.g. the target channel of a move,
	// or the original channel of an alert that was redirected by a move mapping.
	RelatedChannelID string `json:"relatedChannelId,omitempty"`

	// IssueID is the unique ID of the issue the event relates to.
	IssueID string `json:"issueId,omitempty"`

	// CorrelationID is the correlation ID of the alert or issue.
	CorrelationID string `json:"correlationId,omitempty"`

	// Severity is the alert or issue severity after the event.
	Severity AlertSeverity `json:"severity,omitempty"`

	// PrevSeverity is the issue severity before the event, for events that change severity.
	PrevSeverity AlertSeverity `json:"prevSeverity,omitempty"`

	// RouteKey is the route key of the alert.
	RouteKey string `json:"routeKey,omitempty"`

	// AlertType is the alert type.
	AlertType string `json:"alertType,omitempty"`

	// Host is the alert host.
	Host string `json:"host,omitempty"`

	// Author is the alert author.
	Author string `json:"author,omitempty"`

	// ActorID is the Slack user ID of the user who caused the event (commands).
	ActorID string `json:"actorId,omitempty"`

	// ActorName is the real name of the user who caused the event (commands).
	ActorName string `json:"actorName,omitempty"`

	// Count is an event-specific count, e.g. the issue alert count at the time of the event.
	Count int `json:"count,omitempty"`

	// Detail is a short, event-specific free text (at most [InsightsMaxDetailLen] characters).
	Detail string `json:"detail,omitempty"`

	// Values holds event-specific numeric facts, keyed by the InsightsValue* constants. It is only used on
	// low-frequency events, to keep the cost of recording high-frequency events minimal.
	Values map[string]float64 `json:"values,omitempty"`
}

// Normalize prepares the event for storage: it sets Time to now (in UTC) if zero, converts Time to UTC,
// and truncates Detail to [InsightsMaxDetailLen] characters. It is called by recorders before writing.
func (e *InsightsEvent) Normalize(now time.Time) {
	if e.Time.IsZero() {
		e.Time = now
	}

	e.Time = e.Time.UTC()

	if len(e.Detail) > InsightsMaxDetailLen {
		e.Detail = truncateUTF8(e.Detail, InsightsMaxDetailLen)
	}
}

// truncateUTF8 truncates s to at most maxBytes bytes, without splitting a multi-byte UTF-8 character.
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	cut := maxBytes

	// Back off to the start of a UTF-8 character (continuation bytes are 10xxxxxx).
	for cut > 0 && s[cut]&0xC0 == 0x80 {
		cut--
	}

	return s[:cut]
}
