package common

import (
	"context"
	"encoding/json"
)

// DB is an interface for interacting with the database.
// It must be implemented by any database driver used by the Slack Manager.
type DB interface {
	// Init initializes the database, for example by creating necessary tables or collections.
	// Set skipSchemaValidation to true to skip schema validation.
	Init(ctx context.Context, skipSchemaValidation bool) error

	// SaveAlert saves an alert to the database (for auditing purposes).
	// The same alert may be saved multiple times, in case of errors and retries.
	//
	// A database implementation can choose to skip saving the alerts, since they are never read by the manager.
	SaveAlert(ctx context.Context, alert *Alert) error

	// SaveIssue creates or updates a single issue in the database.
	SaveIssue(ctx context.Context, issue Issue) error

	// SaveIssues creates or updates multiple issues in the database.
	SaveIssues(ctx context.Context, issues ...Issue) error

	// MoveIssue moves an issue from one channel to another.
	// This channel ID in the issue must match the targetChannelID.
	// The sourceChannelID is used to find the existing issue in the database.
	MoveIssue(ctx context.Context, issue Issue, sourceChannelID, targetChannelID string) error

	// FindOpenIssueByCorrelationID finds a single open issue in the database, based on the provided channel ID and correlation ID.
	//
	// The database implementation should return an error if the query matches multiple issues, and [nil, nil] if no issue is found.
	FindOpenIssueByCorrelationID(ctx context.Context, channelID, correlationID string) (string, json.RawMessage, error)

	// FindIssueBySlackPostID finds a single issue in the database, based on the provided channel ID and Slack post ID.
	//
	// The database implementation should return an error if the query matches multiple issues, and [nil, nil] if no issue is found.
	FindIssueBySlackPostID(ctx context.Context, channelID, postID string) (string, json.RawMessage, error)

	// FindActiveChannels returns a list of all active channels in the database.
	// An active channel is one that has at least one open issue.
	// The returned list may be empty if no active channels are found.
	FindActiveChannels(ctx context.Context) ([]string, error)

	// LoadOpenIssuesInChannel loads all open (non-archived) issues from the database, for the specified channel ID.
	// The returned list may be empty if no open issues are found in the channel.
	LoadOpenIssuesInChannel(ctx context.Context, channelID string) (map[string]json.RawMessage, error)

	// SaveMoveMapping creates or updates a single move mapping in the database.
	SaveMoveMapping(ctx context.Context, moveMapping MoveMapping) error

	// FindMoveMapping finds a single move mapping in the database, for the specified channel ID and correlation ID.
	//
	// The database implementation should return an error if the query matches multiple mappings, and [nil, nil] if no mapping is found.
	FindMoveMapping(ctx context.Context, channelID, correlationID string) (json.RawMessage, error)

	// SaveChannelProcessingState creates or updates a single channel processing state in the database.
	SaveChannelProcessingState(ctx context.Context, state *ChannelProcessingState) error

	// FindChannelProcessingState finds a single channel processing state in the database, for the specified channel ID.
	//
	// The database implementation should return an error if the query matches multiple states, and [nil, nil] if no state is found.
	FindChannelProcessingState(ctx context.Context, channelID string) (*ChannelProcessingState, error)

	// DropAllData drops *all* data from the database.
	// This is useful for testing purposes, to reset the database state.
	// It should be used with caution, as it will remove all alerts, issues, move mappings, and processing states.
	DropAllData(ctx context.Context) error
}
