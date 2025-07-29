package common

import "encoding/json"

type Issue interface {
	json.Marshaler

	// ChannelID returns the Slack channel ID that this issue belongs to.
	ChannelID() string

	// UniqueID returns a unique and deterministic ID for this issue, for database/storage purposes.
	// The ID is based on certain issue fields, and is base64 encoded to ensure it is safe for use in URLs and as a database key.
	UniqueID() string

	// GetCorrelationID returns the correlation ID for this issue.
	// The correlation ID is used to group related alerts together, and may or may not be client defined.
	// It is not guaranteed to be unique across all issues, even in the same channel, and *cannot* be used as a unique issue identifier.
	// It is not URL safe, and should thus be encoded before being used in URLs or as part of a database key.
	GetCorrelationID() string

	// IsOpen returns true if this issue is currently open (i.e. not archived).
	// A resolved issue is still considered open until it is archived.
	IsOpen() bool

	// CurrentPostID returns the current Slack post ID associated with this issue.
	// The value may change over time as the issue is updated.
	// If the issue has no current post, it returns an empty string.
	CurrentPostID() string
}

type MoveMapping interface {
	json.Marshaler

	// ChannelID returns the Slack channel ID that this move mapping belongs to (i.e. the channel where the move was initiated).
	ChannelID() string

	// UniqueID returns a unique and deterministic ID for this move mapping, for database/storage purposes.
	// The ID is based on the original channel and the correlation ID, and is base64 encoded to ensure it is safe for use in URLs and as a database key.
	UniqueID() string

	// GetCorrelationID returns the correlation ID that this move mapping is associated with.
	// It is unique in a single channel, but not across all channels.
	// It is not URL safe, and should thus be encoded before being used in URLs or as part of a database key.
	GetCorrelationID() string
}
