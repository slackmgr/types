package common

import (
	"context"
	"encoding/json"
)

// DB is an interface for interacting with the database.
// It must be implemented by any database driver used by the Slack Manager.
type DB interface {
	// SaveAlert saves an alert to the database (for auditing purposes).
	// The same alert may be saved multiple times, in case of errors and retries.
	//
	// A database implementation can choose to skip saving the alerts, since they are never read by the manager.
	//
	// channelID is the Slack channel ID that the alert belongs to. It MUST match the channel ID of the alert.
	SaveAlert(ctx context.Context, channelID string, alert *Alert) error

	// SaveIssue creates or updates a single issue in the database.
	//
	// channelID is the Slack channel ID that the issue belongs to. It MUST match the channel ID of the issue.
	SaveIssue(ctx context.Context, channelID string, issue Issue) error

	// SaveIssues creates or updates multiple issues in the database.
	//
	// channelID is the Slack channel ID that the issues belong to. It MUST match the channel ID of each issue.
	SaveIssues(ctx context.Context, channelID string, issues ...Issue) error

	// FindOpenIssueByCorrelationID finds a single open issue in the database, based on the provided channel ID and correlation ID.
	//
	// The database implementation should return an error if the query matches multiple issues, and [nil, nil] if no issue is found.
	FindOpenIssueByCorrelationID(ctx context.Context, channelID, correlationID string) (json.RawMessage, error)

	// FindIssueBySlackPostID finds a single issue in the database, based on the provided channel ID and Slack post ID.
	//
	// The database implementation should return an error if the query matches multiple issues, and [nil, nil] if no issue is found.
	FindIssueBySlackPostID(ctx context.Context, channelID, postID string) (json.RawMessage, error)

	// LoadOpenIssues loads all open (non-archived) issues from the database, across all channels.
	// The returned list may be empty if no open issues are found.
	LoadOpenIssues(ctx context.Context) ([]json.RawMessage, error)

	// LoadOpenIssuesInChannel loads all open (non-archived) issues from the database, for the specified channel ID.
	// The returned list may be empty if no open issues are found in the channel.
	LoadOpenIssuesInChannel(ctx context.Context, channelID string) ([]json.RawMessage, error)

	// SaveMoveMapping creates or updates a single move mapping in the database.
	//
	// channelID is the Slack channel ID that the move mapping belongs to. It MUST match the channel ID of the move mapping.
	SaveMoveMapping(ctx context.Context, channelID string, moveMapping MoveMapping) error

	// FindMoveMapping finds a single move mapping in the database, for the specified channel ID and correlation ID.
	//
	// The database implementation should return an error if the query matches multiple mappings, and [nil, nil] if no mapping is found.
	FindMoveMapping(ctx context.Context, channelID, correlationID string) (json.RawMessage, error)

	// SaveChannelProcessingState creates or updates a single channel processing state in the database.
	//
	// channelID is the Slack channel ID that the processing state belongs to. It MUST match the channel ID of the processing state.
	SaveChannelProcessingState(ctx context.Context, channelID string, state ChannelProcessingState) error

	// FindChannelProcessingState finds a single channel processing state in the database, for the specified channel ID.
	//
	// The database implementation should return an error if the query matches multiple states, and [nil, nil] if no state is found.
	FindChannelProcessingState(ctx context.Context, channelID string) (json.RawMessage, error)
}
