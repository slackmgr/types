# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/slackmgr/slack-manager-common/compare/v0.2.2...HEAD
[0.2.2]: https://github.com/slackmgr/slack-manager-common/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/slackmgr/slack-manager-common/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/slackmgr/slack-manager-common/compare/v0.1.4...v0.2.0
[0.1.4]: https://github.com/slackmgr/slack-manager-common/releases/tag/v0.1.4
