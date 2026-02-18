# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `slack-manager-common`, a Go shared library package providing common interfaces and data structures for the Slack Manager system. It defines contracts for database access, logging, metrics, and core domain types like alerts and issues.

**Module:** `github.com/peteraglen/slack-manager-common`
**Go Version:** 1.25

## Git Conventions

- Do not use the `-C` flag when running git commands directly in the repo directory, as it is not needed.

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

## Code Quality Requirements

**CRITICAL:** Before committing any changes, you MUST ensure both `make test` and `make lint` pass without errors. This applies to ALL changes, regardless of who made them (human, Claude, or other tools/linters).

```bash
# Always run before committing:
make test    # Must pass: gosec, go fmt, go test (with race detector), go vet
make lint    # Must pass: golangci-lint with zero issues
```

If either command fails:
1. Fix all reported issues
2. Re-run both commands to verify
3. Only commit after both pass

This ensures code quality, prevents broken releases, and maintains consistency across the codebase.

## Tagging and Releases

### Process

1. **Update `CHANGELOG.md` first** — this is MANDATORY before creating any tag.
   - Review every commit since the last tagged commit: `git log <last-tag>..HEAD --oneline`
   - Every commit MUST be considered and represented under the correct section (`Added`, `Changed`, `Fixed`, `Removed`)
   - Add the new version section above `[Unreleased]` with today's date
   - Update the comparison links at the bottom of the file

2. **Commit the changelog:**
   ```bash
   git add CHANGELOG.md
   git commit -m "Update CHANGELOG for vX.Y.Z"
   ```

3. **Create and push the tag:**
   ```bash
   git tag vX.Y.Z
   git push origin main
   git push origin vX.Y.Z
   ```

4. **Create the GitHub release:**
   ```bash
   gh release create vX.Y.Z --repo peteraglen/slack-manager-common --title "vX.Y.Z" --notes "..."
   ```
   Use the same content as the changelog entry for the release notes.

### Versioning

Follows [Semantic Versioning](https://semver.org/):
- **Patch** (`Z`): bug fixes, CI/infra changes, documentation updates
- **Minor** (`Y`): new backwards-compatible features or functionality
- **Major** (`X`): breaking changes to the public API

### Rules

- **NEVER** create a tag without updating `CHANGELOG.md` first
- **ALWAYS** review all commits since the last tag — do not rely on memory or summaries

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
