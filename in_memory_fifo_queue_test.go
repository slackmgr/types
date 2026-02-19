package common_test

import (
	"context"
	"testing"
	"time"

	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryFifoQueue(t *testing.T) {
	t.Parallel()

	t.Run("full queue should produce timeout error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		queue := common.NewInMemoryFifoQueue("alerts", 2, time.Millisecond)
		err := queue.Send(ctx, "C000000001", "dedupID_1", "body_1")
		require.NoError(t, err)
		err = queue.Send(ctx, "C000000002", "dedupID_2", "body_2")
		require.NoError(t, err)
		err = queue.Send(ctx, "C000000003", "dedupID_3", "body_3")
		require.ErrorContains(t, err, "timeout")
	})

	t.Run("cancelled context should return context error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		queue := common.NewInMemoryFifoQueue("alerts", 1, time.Second)
		err := queue.Send(ctx, "C000000001", "dedupID_1", "body_1")
		require.NoError(t, err)
		cancel()
		err = queue.Send(ctx, "C000000002", "dedupID_2", "body_2")
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("receive function should return all items in order", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		queue := common.NewInMemoryFifoQueue("alerts", 3, time.Millisecond)
		err := queue.Send(ctx, "C000000001", "dedupID_1", "body_1")
		require.NoError(t, err)
		err = queue.Send(ctx, "C000000002", "dedupID_2", "body_2")
		require.NoError(t, err)
		err = queue.Send(ctx, "C000000003", "dedupID_3", "body_3")
		require.NoError(t, err)

		receivedItems := make(chan *common.FifoQueueItem, 3)

		go func() {
			err := queue.Receive(ctx, receivedItems)
			assert.ErrorIs(t, err, context.Canceled)
		}()

		result := []*common.FifoQueueItem{}

		for item := range receivedItems {
			result = append(result, item)
			if len(result) == 3 {
				cancel()
			}
		}

		assert.Len(t, result, 3)
		assert.Equal(t, "body_1", result[0].Body)
		assert.Equal(t, "body_2", result[1].Body)
		assert.Equal(t, "body_3", result[2].Body)
	})

	t.Run("receive function should react to context cancelled when waiting to write to sinkCh", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		queue := common.NewInMemoryFifoQueue("alerts", 2, time.Second)
		err := queue.Send(ctx, "C000000001", "dedupID_1", "body_1")
		require.NoError(t, err)
		err = queue.Send(ctx, "C000000002", "dedupID_2", "body_2")
		require.NoError(t, err)

		receivedItems := make(chan *common.FifoQueueItem)

		go func() {
			err := queue.Receive(ctx, receivedItems)
			assert.ErrorIs(t, err, context.Canceled)
		}()

		for range receivedItems {
			cancel()
			break
		}
	})
}
