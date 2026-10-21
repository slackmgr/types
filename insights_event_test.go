package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/slackmgr/types"
	"github.com/stretchr/testify/assert"
)

func TestInsightsEvent_Normalize(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("sets zero time to now", func(t *testing.T) {
		t.Parallel()

		e := types.InsightsEvent{Type: types.InsightsEventAlertGrouped}
		e.Normalize(now)
		assert.Equal(t, now, e.Time)
	})

	t.Run("converts time to UTC and keeps the instant", func(t *testing.T) {
		t.Parallel()

		local := time.Date(2026, 1, 2, 5, 4, 5, 0, time.FixedZone("CEST", 2*3600))
		e := types.InsightsEvent{Time: local}
		e.Normalize(now)
		assert.Equal(t, time.UTC, e.Time.Location())
		assert.True(t, local.Equal(e.Time))
	})

	t.Run("truncates long detail", func(t *testing.T) {
		t.Parallel()

		e := types.InsightsEvent{Detail: strings.Repeat("x", types.InsightsMaxDetailLen+50)}
		e.Normalize(now)
		assert.Len(t, e.Detail, types.InsightsMaxDetailLen)
	})

	t.Run("does not split multi-byte characters", func(t *testing.T) {
		t.Parallel()

		// "æ" is two bytes, so an odd byte limit would split it.
		e := types.InsightsEvent{Detail: "a" + strings.Repeat("æ", types.InsightsMaxDetailLen)}
		e.Normalize(now)
		assert.LessOrEqual(t, len(e.Detail), types.InsightsMaxDetailLen)
		assert.True(t, strings.HasPrefix(e.Detail, "aæ"))
		assert.False(t, strings.ContainsRune(e.Detail, '�'))
		assert.Len(t, e.Detail, types.InsightsMaxDetailLen-1)
	})

	t.Run("keeps short detail", func(t *testing.T) {
		t.Parallel()

		e := types.InsightsEvent{Detail: "short"}
		e.Normalize(now)
		assert.Equal(t, "short", e.Detail)
	})
}
