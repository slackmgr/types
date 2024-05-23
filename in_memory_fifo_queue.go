package common

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type InMemoryFifoQueue struct {
	items        chan *FifoQueueItem
	writeTimeout time.Duration
}

func NewInMemoryFifoQueue(bufferSize int, writeTimeout time.Duration) *InMemoryFifoQueue {
	return &InMemoryFifoQueue{
		items:        make(chan *FifoQueueItem, bufferSize),
		writeTimeout: writeTimeout,
	}
}

func (q *InMemoryFifoQueue) Send(ctx context.Context, groupID, dedupID, body string) error {
	item := &FifoQueueItem{
		MessageID:         uuid.New().String(),
		GroupID:           groupID,
		ReceiveTimestamp:  time.Now(),
		VisibilityTimeout: 0,
		Body:              body,
		Ack: func(_ context.Context) error {
			return nil
		},
		ExtendVisibility: nil, // Message visibility extension is not supported in this implementation
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(q.writeTimeout):
		return fmt.Errorf("timeout while writing to queue")
	case q.items <- item:
		return nil
	}
}

func (q *InMemoryFifoQueue) Receive(ctx context.Context, sinkCh chan<- *FifoQueueItem) error {
	defer close(sinkCh)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item := <-q.items:
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sinkCh <- item:
			}
		}
	}
}
