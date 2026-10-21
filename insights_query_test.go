package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validFilter() types.InsightsFilter {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return types.InsightsFilter{From: from, To: from.Add(24 * time.Hour)}
}

func TestInsightsFilter_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		modify  func(f *types.InsightsFilter)
		wantErr string
	}{
		{"valid", func(_ *types.InsightsFilter) {}, ""},
		{"missing from", func(f *types.InsightsFilter) { f.From = time.Time{} }, "from and to are required"},
		{"missing to", func(f *types.InsightsFilter) { f.To = time.Time{} }, "from and to are required"},
		{"to equal to from", func(f *types.InsightsFilter) { f.To = f.From }, "to must be after from"},
		{"range too long", func(f *types.InsightsFilter) { f.To = f.From.Add(types.MaxInsightsQueryRange + time.Hour) }, "time range must not exceed"},
		{"too many channels", func(f *types.InsightsFilter) { f.ChannelIDs = make([]string, types.MaxInsightsFilterValues+1) }, "too many channel IDs"},
		{"too many types", func(f *types.InsightsFilter) {
			f.Types = make([]types.InsightsEventType, types.MaxInsightsFilterValues+1)
		}, "too many types"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := validFilter()
			tt.modify(&f)

			err := f.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestInsightsCountQuery_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   types.InsightsCountQuery
		wantErr string
	}{
		{"valid total", types.InsightsCountQuery{}, ""},
		{"valid bucket", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionBucket}, Bucket: time.Hour}}, ""},
		{"valid distinct", types.InsightsCountQuery{Distinct: types.InsightsDimensionIssue}, ""},
		{"invalid dimension", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{"bogus"}}}, "invalid dimension"},
		{"duplicate dimension", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionType, types.InsightsDimensionType}}}, "duplicate dimension"},
		{"bucket without size", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionBucket}}}, "bucket size must be at least"},
		{"minute buckets over a day", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionBucket}, Bucket: time.Minute, Limit: 0}}, ""},
		{"size without bucket dimension", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{Bucket: time.Hour}}, "only allowed when grouping by bucket"},
		{"negative limit", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{Limit: -1}}, "limit must be between"},
		{"limit too high", types.InsightsCountQuery{InsightsGrouping: types.InsightsGrouping{Limit: types.MaxInsightsRows + 1}}, "limit must be between"},
		{"distinct bucket", types.InsightsCountQuery{Distinct: types.InsightsDimensionBucket}, "invalid distinct dimension"},
		{"distinct invalid", types.InsightsCountQuery{Distinct: "bogus"}, "invalid distinct dimension"},
		{"invalid filter", types.InsightsCountQuery{InsightsFilter: types.InsightsFilter{}}, "from and to are required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := tt.query
			if tt.name != "invalid filter" {
				q.InsightsFilter = validFilter()
			}

			err := q.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}

	t.Run("too many buckets over a long range", func(t *testing.T) {
		t.Parallel()

		q := types.InsightsCountQuery{InsightsFilter: validFilter()}
		q.To = q.From.Add(30 * 24 * time.Hour)
		q.GroupBy = []types.InsightsDimension{types.InsightsDimensionBucket}
		q.Bucket = time.Minute
		require.ErrorContains(t, q.Validate(), "too many buckets")
	})
}

func TestInsightsStatsQuery_Validate(t *testing.T) {
	t.Parallel()

	valid := func() types.InsightsStatsQuery {
		return types.InsightsStatsQuery{
			InsightsFilter: validFilter(),
			ValueKey:       types.InsightsValueOpenSeconds,
			Aggregates:     []types.InsightsAggregate{types.InsightsAggregateP50},
		}
	}

	q := valid()
	require.NoError(t, q.Validate())

	q = valid()
	q.ValueKey = "Bad-Key"
	require.ErrorContains(t, q.Validate(), "invalid value key")

	q = valid()
	q.ValueKey = strings.Repeat("a", 65)
	require.ErrorContains(t, q.Validate(), "invalid value key")

	q = valid()
	q.Aggregates = nil
	require.ErrorContains(t, q.Validate(), "at least one aggregate")

	q = valid()
	q.Aggregates = []types.InsightsAggregate{"median"}
	require.ErrorContains(t, q.Validate(), "invalid aggregate")

	q = valid()
	q.GroupBy = []types.InsightsDimension{"bogus"}
	require.ErrorContains(t, q.Validate(), "invalid dimension")

	q = valid()
	q.From = time.Time{}
	require.ErrorContains(t, q.Validate(), "from and to are required")
}

func TestInsightsListQuery_Validate(t *testing.T) {
	t.Parallel()

	q := types.InsightsListQuery{InsightsFilter: validFilter()}
	require.NoError(t, q.Validate())
	assert.Equal(t, types.DefaultInsightsPageSize, q.EffectiveLimit())

	q.Limit = 10
	assert.Equal(t, 10, q.EffectiveLimit())

	q.Limit = types.MaxInsightsPageSize + 1
	require.ErrorContains(t, q.Validate(), "limit must be between")

	q.Limit = -1
	require.ErrorContains(t, q.Validate(), "limit must be between")

	q = types.InsightsListQuery{}
	require.ErrorContains(t, q.Validate(), "from and to are required")
}

func TestInsightsGrouping_Helpers(t *testing.T) {
	t.Parallel()

	g := types.InsightsGrouping{GroupBy: []types.InsightsDimension{types.InsightsDimensionChannel}}
	assert.True(t, g.HasDimension(types.InsightsDimensionChannel))
	assert.False(t, g.HasDimension(types.InsightsDimensionType))
	assert.Equal(t, types.MaxInsightsRows, g.EffectiveLimit())

	g.Limit = 5
	assert.Equal(t, 5, g.EffectiveLimit())
}

func TestInsightsDimensionAndAggregate_IsValid(t *testing.T) {
	t.Parallel()

	for _, d := range []types.InsightsDimension{
		types.InsightsDimensionBucket, types.InsightsDimensionChannel, types.InsightsDimensionRelatedChannel,
		types.InsightsDimensionType, types.InsightsDimensionReason, types.InsightsDimensionSeverity,
		types.InsightsDimensionRouteKey, types.InsightsDimensionAlertType, types.InsightsDimensionHost,
		types.InsightsDimensionAuthor, types.InsightsDimensionActor, types.InsightsDimensionCorrelationID,
		types.InsightsDimensionIssue,
	} {
		assert.True(t, d.IsValid(), d)
	}

	assert.False(t, types.InsightsDimension("bogus").IsValid())

	for _, a := range []types.InsightsAggregate{
		types.InsightsAggregateSum, types.InsightsAggregateAvg, types.InsightsAggregateMin, types.InsightsAggregateMax,
		types.InsightsAggregateP50, types.InsightsAggregateP90, types.InsightsAggregateP99,
	} {
		assert.True(t, a.IsValid(), a)
	}

	assert.False(t, types.InsightsAggregate("median").IsValid())
}
