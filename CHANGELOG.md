# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.1] - 2026-05-08

### Changed
- `Alert.ValidateWebhooks()` now rejects `Webhook.SkipConfirmationDialog: true` when `PlainTextInput` or `CheckboxInput` is non-empty (input fields are collected via the confirmation dialog, so skipping the dialog leaves nowhere to render them)

## [0.6.0] - 2026-05-08

### Added
- `Webhook.SkipConfirmationDialog` (bool, defaults to `false`): when `true`, the core manager posts the webhook without first showing a confirmation dialog. Validation rejects `SkipConfirmationDialog: true` together with `ButtonStyle: WebhookButtonStyleDanger`

## [0.5.2] - 2026-05-08

### Changed
- `Alert.ValidateWebhooks()` no longer rejects `WebhookPlainTextInput.InitialValue` when its length is shorter than `MinLength` — `MinLength` constrains the user-submitted value, not the pre-filled initial value

## [0.5.1] - 2026-05-07

### Changed
- Fix incorrect godoc on `Webhook.ConfirmationText`, `AccessLevel`, `DisplayMode`, and on `WebhookPlainTextInput`/`WebhookCheckboxInput`/`WebhookCheckboxOption` (replace ambiguous "webhook payload" wording with the correct `WebhookCallback` field references)
- `Clean()` now defaults empty `Webhook.ButtonStyle` to `WebhookButtonStyleDefault` and empty `Webhook.AccessLevel` to `WebhookAccessLevelGlobalAdmins` (previously `Clean()` normalised `"default"` to empty string for `ButtonStyle`)
- `WebhookButtonStyleDefault` is now an exported constant and accepted as a valid style
- Clarify that `WebhookAccessLevelChannelAdmins`/`ChannelMembers` include higher access levels

## [0.5.0] - 2026-05-06

### Changed
- `WebhookCallback.Input` renamed to `PlainTextInput`; JSON tag changed from `"input"` to `"plainTextInput"` (**breaking** — consumers must update field references and any JSON deserialization)
- Fix incorrect and incomplete godoc comments on `Alert` fields: `CorrelationID` hash order, `HeaderWhenResolved`/`TextWhenResolved` state description, `IssueFollowUpEnabled` fire-and-forget behavior, `AutoResolveAsInconclusive` conditions, `NotificationDelaySeconds` dependency on `IssueFollowUpEnabled`, `IgnoreIfTextContains` case-sensitivity and field scope, `Webhook.Payload`/`PlainTextInput`/`CheckboxInput`
- Add struct-level and field-level godoc to `WebhookCallback`

## [0.4.1] - 2026-04-14

### Changed
- Refactor `Clean()`: extract `truncateStringIfNeeded` helper to eliminate repetition
- `Clean()` now normalises severity aliases `"resolve"`, `"recovered"`, and `"recover"` to `AlertResolved`

## [0.4.0] - 2026-02-26

### Changed
- `Metrics` interface: replace `Add`, `Inc`, `Set` with type-prefixed `CounterAdd`, `CounterInc`, `GaugeSet`; add `GaugeAdd` for atomic gauge adjustments (breaking — consumers must update call sites)

## [0.3.1] - 2026-02-20

### Added
- `InMemoryDB`: in-memory implementation of the `DB` interface for test use, modelled after `InMemoryFifoQueue`

### Changed
- Update README

### Fixed
- `dbtests.TestMoveIssue`: reset DB state before running to avoid interference from prior tests
- `dbtests.TestConcurrentSaveIssue`: copy issue per goroutine to eliminate data race under `-race`

## [0.3.0] - 2026-02-19

### Changed
- Rename repository from `slack-manager-common` to `types`: update module path to `github.com/slackmgr/types`, all import paths, README badges/links, and CHANGELOG comparison links
- Rename Go package from `common` to `types`: consumers must update `common.Xxx` references to `types.Xxx` (**breaking**)
- Update README usage example to use `types.` package prefix

## [0.2.2] - 2026-02-19

### Changed
- Migrate GitHub organisation from `peteraglen` to `slackmgr`: update module path to `github.com/slackmgr/slack-manager-common`, all import paths, README badges/links, and CHANGELOG comparison links
- Add tagging, release, and git conventions to CLAUDE.md

## [0.2.1] - 2026-02-18

### Added
- Concurrency control in CI: cancel outdated workflow runs when new commits are pushed to the same ref

### Changed
- Standardize on `google/uuid`, remove `ksuid` dependency
- Expand dbtests with comprehensive test coverage
- Clarify in README when to use this library directly
- Update README license section
- Reduce CI test output verbosity: remove `-v` flag, show per-package coverage instead of per-function
- Simplify CI to test only Go 1.25
- Pin all GitHub Actions to specific versions

### Fixed
- Fix lint issues flagged by golangci-lint
- Fix golangci-lint v2 incompatibility (`--out-format` flag removed, output format configured in `.golangci.yaml`)

## [0.2.0] - 2026-02-18

### Added
- Comprehensive README with installation, usage examples, and API documentation
- MIT License for open source distribution
- Package-level documentation in doc.go for godoc/pkg.go.dev
- GitHub Actions CI workflow (testing, linting, security scanning)
- Badges in README (pkg.go.dev, Go Report Card, license, CI status)

### Changed
- Repository prepared for public open source release

## [0.1.4] - (Previous Release)

See git history for changes in v0.1.4 and earlier versions.

[Unreleased]: https://github.com/slackmgr/types/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/slackmgr/types/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/slackmgr/types/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/slackmgr/types/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/slackmgr/types/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/slackmgr/types/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/slackmgr/types/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/slackmgr/types/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/slackmgr/types/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/slackmgr/types/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/slackmgr/types/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/slackmgr/types/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/slackmgr/types/compare/v0.1.4...v0.2.0
[0.1.4]: https://github.com/slackmgr/types/releases/tag/v0.1.4
