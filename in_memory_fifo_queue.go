package common

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// InMemoryFifoQueue is an in-memory FIFO queue implementation
// For TEST purposes only! Do not use in production!
type InMemoryFifoQueue struct {
	name         string
	items        chan *FifoQueueItem
	writeTimeout time.Duration
}

// NewInMemoryFifoQueue creates a new InMemoryFifoQueue instance.
// name is the name of the queue (for logging purposes only).
// bufferSize is the maximum number of items that can be stored in the queue.
// writeTimeout is the maximum time to wait for writing an item to the queue.
//
// For TEST purposes only! Do not use in production!
func NewInMemoryFifoQueue(name string, bufferSize int, writeTimeout time.Duration) *InMemoryFifoQueue {
	return &InMemoryFifoQueue{
		name:         name,
		items:        make(chan *FifoQueueItem, bufferSize),
		writeTimeout: writeTimeout,
	}
}

// Name returns the name of the queue.
func (q *InMemoryFifoQueue) Name() string {
	return q.name
}

// Send sends a message to the queue.
// An error is returned if the context is canceled or the write timeout is reached.
func (q *InMemoryFifoQueue) Send(ctx context.Context, slackChannelID, _, body string) error {
	item := &FifoQueueItem{
		MessageID:        uuid.New().String(),
		SlackChannelID:   slackChannelID,
		ReceiveTimestamp: time.Now(),
		Body:             body,
		Ack:              func() {},
		Nack:             func() {},
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(q.writeTimeout):
		return errors.New("timeout while writing to queue")
	case q.items <- item:
		return nil
	}
}

// Receive receives messages from the queue, to the specified sink channel.
// An error is returned if the context is canceled.
// The sink channel is closed when the function returns.
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
