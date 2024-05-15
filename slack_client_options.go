package common

import "time"

type SlackClientOption func(*SlackClientOptions)

type SlackClientOptions struct {
	DebugLogging                 bool
	DryRun                       bool
	Concurrency                  int
	MaxAttemtsForRateLimitError  int
	MaxAttemptsForTransientError int
	MaxAttemptsForFatalError     int
	MaxRateLimitErrorWaitTime    time.Duration
	MaxTransientErrorWaitTime    time.Duration
	MaxFatalErrorWaitTime        time.Duration
	HTTPTimeout                  time.Duration
}

func NewSlackOptions() *SlackClientOptions {
	return &SlackClientOptions{
		DebugLogging:                 false,
		DryRun:                       false,
		Concurrency:                  10,
		MaxAttemtsForRateLimitError:  10,
		MaxAttemptsForTransientError: 5,
		MaxAttemptsForFatalError:     5,
		MaxRateLimitErrorWaitTime:    2 * time.Minute,
		MaxTransientErrorWaitTime:    30 * time.Second,
		MaxFatalErrorWaitTime:        30 * time.Second,
		HTTPTimeout:                  30 * time.Second,
	}
}

func WithDebugLogging(debug bool) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.DebugLogging = debug
	}
}

func WithDryRun(dryRun bool) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.DryRun = dryRun
	}
}

func WithConcurrency(concurrency int) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.Concurrency = concurrency
	}
}

func WithMaxAttemptsForRateLimitError(maxAttempts int) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxAttemtsForRateLimitError = maxAttempts
	}
}

func WithMaxAttemptsForTransientError(maxAttempts int) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxAttemptsForTransientError = maxAttempts
	}
}

func WithMaxAttemptsForFatalError(maxAttempts int) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxAttemptsForFatalError = maxAttempts
	}
}

func WithMaxRateLimitErrorWaitTime(waitTime time.Duration) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxRateLimitErrorWaitTime = waitTime
	}
}

func WithMaxTransientErrorWaitTime(waitTime time.Duration) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxTransientErrorWaitTime = waitTime
	}
}

func WithMaxFatalErrorWaitTime(waitTime time.Duration) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.MaxFatalErrorWaitTime = waitTime
	}
}

func WithHTTPTimeout(timeout time.Duration) SlackClientOption {
	return func(o *SlackClientOptions) {
		o.HTTPTimeout = timeout
	}
}
