# slack-manager-common

[![Go Reference](https://pkg.go.dev/badge/github.com/slackmgr/slack-manager-common.svg)](https://pkg.go.dev/github.com/slackmgr/slack-manager-common)
[![Go Report Card](https://goreportcard.com/badge/github.com/slackmgr/slack-manager-common)](https://goreportcard.com/report/github.com/slackmgr/slack-manager-common)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/slackmgr/slack-manager-common/workflows/CI/badge.svg)](https://github.com/slackmgr/slack-manager-common/actions)

A Go shared library package providing common interfaces and data structures for the Slack Manager system. This package defines contracts for database access, logging, metrics, and core domain types like alerts and issues.

## Important: When to Use This Library

**Most users don't need to import this library directly.** When using the Slack Manager, this package is automatically pulled in as a dependency by other components in the ecosystem:

- [Slack Manager](https://github.com/slackmgr/slack-manager) (main application)
- [Go Client](https://github.com/slackmgr/slack-manager-go-client) (for sending alerts)
- Database plugins ([DynamoDB](https://github.com/slackmgr/slack-manager-dynamodb-plugin), [PostgreSQL](https://github.com/slackmgr/slack-manager-postgres-plugin))
- Messaging plugins ([SQS](https://github.com/slackmgr/slack-manager-sqs-plugin), [PubSub](https://github.com/slackmgr/slack-manager-pubsub-plugin))

**You only need to import this library directly if:**
- You're developing a **custom database plugin** (implementing the `DB` interface)
- You're developing a **custom queue/messaging plugin**
- You're building custom tooling that needs access to the core types

For sending alerts to Slack Manager, use the [Go Client](https://github.com/slackmgr/slack-manager-go-client) instead.

## Overview

The `slack-manager-common` package serves as the foundation for the Slack Manager ecosystem, providing:

- **Dependency Injection Interfaces**: Abstractions for database, logging, and metrics that allow implementations to be plugged in
- **Core Domain Types**: Rich type definitions for alerts, issues, webhooks, and escalations
- **Validation & Cleaning**: Built-in validation and normalization for all data structures
- **Type Safety**: Strongly-typed enums for severities, access levels, button styles, and display modes

This library is used by the main [Slack Manager](https://github.com/slackmgr/slack-manager) application, database plugins ([DynamoDB](https://github.com/slackmgr/slack-manager-dynamodb-plugin), [PostgreSQL](https://github.com/slackmgr/slack-manager-postgres-plugin)), messaging plugins ([SQS](https://github.com/slackmgr/slack-manager-sqs-plugin), [PubSub](https://github.com/slackmgr/slack-manager-pubsub-plugin)), and the [Go client](https://github.com/slackmgr/slack-manager-go-client).

## Installation

```bash
go get github.com/slackmgr/slack-manager-common
```

## Core Interfaces

These interfaces define the contracts that implementations must satisfy. They enable dependency injection and allow the Slack Manager to work with different storage backends, logging frameworks, and metrics systems.

### DB Interface

The `DB` interface abstracts all database operations for the Slack Manager. Implementations must handle persistence for alerts, issues, move mappings, and channel processing state.

```go
type DB interface {
    Init(ctx context.Context, skipSchemaValidation bool) error
    SaveAlert(ctx context.Context, alert *Alert) error
    SaveIssue(ctx context.Context, issue Issue) error
    SaveIssues(ctx context.Context, issues ...Issue) error
    MoveIssue(ctx context.Context, issue Issue, sourceChannelID, targetChannelID string) error
    FindOpenIssueByCorrelationID(ctx context.Context, channelID, correlationID string) (string, json.RawMessage, error)
    FindIssueBySlackPostID(ctx context.Context, channelID, postID string) (string, json.RawMessage, error)
    FindActiveChannels(ctx context.Context) ([]string, error)
    LoadOpenIssuesInChannel(ctx context.Context, channelID string) (map[string]json.RawMessage, error)
    SaveMoveMapping(ctx context.Context, moveMapping MoveMapping) error
    FindMoveMapping(ctx context.Context, channelID, correlationID string) (json.RawMessage, error)
    DeleteMoveMapping(ctx context.Context, channelID, correlationID string) error
    SaveChannelProcessingState(ctx context.Context, state *ChannelProcessingState) error
    FindChannelProcessingState(ctx context.Context, channelID string) (*ChannelProcessingState, error)
    DropAllData(ctx context.Context) error
}
```

**Key Points:**
- Issues and move mappings are stored as opaque JSON (`json.RawMessage`) to allow implementation flexibility
- Database implementations should never depend on the internal structure of issues or move mappings
- Implementations available: DynamoDB plugin, PostgreSQL plugin

### Logger Interface

The `Logger` interface provides structured logging with field support and multiple log levels.

```go
type Logger interface {
    Debug(msg string)
    Debugf(format string, args ...any)
    Info(msg string)
    Infof(format string, args ...any)
    Error(msg string)
    Errorf(format string, args ...any)
    WithField(key string, value any) Logger
    WithFields(fields map[string]any) Logger
}
```

**Key Points:**
- Supports Debug, Info, and Error levels
- Allows chaining with `WithField` and `WithFields` for structured logging
- A no-op implementation (`NoopLogger`) is provided for testing

### Metrics Interface

The `Metrics` interface provides Prometheus-style metrics with support for counters, gauges, and histograms.

```go
type Metrics interface {
    RegisterCounter(name, help string, labels ...string)
    RegisterGauge(name, help string, labels ...string)
    RegisterHistogram(name, help string, buckets []float64, labels ...string)
    Add(name string, value float64, labelValues ...string)
    Inc(name string, labelValues ...string)
    Set(name string, value float64, labelValues ...string)
    Observe(name string, value float64, labelValues ...string)
}
```

**Key Points:**
- Supports standard Prometheus metric types
- Labels can be defined at registration and specified at observation time
- A no-op implementation (`NoopMetrics`) is provided for testing

## Core Domain Types

### Alert

The `Alert` struct is the central type representing an alert sent to the Slack Manager. It contains comprehensive information about the alert including severity, content, routing, escalations, and webhooks.

**Constructor Functions:**
```go
func NewPanicAlert() *Alert     // Severity: panic
func NewErrorAlert() *Alert     // Severity: error
func NewWarningAlert() *Alert   // Severity: warning
func NewResolvedAlert() *Alert  // Severity: resolved
func NewInfoAlert() *Alert      // Severity: info
```

**Key Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Timestamp` | `time.Time` | When the alert was created (auto-replaced if > 7 days old) |
| `CorrelationID` | `string` | Groups related alerts into issues (auto-generated if not set) |
| `Severity` | `AlertSeverity` | Alert severity: panic, error, warning, resolved, or info |
| `Header` | `string` | Alert title (max 130 chars, auto-truncated) |
| `Text` | `string` | Alert body (max 10,000 chars, auto-truncated) |
| `SlackChannelID` | `string` | Target Slack channel ID or name |
| `RouteKey` | `string` | Alternative routing via configured routes |
| `IssueFollowUpEnabled` | `bool` | Whether to track this alert as an issue |
| `AutoResolveSeconds` | `int` | Auto-resolve after N seconds (30 - 63,113,851) |
| `Webhooks` | `[]*Webhook` | Interactive buttons (max 5) |
| `Escalation` | `[]*Escalation` | Escalation points (max 3) |
| `Fields` | `[]*Field` | Additional key-value fields (max 20) |

**Methods:**
- `Clean()`: Normalizes and truncates all fields to valid values
- `Validate()`: Returns error if any field is invalid
- `UniqueID()`: Returns a deterministic, base64-encoded unique ID

**Validation:**
- The package defines extensive constants for maximum lengths (e.g., `MaxHeaderLength = 130`)
- All validation methods return descriptive errors
- Validation includes: channel IDs, URLs, emoji format, severity values, escalation timing

**Special Features:**
- **Status Emoji Replacement**: Use `:status:` in header or text, and it will be replaced with the appropriate emoji based on severity
- **Conditional Content**: `HeaderWhenResolved` and `TextWhenResolved` allow different content for resolved states
- **Auto-correlation**: If no `CorrelationID` is provided, one is generated by hashing key fields
- **Ignore Patterns**: `IgnoreIfTextContains` allows filtering out known noise

### AlertSeverity

Alert severity levels with associated emojis in Slack:

```go
const (
    AlertPanic    AlertSeverity = "panic"    // Panic icon
    AlertError    AlertSeverity = "error"    // Red error icon
    AlertWarning  AlertSeverity = "warning"  // Yellow warning icon
    AlertResolved AlertSeverity = "resolved" // Green OK icon
    AlertInfo     AlertSeverity = "info"     // Blue info icon
)
```

**Helper Functions:**
- `SeverityIsValid(s AlertSeverity) bool`
- `SeverityPriority(s AlertSeverity) int` (3 = panic, 2 = error, 1 = warning, 0 = resolved/info)
- `ValidSeverities() []string`

### Escalation

Defines escalation points that trigger when an issue remains unresolved.

```go
type Escalation struct {
    Severity      AlertSeverity  // New severity when escalation triggers
    DelaySeconds  int            // Delay since issue creation (min 30s)
    SlackMentions []string       // Mentions to add (e.g., "<!here>", "<@U12345678>")
    MoveToChannel string         // Move issue to different channel
}
```

**Key Points:**
- Escalations are sorted by `DelaySeconds` and triggered in order
- Minimum delay: 30 seconds, minimum diff between escalations: 30 seconds
- Severity can only be panic, error, or warning (not resolved or info)
- Maximum 3 escalation points per alert

### Webhook

Interactive buttons that appear on Slack posts. When clicked, they trigger HTTP POST requests or custom handlers.

```go
type Webhook struct {
    ID               string                    // Unique within alert
    URL              string                    // HTTP URL or handler identifier
    ButtonText       string                    // Button label (max 25 chars)
    ButtonStyle      WebhookButtonStyle        // "primary" or "danger"
    AccessLevel      WebhookAccessLevel        // Who can click: global_admins, channel_admins, channel_members
    DisplayMode      WebhookDisplayMode        // When to show: always, open_issue, resolved_issue
    ConfirmationText string                    // Optional confirmation dialog text
    Payload          map[string]any            // Data sent in POST body
    PlainTextInput   []*WebhookPlainTextInput  // Text input fields
    CheckboxInput    []*WebhookCheckboxInput   // Checkbox groups
}
```

**Related Types:**
- `WebhookPlainTextInput`: Text input with min/max length, multiline support, initial value
- `WebhookCheckboxInput`: Checkbox group with label and multiple options
- `WebhookCheckboxOption`: Individual checkbox with value, text, and selected state

**Enums:**
- `WebhookButtonStyle`: `primary`, `danger`
- `WebhookAccessLevel`: `global_admins`, `channel_admins`, `channel_members`
- `WebhookDisplayMode`: `always`, `open_issue`, `resolved_issue`

### WebhookCallback

Represents data received when a webhook is triggered by a user clicking a button.

```go
type WebhookCallback struct {
    ID            string              // Webhook ID
    UserID        string              // Slack user ID who clicked
    UserRealName  string              // User's display name
    ChannelID     string              // Channel where button was clicked
    MessageID     string              // Slack message ID
    Timestamp     time.Time           // When button was clicked
    Input         map[string]string   // Text input values
    CheckboxInput map[string][]string // Checkbox selected values
    Payload       map[string]any      // Original webhook payload + metadata
}
```

**Helper Methods:**
- `GetPayloadValue(key string) any`
- `GetPayloadString(key string) string`
- `GetPayloadInt(key string, defaultValue int) int`
- `GetPayloadBool(key string, defaultValue bool) bool`
- `GetInputValue(key string) string`
- `GetCheckboxInputSelectedValues(key string) []string`

### Issue

The `Issue` interface represents an issue in a Slack channel. Issues group related alerts together and track their resolution status.

```go
type Issue interface {
    json.Marshaler
    ChannelID() string        // Slack channel ID
    UniqueID() string         // Base64-encoded unique ID for storage
    GetCorrelationID() string // Correlation ID for grouping alerts
    IsOpen() bool             // Whether issue is open (not archived)
    CurrentPostID() string    // Slack message ID (empty if no post yet)
}
```

**Key Points:**
- The actual implementation is internal to the Slack Manager and may change
- Database implementations must store issues as opaque JSON
- Correlation IDs are not guaranteed to be unique and should not be used as database keys

### MoveMapping

The `MoveMapping` interface tracks issues that have been moved from one channel to another.

```go
type MoveMapping interface {
    json.Marshaler
    ChannelID() string        // Original channel ID
    UniqueID() string         // Base64-encoded unique ID for storage
    GetCorrelationID() string // Correlation ID that was moved
}
```

**Key Points:**
- Ensures new alerts with the same correlation ID go to the new channel
- Stored as opaque JSON in the database
- Internal implementation may change without notice

### ChannelProcessingState

Tracks per-channel processing state to prevent concurrent processing and ensure regular intervals.

```go
type ChannelProcessingState struct {
    ChannelID           string    // Slack channel ID
    Created             time.Time // When processing state was created
    LastChannelActivity time.Time // Last activity in channel
    LastProcessed       time.Time // When channel was last processed
    OpenIssues          int       // Number of open issues
}
```

**Constructor:**
```go
func NewChannelProcessingState(channelID string) *ChannelProcessingState
```

## Field Specifications

The `Field` struct is used for additional key-value pairs displayed in Slack:

```go
type Field struct {
    Title string // Max 30 chars, auto-truncated
    Value string // Max 200 chars, auto-truncated
}
```

Maximum 20 fields per alert.

## Testing Utilities

### Database Testing

The `dbtests` package provides a shared test suite that can be run against any `DB` implementation:

```go
import "github.com/slackmgr/slack-manager-common/dbtests"

// In your database implementation tests:
func TestDatabaseCompliance(t *testing.T) {
    db := NewYourDatabase()
    dbtests.RunAllTests(t, db)
}
```

This ensures your database implementation correctly satisfies the `DB` interface contract.

### No-op Implementations

For testing purposes, no-op implementations are provided:

- `NoopLogger`: Logger that does nothing
- `NoopMetrics`: Metrics that do nothing
- `InMemoryFifoQueue`: Simple in-memory FIFO queue (test-only, not for production)

## Usage Example

```go
package main

import (
    "github.com/slackmgr/slack-manager-common"
)

func main() {
    // Create an alert
    alert := common.NewErrorAlert()
    alert.Header = "Database Connection Failed"
    alert.Text = "Unable to connect to production database. Error: connection timeout"
    alert.SlackChannelID = "C12345678"
    alert.IssueFollowUpEnabled = true
    alert.AutoResolveSeconds = 300 // Auto-resolve after 5 minutes

    // Add fields
    alert.Fields = []*common.Field{
        {Title: "Host", Value: "db-prod-01"},
        {Title: "Port", Value: "5432"},
    }

    // Add escalation
    alert.Escalation = []*common.Escalation{
        {
            Severity:      common.AlertPanic,
            DelaySeconds:  300,
            SlackMentions: []string{"<!here>"},
        },
    }

    // Add webhook button
    alert.Webhooks = []*common.Webhook{
        {
            ID:          "restart",
            URL:         "https://example.com/webhook/restart",
            ButtonText:  "Restart DB",
            ButtonStyle: common.WebhookButtonStyleDanger,
            AccessLevel: common.WebhookAccessLevelChannelAdmins,
            ConfirmationText: "Are you sure you want to restart the database?",
            Payload: map[string]any{
                "action": "restart_database",
                "host":   "db-prod-01",
            },
        },
    }

    // Clean and validate
    alert.Clean()
    if err := alert.Validate(); err != nil {
        panic(err)
    }

    // Use with Slack Manager API or directly with the manager
}
```

## Validation Constants

The package defines extensive validation constants. Key limits include:

| Constant | Value | Description |
|----------|-------|-------------|
| `MaxHeaderLength` | 130 | Alert header length |
| `MaxTextLength` | 10,000 | Alert text length |
| `MaxFieldCount` | 20 | Fields per alert |
| `MaxWebhookCount` | 5 | Webhooks per alert |
| `MaxEscalationCount` | 3 | Escalation points per alert |
| `MinAutoResolveSeconds` | 30 | Minimum auto-resolve time |
| `MaxAutoResolveSeconds` | 63,113,851 | Maximum auto-resolve time (~2 years) |
| `MinEscalationDelaySeconds` | 30 | Minimum first escalation delay |
| `MinEscalationDelayDiffSeconds` | 30 | Minimum time between escalations |

See `alert.go` for the complete list of constants.

## Build Commands

```bash
make init    # Download and clean dependencies (go mod tidy)
make test    # Run full test suite (gosec, go fmt, go test, go vet)
make lint    # Run golangci-lint
```

Run a specific test:
```bash
go test -v -run TestFunctionName ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Peter Aglen
