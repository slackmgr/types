# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `slack-manager-common`, a Go shared library package providing common interfaces and data structures for the Slack Manager system. It defines contracts for database access, logging, metrics, and core domain types like alerts and issues.

**Module:** `github.com/peteraglen/slack-manager-common`
**Go Version:** 1.25

## Build Commands

```bash
make init    # go mod tidy - download/clean dependencies
make test    # Full suite: gosec, go fmt, go test, go vet
make lint    # golangci-lint run ./...
```

Run a single test:
```bash
go test -v -run TestFunctionName ./...
```

## Architecture

### Core Interfaces (Dependency Injection Points)

- **DB** (`db.go`) - Database abstraction for persisting alerts, issues, move mappings, and channel processing state. Implementations provided by consumers.
- **Logger** (`logger.go`) - Structured logging interface with Debug/Info/Error levels and field support.
- **Metrics** (`metrics.go`) - Prometheus-style metrics (counter, gauge, histogram).

Each interface has a no-op implementation (`noop_logger.go`, `noop_metrics.go`) for testing.

### Core Domain Types

- **Alert** (`alert.go`) - Central type representing an alert with comprehensive validation. Contains severity, header, text, fields, webhooks, escalations. Use constructor functions: `NewPanicAlert()`, `NewErrorAlert()`, `NewWarningAlert()`, `NewResolvedAlert()`, `NewInfoAlert()`.

- **Issue** / **MoveMapping** (`issue.go`, `move_mapping.go`) - Interfaces for issue tracking and escalation moves. Stored as opaque JSON to allow implementation flexibility.

- **ChannelProcessingState** (`channel_processing_state.go`) - Tracks per-channel processing timestamps and open issue counts.

- **Webhook types** - Interactive webhook buttons with access levels, display modes, and input forms.

### Testing Utilities

- `dbtests/tests.go` - Shared database test suite that can be run against any DB implementation.
- `InMemoryFifoQueue` - Test-only FIFO queue (NOT for production use).

## Code Patterns

**Validation:** Alert has extensive validation methods (`ValidateChannelID`, `ValidateHeader`, etc.) that return descriptive errors. Call `Validate()` to run all validations.

**Cleaning:** Alert has cleaning methods (`CleanHeader`, `CleanText`, etc.) that normalize and truncate fields. Call `Clean()` to clean all fields.

**Testing:** Uses table-driven tests with `t.Parallel()`. All tests use `testify/assert` and `testify/require`.

**Limits:** Key validation limits are defined as constants (e.g., `MaxHeaderLen = 130`, `MaxTextLen = 10000`, `MaxFields = 20`, `MaxWebhooks = 5`, `MaxEscalations = 10`).
