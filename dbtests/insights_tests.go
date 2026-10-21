//nolint:goconst
package dbtests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunAllInsightsTests runs all insights compliance tests against an implementation of [types.InsightsSink] and
// [types.InsightsReader]. The db is used to create issues for [TestInsightsLoadOpenIssues].
//
// Each test uses unique channel IDs, so the tests can run against a database that already contains other events.
// Events are written with timestamps within the last two hours.
func RunAllInsightsTests(t *testing.T, db types.DB, sink types.InsightsSink, reader types.InsightsReader) {
	t.Helper()

	t.Run("InsightsWriteAndList", func(t *testing.T) { TestInsightsWriteAndList(t, sink, reader) })
	t.Run("InsightsListPagination", func(t *testing.T) { TestInsightsListPagination(t, sink, reader) })
	t.Run("InsightsCount", func(t *testing.T) { TestInsightsCount(t, sink, reader) })
	t.Run("InsightsCountFilters", func(t *testing.T) { TestInsightsCountFilters(t, sink, reader) })
	t.Run("InsightsValueStats", func(t *testing.T) { TestInsightsValueStats(t, sink, reader) })
	t.Run("InsightsTooManyRows", func(t *testing.T) { TestInsightsTooManyRows(t, sink, reader) })
	t.Run("InsightsLoadOpenIssues", func(t *testing.T) { TestInsightsLoadOpenIssues(t, db, reader) })
}

// insightsTestBase returns an hour-aligned base time, one hour before the current hour, for deterministic buckets.
func insightsTestBase() time.Time {
	return time.Now().UTC().Truncate(time.Hour).Add(-time.Hour)
}

func uniqueChannelID() string {
	return "C" + uuid.New().String()[:12]
}

// TestInsightsWriteAndList verifies that all event fields round-trip through the sink and reader.
func TestInsightsWriteAndList(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel := uniqueChannelID()

	full := types.InsightsEvent{
		Time:             base.Add(5*time.Minute + 123456*time.Microsecond),
		Type:             types.InsightsEventIssueArchived,
		Reason:           types.InsightsReasonUserResolved,
		ChannelID:        channel,
		RelatedChannelID: "C_RELATED",
		IssueID:          "issue-1",
		CorrelationID:    "corr/with special chars æøå",
		Severity:         types.AlertResolved,
		PrevSeverity:     types.AlertError,
		RouteKey:         "route-1",
		AlertType:        "type-1",
		Host:             "host-1",
		Author:           "author-1",
		ActorID:          "U123",
		ActorName:        "Jane Doe",
		Count:            42,
		Detail:           "some detail",
		Values:           map[string]float64{types.InsightsValueOpenSeconds: 3600.5, types.InsightsValueNotified: 1},
	}
	minimal := types.InsightsEvent{
		Time:      base.Add(10 * time.Minute),
		Type:      types.InsightsEventAlertNewIssue,
		ChannelID: channel,
	}

	require.NoError(t, sink.WriteInsightsEvents(ctx, []types.InsightsEvent{minimal, full}))

	query := &types.InsightsListQuery{InsightsFilter: types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel}}}
	require.NoError(t, query.Validate())

	page, err := reader.ListInsightsEvents(ctx, query)
	require.NoError(t, err)
	require.Len(t, page.Events, 2)
	assert.Empty(t, page.NextCursor)

	got := page.Events[0]
	assert.True(t, full.Time.Equal(got.Time), "time should round-trip with microsecond precision: want %s, got %s", full.Time, got.Time)
	got.Time = full.Time
	assert.Equal(t, full, *got)

	gotMinimal := page.Events[1]
	assert.True(t, minimal.Time.Equal(gotMinimal.Time))
	assert.Equal(t, minimal.Type, gotMinimal.Type)
	assert.Equal(t, minimal.ChannelID, gotMinimal.ChannelID)
	assert.Empty(t, gotMinimal.Reason)
	assert.Empty(t, gotMinimal.Values)

	// Descending order
	query.Descending = true
	page, err = reader.ListInsightsEvents(ctx, query)
	require.NoError(t, err)
	require.Len(t, page.Events, 2)
	assert.Equal(t, types.InsightsEventAlertNewIssue, page.Events[0].Type)

	// Time range is [From, To)
	query = &types.InsightsListQuery{InsightsFilter: types.InsightsFilter{From: base, To: minimal.Time, ChannelIDs: []string{channel}}}
	page, err = reader.ListInsightsEvents(ctx, query)
	require.NoError(t, err)
	require.Len(t, page.Events, 1)
	assert.Equal(t, types.InsightsEventIssueArchived, page.Events[0].Type)

	// Issue filter (timeline)
	query = &types.InsightsListQuery{InsightsFilter: types.InsightsFilter{From: base, To: base.Add(time.Hour), IssueID: "issue-1"}}
	page, err = reader.ListInsightsEvents(ctx, query)
	require.NoError(t, err)
	require.NotEmpty(t, page.Events)

	for _, e := range page.Events {
		assert.Equal(t, "issue-1", e.IssueID)
	}
}

// TestInsightsListPagination verifies cursor-based pagination in both directions.
func TestInsightsListPagination(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel := uniqueChannelID()

	// 7 events, two of them with identical timestamps to exercise tie-breaking.
	events := make([]types.InsightsEvent, 0, 7)
	for i := range 6 {
		events = append(events, types.InsightsEvent{Time: base.Add(time.Duration(i) * time.Minute), Type: types.InsightsEventAlertGrouped, ChannelID: channel, Count: i})
	}
	events = append(events, types.InsightsEvent{Time: base.Add(5 * time.Minute), Type: types.InsightsEventAlertGrouped, ChannelID: channel, Count: 6})

	require.NoError(t, sink.WriteInsightsEvents(ctx, events))

	for _, descending := range []bool{false, true} {
		query := &types.InsightsListQuery{
			InsightsFilter: types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel}},
			Limit:          3,
			Descending:     descending,
		}

		seen := map[int]bool{}
		pages := 0
		var prev time.Time

		for {
			page, err := reader.ListInsightsEvents(ctx, query)
			require.NoError(t, err)
			pages++

			for _, e := range page.Events {
				assert.False(t, seen[e.Count], "event %d returned twice", e.Count)
				seen[e.Count] = true

				if !prev.IsZero() {
					if descending {
						assert.False(t, e.Time.After(prev), "events must be ordered newest first")
					} else {
						assert.False(t, e.Time.Before(prev), "events must be ordered oldest first")
					}
				}

				prev = e.Time
			}

			if page.NextCursor == "" {
				break
			}

			require.Less(t, pages, 10, "pagination does not terminate")
			query.Cursor = page.NextCursor
		}

		assert.Len(t, seen, 7, "descending=%v", descending)
	}
}

// TestInsightsCount verifies grouping, time buckets, distinct counts and top-N ordering.
func TestInsightsCount(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel1 := uniqueChannelID()
	channel2 := uniqueChannelID()

	ev := func(offset time.Duration, typ types.InsightsEventType, reason, channel, issue string) types.InsightsEvent {
		return types.InsightsEvent{Time: base.Add(offset), Type: typ, Reason: reason, ChannelID: channel, IssueID: issue}
	}

	require.NoError(t, sink.WriteInsightsEvents(ctx, []types.InsightsEvent{
		ev(1*time.Minute, types.InsightsEventAlertIgnored, types.InsightsReasonNotNewer, channel1, "i1"),
		ev(2*time.Minute, types.InsightsEventAlertIgnored, types.InsightsReasonNotNewer, channel1, "i1"),
		ev(3*time.Minute, types.InsightsEventAlertIgnored, types.InsightsReasonAlreadyResolved, channel1, "i2"),
		ev(16*time.Minute, types.InsightsEventIssueUpdated, "", channel1, "i1"),
		ev(17*time.Minute, types.InsightsEventIssueUpdated, "", channel1, "i1"),
		ev(18*time.Minute, types.InsightsEventIssueUpdated, "", channel1, "i2"),
		ev(31*time.Minute, types.InsightsEventIssueUpdated, "", channel2, "i3"),
	}))

	filter := types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel1, channel2}}

	t.Run("total", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{InsightsFilter: filter})
		require.Len(t, rows, 1)
		assert.Equal(t, int64(7), rows[0].Count)
	})

	t.Run("group by type and reason", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter:   filter,
			InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionType, types.InsightsDimensionReason}},
		})

		counts := map[string]int64{}
		for _, r := range rows {
			assert.Empty(t, r.ChannelID, "non-grouped dimensions must be empty")
			counts[string(r.Type)+"/"+r.Reason] = r.Count
		}

		assert.Equal(t, map[string]int64{
			"alert.ignored/not_newer":        2,
			"alert.ignored/already_resolved": 1,
			"issue.updated/":                 4,
		}, counts)
	})

	t.Run("group by 15 minute bucket and channel", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter: filter,
			InsightsGrouping: types.InsightsGrouping{
				GroupBy: []types.InsightsDimension{types.InsightsDimensionBucket, types.InsightsDimensionChannel},
				Bucket:  15 * time.Minute,
			},
		})

		counts := map[string]int64{}
		for _, r := range rows {
			counts[r.Bucket.UTC().Format("15:04")+"/"+r.ChannelID] = r.Count
		}

		assert.Equal(t, map[string]int64{
			base.Format("15:04") + "/" + channel1:                     3,
			base.Add(15*time.Minute).Format("15:04") + "/" + channel1: 3,
			base.Add(30*time.Minute).Format("15:04") + "/" + channel2: 1,
		}, counts)
	})

	t.Run("distinct issues per channel", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter:   types.InsightsFilter{From: filter.From, To: filter.To, ChannelIDs: filter.ChannelIDs, Types: []types.InsightsEventType{types.InsightsEventIssueUpdated}},
			InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionChannel}},
			Distinct:         types.InsightsDimensionIssue,
		})

		counts := map[string]int64{}
		for _, r := range rows {
			counts[r.ChannelID] = r.Count
		}

		assert.Equal(t, map[string]int64{channel1: 2, channel2: 1}, counts)
	})

	t.Run("top N by count", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter:   filter,
			InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionIssue}, Limit: 2},
			OrderByCountDesc: true,
		})

		require.Len(t, rows, 2)
		assert.Equal(t, "i1", rows[0].IssueID)
		assert.Equal(t, int64(4), rows[0].Count)
		assert.Equal(t, "i2", rows[1].IssueID)
		assert.Equal(t, int64(2), rows[1].Count)
	})

	t.Run("no matches returns no rows or a zero total", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter:   types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{uniqueChannelID()}},
			InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionType}},
		})
		assert.Empty(t, rows)
	})
}

// TestInsightsCountFilters verifies every filter field.
func TestInsightsCountFilters(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel := uniqueChannelID()

	matching := types.InsightsEvent{
		Time: base.Add(time.Minute), Type: types.InsightsEventIssueMoved, Reason: types.InsightsReasonUser, ChannelID: channel,
		RelatedChannelID: "C_TARGET", IssueID: "issue-x", CorrelationID: "corr-x", Severity: types.AlertWarning,
		RouteKey: "rk-x", ActorID: "U_X", Host: "h", Author: "a", AlertType: "t",
	}
	other := types.InsightsEvent{Time: base.Add(2 * time.Minute), Type: types.InsightsEventIssueCreated, ChannelID: channel, Severity: types.AlertError}

	require.NoError(t, sink.WriteInsightsEvents(ctx, []types.InsightsEvent{matching, other}))

	baseFilter := func() types.InsightsFilter {
		return types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel}}
	}

	tests := []struct {
		name   string
		modify func(f *types.InsightsFilter)
	}{
		{"types", func(f *types.InsightsFilter) { f.Types = []types.InsightsEventType{types.InsightsEventIssueMoved} }},
		{"reasons", func(f *types.InsightsFilter) { f.Reasons = []string{types.InsightsReasonUser} }},
		{"related channels", func(f *types.InsightsFilter) { f.RelatedChannelIDs = []string{"C_TARGET"} }},
		{"severities", func(f *types.InsightsFilter) { f.Severities = []types.AlertSeverity{types.AlertWarning} }},
		{"route keys", func(f *types.InsightsFilter) { f.RouteKeys = []string{"rk-x"} }},
		{"issue ID", func(f *types.InsightsFilter) { f.IssueID = "issue-x" }},
		{"correlation ID", func(f *types.InsightsFilter) { f.CorrelationID = "corr-x" }},
		{"actor ID", func(f *types.InsightsFilter) { f.ActorID = "U_X" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := baseFilter()
			tt.modify(&f)

			rows := countRows(t, reader, &types.InsightsCountQuery{InsightsFilter: f})
			require.Len(t, rows, 1)
			assert.Equal(t, int64(1), rows[0].Count)
		})
	}

	t.Run("group by attribute dimensions", func(t *testing.T) {
		f := baseFilter()
		f.Types = []types.InsightsEventType{types.InsightsEventIssueMoved}

		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter: f,
			InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{
				types.InsightsDimensionRelatedChannel, types.InsightsDimensionSeverity, types.InsightsDimensionRouteKey,
				types.InsightsDimensionAlertType, types.InsightsDimensionHost, types.InsightsDimensionAuthor,
				types.InsightsDimensionActor, types.InsightsDimensionCorrelationID,
			}},
		})

		require.Len(t, rows, 1)
		assert.Equal(t, types.InsightsGroupKey{
			RelatedChannelID: "C_TARGET", Severity: types.AlertWarning, RouteKey: "rk-x", AlertType: "t",
			Host: "h", Author: "a", ActorID: "U_X", CorrelationID: "corr-x",
		}, rows[0].InsightsGroupKey)
	})
}

// TestInsightsValueStats verifies numeric value aggregates.
func TestInsightsValueStats(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel := uniqueChannelID()

	events := make([]types.InsightsEvent, 0, 11)
	for i := 1; i <= 10; i++ {
		events = append(events, types.InsightsEvent{
			Time: base.Add(time.Duration(i) * time.Minute), Type: types.InsightsEventIssueArchived, ChannelID: channel,
			Reason: types.InsightsReasonAutoResolved,
			Values: map[string]float64{types.InsightsValueOpenSeconds: float64(i * 10)},
		})
	}
	// Event without the value is ignored.
	events = append(events, types.InsightsEvent{Time: base.Add(30 * time.Minute), Type: types.InsightsEventIssueArchived, ChannelID: channel, Reason: types.InsightsReasonTerminated})

	require.NoError(t, sink.WriteInsightsEvents(ctx, events))

	query := &types.InsightsStatsQuery{
		InsightsFilter: types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel}},
		ValueKey:       types.InsightsValueOpenSeconds,
		Aggregates: []types.InsightsAggregate{
			types.InsightsAggregateSum, types.InsightsAggregateAvg, types.InsightsAggregateMin, types.InsightsAggregateMax,
			types.InsightsAggregateP50, types.InsightsAggregateP90, types.InsightsAggregateP99,
		},
	}
	require.NoError(t, query.Validate())

	rows, err := reader.InsightsValueStats(ctx, query)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	r := rows[0]
	assert.Equal(t, int64(10), r.Count)
	assert.InDelta(t, 550, r.Values[types.InsightsAggregateSum], 0.001)
	assert.InDelta(t, 55, r.Values[types.InsightsAggregateAvg], 0.001)
	assert.InDelta(t, 10, r.Values[types.InsightsAggregateMin], 0.001)
	assert.InDelta(t, 100, r.Values[types.InsightsAggregateMax], 0.001)
	assert.InDelta(t, 55, r.Values[types.InsightsAggregateP50], 0.001)
	assert.InDelta(t, 91, r.Values[types.InsightsAggregateP90], 0.001)
	assert.InDelta(t, 99.1, r.Values[types.InsightsAggregateP99], 0.001)

	// Grouped by reason: the terminated event has no value and produces no row.
	query.GroupBy = []types.InsightsDimension{types.InsightsDimensionReason}
	query.Aggregates = []types.InsightsAggregate{types.InsightsAggregateMax}

	rows, err = reader.InsightsValueStats(ctx, query)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, types.InsightsReasonAutoResolved, rows[0].Reason)
	assert.InDelta(t, 100, rows[0].Values[types.InsightsAggregateMax], 0.001)
}

// TestInsightsTooManyRows verifies that count and stats queries without an explicit limit fail with
// [types.ErrInsightsTooManyRows] instead of returning a truncated result, and that an explicit limit truncates.
func TestInsightsTooManyRows(t *testing.T, sink types.InsightsSink, reader types.InsightsReader) {
	ctx := context.Background()
	base := insightsTestBase()
	channel := uniqueChannelID()

	events := make([]types.InsightsEvent, types.MaxInsightsRows+1)
	for i := range events {
		events[i] = types.InsightsEvent{
			Time: base.Add(time.Minute), Type: types.InsightsEventIssueArchived, ChannelID: channel,
			IssueID: fmt.Sprintf("issue-%05d", i), Values: map[string]float64{types.InsightsValueAlertCount: 1},
		}
	}

	// One event of another type, so that filtering on the archived type matches exactly the maximum.
	events[types.MaxInsightsRows].Type = types.InsightsEventIssueCreated

	require.NoError(t, sink.WriteInsightsEvents(ctx, events))

	filter := types.InsightsFilter{From: base, To: base.Add(time.Hour), ChannelIDs: []string{channel}}
	grouping := types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionIssue}}

	t.Run("count without limit", func(t *testing.T) {
		_, err := reader.CountInsightsEvents(ctx, &types.InsightsCountQuery{InsightsFilter: filter, InsightsGrouping: grouping})
		require.ErrorIs(t, err, types.ErrInsightsTooManyRows)
	})

	t.Run("stats without limit", func(t *testing.T) {
		_, err := reader.InsightsValueStats(ctx, &types.InsightsStatsQuery{
			InsightsFilter: filter, InsightsGrouping: grouping,
			ValueKey: types.InsightsValueAlertCount, Aggregates: []types.InsightsAggregate{types.InsightsAggregateSum},
		})
		require.ErrorIs(t, err, types.ErrInsightsTooManyRows)
	})

	t.Run("count with explicit limit truncates", func(t *testing.T) {
		limited := grouping
		limited.Limit = types.MaxInsightsRows

		rows := countRows(t, reader, &types.InsightsCountQuery{InsightsFilter: filter, InsightsGrouping: limited})
		assert.Len(t, rows, types.MaxInsightsRows)
	})

	t.Run("exactly the maximum without limit", func(t *testing.T) {
		rows := countRows(t, reader, &types.InsightsCountQuery{
			InsightsFilter: withTypes(filter, types.InsightsEventIssueArchived), InsightsGrouping: grouping,
		})
		assert.Len(t, rows, types.MaxInsightsRows)
	})
}

// TestInsightsLoadOpenIssues verifies loading open issues across channels.
func TestInsightsLoadOpenIssues(t *testing.T, db types.DB, reader types.InsightsReader) {
	ctx := context.Background()
	channel1 := uniqueChannelID()
	channel2 := uniqueChannelID()

	open1 := newTestIssue(newTestAlert(channel1, uuid.New().String()), "")
	open2 := newTestIssue(newTestAlert(channel2, uuid.New().String()), "")
	archived := newTestIssue(newTestAlert(channel1, uuid.New().String()), "")
	archived.Archived = true

	require.NoError(t, db.SaveIssues(ctx, open1, open2, archived))

	issues, err := reader.LoadOpenIssues(ctx, []string{channel1, channel2})
	require.NoError(t, err)
	assert.Len(t, issues, 2)
	assert.Contains(t, issues, open1.ID)
	assert.Contains(t, issues, open2.ID)

	issues, err = reader.LoadOpenIssues(ctx, []string{channel2})
	require.NoError(t, err)
	assert.Len(t, issues, 1)
	assert.Equal(t, open2.CorrelationID, testIssueFromJSON(issues[open2.ID]).CorrelationID)

	// Empty channel list means all channels.
	issues, err = reader.LoadOpenIssues(ctx, nil)
	require.NoError(t, err)
	assert.Contains(t, issues, open1.ID)
	assert.Contains(t, issues, open2.ID)
	assert.NotContains(t, issues, archived.ID)
}

func withTypes(f types.InsightsFilter, eventTypes ...types.InsightsEventType) types.InsightsFilter {
	f.Types = eventTypes
	return f
}

func countRows(t *testing.T, reader types.InsightsReader, query *types.InsightsCountQuery) []*types.InsightsCountRow {
	t.Helper()

	require.NoError(t, query.Validate())

	rows, err := reader.CountInsightsEvents(context.Background(), query)
	require.NoError(t, err)

	return rows
}
