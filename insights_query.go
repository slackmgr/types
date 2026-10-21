package types

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"
)

// Limits enforced by the Validate methods of the insights query types.
const (
	// MaxInsightsQueryRange is the maximum time range of an insights query.
	MaxInsightsQueryRange = 366 * 24 * time.Hour

	// MinInsightsBucket is the minimum time bucket size.
	MinInsightsBucket = time.Minute

	// MaxInsightsBuckets is the maximum number of time buckets in a single query.
	MaxInsightsBuckets = 2000

	// MaxInsightsRows is the maximum number of rows returned by a count or stats query.
	MaxInsightsRows = 10000

	// MaxInsightsPageSize is the maximum number of events in a single page.
	MaxInsightsPageSize = 1000

	// DefaultInsightsPageSize is the page size used when [InsightsListQuery.Limit] is zero.
	DefaultInsightsPageSize = 100

	// MaxInsightsFilterValues is the maximum number of values in each list filter of [InsightsFilter].
	MaxInsightsFilterValues = 1000
)

var insightsValueKeyRegex = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// InsightsDimension is an event attribute that insights queries can group by.
type InsightsDimension string

// Supported insights dimensions.
const (
	// InsightsDimensionBucket groups by time bucket (the start of the bucket). Requires a bucket size in the query.
	InsightsDimensionBucket InsightsDimension = "bucket"

	// InsightsDimensionChannel groups by [InsightsEvent.ChannelID].
	InsightsDimensionChannel InsightsDimension = "channel"

	// InsightsDimensionRelatedChannel groups by [InsightsEvent.RelatedChannelID].
	InsightsDimensionRelatedChannel InsightsDimension = "related_channel"

	// InsightsDimensionType groups by [InsightsEvent.Type].
	InsightsDimensionType InsightsDimension = "type"

	// InsightsDimensionReason groups by [InsightsEvent.Reason].
	InsightsDimensionReason InsightsDimension = "reason"

	// InsightsDimensionSeverity groups by [InsightsEvent.Severity].
	InsightsDimensionSeverity InsightsDimension = "severity"

	// InsightsDimensionRouteKey groups by [InsightsEvent.RouteKey].
	InsightsDimensionRouteKey InsightsDimension = "route_key"

	// InsightsDimensionAlertType groups by [InsightsEvent.AlertType].
	InsightsDimensionAlertType InsightsDimension = "alert_type"

	// InsightsDimensionHost groups by [InsightsEvent.Host].
	InsightsDimensionHost InsightsDimension = "host"

	// InsightsDimensionAuthor groups by [InsightsEvent.Author].
	InsightsDimensionAuthor InsightsDimension = "author"

	// InsightsDimensionActor groups by [InsightsEvent.ActorID].
	InsightsDimensionActor InsightsDimension = "actor"

	// InsightsDimensionCorrelationID groups by [InsightsEvent.CorrelationID].
	InsightsDimensionCorrelationID InsightsDimension = "correlation_id"

	// InsightsDimensionIssue groups by [InsightsEvent.IssueID].
	InsightsDimensionIssue InsightsDimension = "issue"
)

// IsValid returns true if d is a supported dimension.
func (d InsightsDimension) IsValid() bool {
	switch d {
	case InsightsDimensionBucket, InsightsDimensionChannel, InsightsDimensionRelatedChannel, InsightsDimensionType,
		InsightsDimensionReason, InsightsDimensionSeverity, InsightsDimensionRouteKey, InsightsDimensionAlertType,
		InsightsDimensionHost, InsightsDimensionAuthor, InsightsDimensionActor, InsightsDimensionCorrelationID,
		InsightsDimensionIssue:
		return true
	default:
		return false
	}
}

// InsightsAggregate is an aggregate function computed by [InsightsReader.InsightsValueStats].
type InsightsAggregate string

// Supported insights aggregates.
const (
	// InsightsAggregateSum is the sum of the values.
	InsightsAggregateSum InsightsAggregate = "sum"

	// InsightsAggregateAvg is the arithmetic mean of the values.
	InsightsAggregateAvg InsightsAggregate = "avg"

	// InsightsAggregateMin is the smallest value.
	InsightsAggregateMin InsightsAggregate = "min"

	// InsightsAggregateMax is the largest value.
	InsightsAggregateMax InsightsAggregate = "max"

	// InsightsAggregateP50 is the median (continuous percentile).
	InsightsAggregateP50 InsightsAggregate = "p50"

	// InsightsAggregateP90 is the 90th percentile (continuous percentile).
	InsightsAggregateP90 InsightsAggregate = "p90"

	// InsightsAggregateP99 is the 99th percentile (continuous percentile).
	InsightsAggregateP99 InsightsAggregate = "p99"
)

// IsValid returns true if a is a supported aggregate.
func (a InsightsAggregate) IsValid() bool {
	switch a {
	case InsightsAggregateSum, InsightsAggregateAvg, InsightsAggregateMin, InsightsAggregateMax,
		InsightsAggregateP50, InsightsAggregateP90, InsightsAggregateP99:
		return true
	default:
		return false
	}
}

// InsightsFilter selects insights events. Events match when their time is in [From, To) and every non-empty
// filter field matches. List fields match any of their values.
type InsightsFilter struct {
	// From is the inclusive start of the time range. Required.
	From time.Time

	// To is the exclusive end of the time range. Required, and must be after From.
	To time.Time

	// ChannelIDs matches [InsightsEvent.ChannelID].
	ChannelIDs []string

	// RelatedChannelIDs matches [InsightsEvent.RelatedChannelID].
	RelatedChannelIDs []string

	// Types matches [InsightsEvent.Type].
	Types []InsightsEventType

	// Reasons matches [InsightsEvent.Reason].
	Reasons []string

	// Severities matches [InsightsEvent.Severity].
	Severities []AlertSeverity

	// RouteKeys matches [InsightsEvent.RouteKey].
	RouteKeys []string

	// IssueID matches [InsightsEvent.IssueID].
	IssueID string

	// CorrelationID matches [InsightsEvent.CorrelationID].
	CorrelationID string

	// ActorID matches [InsightsEvent.ActorID].
	ActorID string
}

// Validate returns an error if the filter is invalid.
func (f *InsightsFilter) Validate() error {
	if f.From.IsZero() || f.To.IsZero() {
		return errors.New("from and to are required")
	}

	if !f.To.After(f.From) {
		return errors.New("to must be after from")
	}

	if f.To.Sub(f.From) > MaxInsightsQueryRange {
		return fmt.Errorf("time range must not exceed %s", MaxInsightsQueryRange)
	}

	for name, n := range map[string]int{
		"channel IDs":         len(f.ChannelIDs),
		"related channel IDs": len(f.RelatedChannelIDs),
		"types":               len(f.Types),
		"reasons":             len(f.Reasons),
		"severities":          len(f.Severities),
		"route keys":          len(f.RouteKeys),
	} {
		if n > MaxInsightsFilterValues {
			return fmt.Errorf("too many %s in filter: %d (max %d)", name, n, MaxInsightsFilterValues)
		}
	}

	return nil
}

// InsightsGrouping defines how events are grouped in count and stats queries.
type InsightsGrouping struct {
	// GroupBy lists the dimensions to group by. Empty means a single total row.
	GroupBy []InsightsDimension

	// Bucket is the time bucket size. Required if GroupBy contains [InsightsDimensionBucket].
	// Buckets are aligned to the Unix epoch, so e.g. hourly buckets start on the hour (UTC).
	Bucket time.Duration

	// Limit is the maximum number of rows to return. Zero means [MaxInsightsRows].
	Limit int
}

// HasDimension returns true if GroupBy contains d.
func (g *InsightsGrouping) HasDimension(d InsightsDimension) bool {
	return slices.Contains(g.GroupBy, d)
}

// EffectiveLimit returns Limit, or [MaxInsightsRows] if Limit is zero.
func (g *InsightsGrouping) EffectiveLimit() int {
	if g.Limit == 0 {
		return MaxInsightsRows
	}

	return g.Limit
}

// validate returns an error if the grouping is invalid for the given filter.
func (g *InsightsGrouping) validate(filter *InsightsFilter) error {
	seen := make(map[InsightsDimension]bool, len(g.GroupBy))

	for _, d := range g.GroupBy {
		if !d.IsValid() {
			return fmt.Errorf("invalid dimension %q", d)
		}

		if seen[d] {
			return fmt.Errorf("duplicate dimension %q", d)
		}

		seen[d] = true
	}

	if seen[InsightsDimensionBucket] {
		if g.Bucket < MinInsightsBucket {
			return fmt.Errorf("bucket size must be at least %s", MinInsightsBucket)
		}

		buckets := int64(filter.To.Sub(filter.From)/g.Bucket) + 1
		if buckets > MaxInsightsBuckets {
			return fmt.Errorf("too many buckets: %d (max %d)", buckets, MaxInsightsBuckets)
		}
	} else if g.Bucket != 0 {
		return errors.New("bucket size is only allowed when grouping by bucket")
	}

	if g.Limit < 0 || g.Limit > MaxInsightsRows {
		return fmt.Errorf("limit must be between 0 and %d", MaxInsightsRows)
	}

	return nil
}

// InsightsCountQuery counts events, optionally grouped. See [InsightsReader.CountInsightsEvents].
type InsightsCountQuery struct {
	InsightsFilter
	InsightsGrouping

	// Distinct, if set, counts distinct non-empty values of this dimension instead of events
	// (e.g. distinct issues). [InsightsDimensionBucket] is not allowed.
	Distinct InsightsDimension

	// OrderByCountDesc orders rows by count, highest first (for top-N queries). Otherwise rows are ordered by the group keys.
	OrderByCountDesc bool
}

// Validate returns an error if the query is invalid.
func (q *InsightsCountQuery) Validate() error {
	if err := q.InsightsFilter.Validate(); err != nil {
		return err
	}

	if err := q.validate(&q.InsightsFilter); err != nil {
		return err
	}

	if q.Distinct != "" && (!q.Distinct.IsValid() || q.Distinct == InsightsDimensionBucket) {
		return fmt.Errorf("invalid distinct dimension %q", q.Distinct)
	}

	return nil
}

// InsightsStatsQuery computes aggregates of a numeric event value. See [InsightsReader.InsightsValueStats].
type InsightsStatsQuery struct {
	InsightsFilter
	InsightsGrouping

	// ValueKey is the key in [InsightsEvent.Values] to aggregate, e.g. [InsightsValueOpenSeconds]. Required.
	ValueKey string

	// Aggregates lists the aggregates to compute. Required.
	Aggregates []InsightsAggregate
}

// Validate returns an error if the query is invalid.
func (q *InsightsStatsQuery) Validate() error {
	if err := q.InsightsFilter.Validate(); err != nil {
		return err
	}

	if err := q.validate(&q.InsightsFilter); err != nil {
		return err
	}

	if !insightsValueKeyRegex.MatchString(q.ValueKey) {
		return fmt.Errorf("invalid value key %q", q.ValueKey)
	}

	if len(q.Aggregates) == 0 {
		return errors.New("at least one aggregate is required")
	}

	for _, a := range q.Aggregates {
		if !a.IsValid() {
			return fmt.Errorf("invalid aggregate %q", a)
		}
	}

	return nil
}

// InsightsListQuery lists events. See [InsightsReader.ListInsightsEvents].
type InsightsListQuery struct {
	InsightsFilter

	// Cursor continues a previous listing. It must be a NextCursor value returned for the same query, or empty.
	Cursor string

	// Limit is the maximum number of events to return. Zero means [DefaultInsightsPageSize].
	Limit int

	// Descending orders events newest first. Otherwise oldest first.
	Descending bool
}

// EffectiveLimit returns Limit, or [DefaultInsightsPageSize] if Limit is zero.
func (q *InsightsListQuery) EffectiveLimit() int {
	if q.Limit == 0 {
		return DefaultInsightsPageSize
	}

	return q.Limit
}

// Validate returns an error if the query is invalid.
func (q *InsightsListQuery) Validate() error {
	if err := q.InsightsFilter.Validate(); err != nil {
		return err
	}

	if q.Limit < 0 || q.Limit > MaxInsightsPageSize {
		return fmt.Errorf("limit must be between 0 and %d", MaxInsightsPageSize)
	}

	return nil
}

// InsightsGroupKey holds the group values of a count or stats row. Only the fields of the dimensions in the
// query's GroupBy are set.
type InsightsGroupKey struct {
	Bucket           time.Time         `json:"bucket,omitzero"`
	ChannelID        string            `json:"channelId,omitempty"`
	RelatedChannelID string            `json:"relatedChannelId,omitempty"`
	Type             InsightsEventType `json:"type,omitempty"`
	Reason           string            `json:"reason,omitempty"`
	Severity         AlertSeverity     `json:"severity,omitempty"`
	RouteKey         string            `json:"routeKey,omitempty"`
	AlertType        string            `json:"alertType,omitempty"`
	Host             string            `json:"host,omitempty"`
	Author           string            `json:"author,omitempty"`
	ActorID          string            `json:"actorId,omitempty"`
	CorrelationID    string            `json:"correlationId,omitempty"`
	IssueID          string            `json:"issueId,omitempty"`
}

// InsightsCountRow is a row returned by [InsightsReader.CountInsightsEvents].
type InsightsCountRow struct {
	InsightsGroupKey

	// Count is the number of matching events (or distinct values, see [InsightsCountQuery.Distinct]).
	Count int64 `json:"count"`
}

// InsightsStatsRow is a row returned by [InsightsReader.InsightsValueStats].
type InsightsStatsRow struct {
	InsightsGroupKey

	// Count is the number of events with the value.
	Count int64 `json:"count"`

	// Values holds the computed aggregates.
	Values map[InsightsAggregate]float64 `json:"values"`
}

// InsightsEventPage is a page of events returned by [InsightsReader.ListInsightsEvents].
type InsightsEventPage struct {
	// Events holds the events in the page.
	Events []*InsightsEvent `json:"events"`

	// NextCursor continues the listing. It is empty when there are no more events.
	NextCursor string `json:"nextCursor,omitempty"`
}
