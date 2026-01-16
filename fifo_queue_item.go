package common

import (
	"time"
)

// FifoQueueItem represents an item received from a FIFO queue.
type FifoQueueItem struct {
	// MessageID is the unique identifier of the message (as defined by the queue implementation).
	MessageID string

	// SlackChannelID is the ID of the Slack channel to which the message is related.
	SlackChannelID string

	// ReceiveTimestamp is the time when the message was received from the queue.
	ReceiveTimestamp time.Time

	// Body is the body of the message.
	Body string

	// Ack acknowledges the successful processing of the message, effectively removing it from the queue.
	// This function cannot be nil.
	//
	// Ack does not accept a context parameter because acknowledgment is a commitment that must complete
	// regardless of the caller's context state. Each queue implementation is responsible for managing
	// its own timeouts and retry logic internally.
	Ack func()

	// Nack negatively acknowledges the processing of the message, thus making it available for reprocessing.
	// This function cannot be nil.
	//
	// Nack does not accept a context parameter because negative acknowledgment is a commitment that must
	// complete regardless of the caller's context state. Each queue implementation is responsible for
	// managing its own timeouts and retry logic internally.
	Nack func()
}
