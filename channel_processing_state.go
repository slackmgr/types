package types

import (
	"time"
)

// ChannelProcessingState represents the state of processing for a specific Slack channel, used internally by the Slack Manager.
// It tracks when the processing started and when it was last processed.
// This is used to prevent multiple instances of the Slack Manager from processing the same channel at the same time,
// and to ensure that processing is done at regular intervals.
type ChannelProcessingState struct {
	ChannelID           string    `json:"channelId"`
	Created             time.Time `json:"created"`
	LastChannelActivity time.Time `json:"lastChannelActivity"`
	LastProcessed       time.Time `json:"lastProcessed"`
	// OpenIssues is the number of open issues in the channel at the last processing run, before archiving.
	// Fire-and-forget notifications are not issues, and are not counted.
	OpenIssues int `json:"openIssues"`
}

func NewChannelProcessingState(channelID string) *ChannelProcessingState {
	return &ChannelProcessingState{
		ChannelID:           channelID,
		Created:             time.Now().UTC(),
		LastChannelActivity: time.Now().UTC(),
		LastProcessed:       time.Time{}.UTC(),
	}
}
