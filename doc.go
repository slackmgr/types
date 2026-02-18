// Package common provides shared interfaces and data structures for the Slack Manager system.
//
// This package defines contracts for database access, logging, metrics, and core domain types
// like alerts and issues. It serves as the foundation for the Slack Manager ecosystem, enabling
// dependency injection and allowing different implementations to be plugged in.
//
// # Core Interfaces
//
// DB - Database abstraction for persisting alerts, issues, move mappings, and channel processing state.
// Implementations must handle storage as opaque JSON to allow flexibility.
//
// Logger - Structured logging interface with Debug/Info/Error levels and field support.
// Supports method chaining with WithField and WithFields.
//
// Metrics - Prometheus-style metrics interface supporting counters, gauges, and histograms.
// Allows registration of metrics with labels and observation of values.
//
// # Core Domain Types
//
// Alert - The central type representing an alert with comprehensive validation and cleaning.
// Contains severity, header, text, fields, webhooks, escalations, and routing information.
// Use constructor functions: NewPanicAlert(), NewErrorAlert(), NewWarningAlert(), NewResolvedAlert(), NewInfoAlert().
//
// Issue - Interface for tracking issue state in channels. Issues group related alerts together
// using correlation IDs. The actual implementation is internal and stored as opaque JSON.
//
// MoveMapping - Interface for tracking issues that have been moved between channels.
// Ensures new alerts with the same correlation ID go to the new channel.
//
// ChannelProcessingState - Tracks per-channel processing timestamps and open issue counts
// to prevent concurrent processing and ensure regular intervals.
//
// # Webhook Support
//
// The package provides rich webhook support with interactive buttons, input forms, and access control:
//
//   - Webhook - Interactive button configuration with HTTP callbacks or custom handlers
//   - WebhookCallback - Data received when a webhook is triggered
//   - WebhookAccessLevel - Control who can click buttons (global_admins, channel_admins, channel_members)
//   - WebhookButtonStyle - Visual style (primary, danger)
//   - WebhookDisplayMode - When to show buttons (always, open_issue, resolved_issue)
//
// # Validation and Cleaning
//
// Alert provides extensive validation and cleaning methods:
//
//   - Clean() - Normalizes and truncates all fields to valid values
//   - Validate() - Returns error if any field is invalid
//   - Individual validation methods for specific fields (ValidateSlackChannelIDAndRouteKey, etc.)
//
// The package defines comprehensive constants for maximum lengths and limits (e.g., MaxHeaderLength = 130).
//
// # Testing Utilities
//
// The dbtests subpackage provides a shared test suite that can be run against any DB implementation
// to ensure compliance with the interface contract.
//
// No-op implementations (NoopLogger, NoopMetrics) are provided for testing purposes.
// InMemoryFifoQueue is provided for testing but should not be used in production.
//
// # Usage Example
//
//	// Initialize a new error-level notification
//	alert := common.NewErrorAlert()
//	alert.Header = "Database Connection Failed"
//	alert.Text = "Unable to connect to production database"
//	alert.SlackChannelID = "C12345678"
//	alert.IssueFollowUpEnabled = true
//	alert.AutoResolveSeconds = 300
//
//	// Add fields
//	alert.Fields = []*common.Field{
//	    {Title: "Host", Value: "db-prod-01"},
//	    {Title: "Port", Value: "5432"},
//	}
//
//	// Add escalation
//	alert.Escalation = []*common.Escalation{
//	    {
//	        Severity:      common.AlertPanic,
//	        DelaySeconds:  300,
//	        SlackMentions: []string{"<!here>"},
//	    },
//	}
//
//	// Add webhook button
//	alert.Webhooks = []*common.Webhook{
//	    {
//	        ID:          "restart",
//	        URL:         "https://example.com/webhook/restart",
//	        ButtonText:  "Restart DB",
//	        ButtonStyle: common.WebhookButtonStyleDanger,
//	        AccessLevel: common.WebhookAccessLevelChannelAdmins,
//	    },
//	}
//
//	// Clean and validate
//	alert.Clean()
//	if err := alert.Validate(); err != nil {
//	    panic(err)
//	}
package common
