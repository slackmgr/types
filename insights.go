package types

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrInsightsNotSupported is returned by insights readers and services when the underlying storage does not support insights.
var ErrInsightsNotSupported = errors.New("insights not supported by the database implementation")

// ErrInsightsTooManyRows is returned by count and stats queries without an explicit [InsightsGrouping.Limit]
// when more than [MaxInsightsRows] groups match. Narrow the time range or filter, use a larger bucket, or group
// by fewer dimensions.
var ErrInsightsTooManyRows = errors.New("insights query matches too many rows")

// InsightsSink persists batches of [InsightsEvent] values.
//
// Implementations must honour the context deadline, must be safe for concurrent use, and must not share
// transactions or connection capacity with the regular alert and issue writes, so that insights can never
// slow down alert processing. Events are best-effort data: an implementation may return an error and lose
// the batch rather than retrying for a long time.
type InsightsSink interface {
	// WriteInsightsEvents persists the events. The slice must not be retained after the call returns.
	WriteInsightsEvents(ctx context.Context, events []InsightsEvent) error
}

// InsightsSinkProvider is optionally implemented by a [DB] implementation that supports recording insights.
// The Slack Manager discovers it with a type assertion.
type InsightsSinkProvider interface {
	// InsightsSink returns the sink, or nil if insights recording is disabled in the implementation.
	InsightsSink() InsightsSink
}

// InsightsReader answers insights queries.
//
// Implementations must only require read (SELECT) privileges, and must never modify data or schema.
// Queries must be validated with the Validate method of the query type before they are executed.
type InsightsReader interface {
	// CountInsightsEvents counts events matching the query filter, grouped by the query dimensions.
	// If [InsightsGrouping.Limit] is zero and more than [MaxInsightsRows] groups match, it returns
	// [ErrInsightsTooManyRows] rather than a truncated result.
	CountInsightsEvents(ctx context.Context, query *InsightsCountQuery) ([]*InsightsCountRow, error)

	// InsightsValueStats computes aggregates of one numeric event value (see [InsightsEvent.Values]) for events
	// matching the query filter, grouped by the query dimensions. Events without the value are ignored.
	// If [InsightsGrouping.Limit] is zero and more than [MaxInsightsRows] groups match, it returns
	// [ErrInsightsTooManyRows] rather than a truncated result.
	InsightsValueStats(ctx context.Context, query *InsightsStatsQuery) ([]*InsightsStatsRow, error)

	// ListInsightsEvents returns a page of events matching the query filter, ordered by time.
	ListInsightsEvents(ctx context.Context, query *InsightsListQuery) (*InsightsEventPage, error)

	// LoadOpenIssues loads all open (non-archived) issues in the given channels, or in all channels if channelIDs is empty.
	// The returned map is keyed by issue ID, with the opaque issue JSON as value.
	LoadOpenIssues(ctx context.Context, channelIDs []string) (map[string]json.RawMessage, error)
}

// InsightsReaderProvider is optionally implemented by a [DB] implementation that supports reading insights.
type InsightsReaderProvider interface {
	// InsightsReader returns the reader, or nil if insights reading is not available.
	InsightsReader() InsightsReader
}
