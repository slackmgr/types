package common

import (
	"time"
)

// ChannelProcessingState represents the state of processing for a specific Slack channel, used internally by the Slack Manager.
// It tracks when the processing started and when it was last processed.
// This is used to prevent multiple instances of the Slack Manager from processing the same channel at the same time,
// and to ensure that processing is done at regular intervals.
type ChannelProcessingState struct {
	ChannelID     string    `json:"channelId"`
	Created       time.Time `json:"created"`
	LastProcessed time.Time `json:"lastProcessed"`
}

func NewChannelProcessingState(channelID string) *ChannelProcessingState {
	return &ChannelProcessingState{
		ChannelID:     channelID,
		Created:       time.Now(),
		LastProcessed: time.Time{},
	}
}
