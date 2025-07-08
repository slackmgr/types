package common

import "encoding/json"

type Issue interface {
	json.Marshaler

	// UniqueID returns a unique and deterministic ID for this issue, for database/storage purposes.
	// The ID is based on certain issue fields, and is base64 encoded to ensure it is safe for use in URLs and as a database key.
	UniqueID() string

	// GetCorrelationID returns the correlation ID for this issue.
	// The correlation ID is used to group related alerts together, and may or may not be client defined.
	// It is not guaranteed to be unique across all issues, and *cannot* be used as a unique issue identifier.
	// It is not URL safe, and should thus be encoded before being used in URLs or as part of a database key.
	GetCorrelationID()

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

	// UniqueID returns a unique and deterministic ID for this move mapping, for database/storage purposes.
	// The ID is based on certain move mapping fields, and is base64 encoded to ensure it is safe for use in URLs and as a database key.
	UniqueID() string
}
