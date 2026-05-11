# Changelog

All notable changes to ContextFlow are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial CI pipeline with build, test, and coverage
- Dependabot configuration for Go modules and GitHub Actions
- golangci-lint configuration
- Tests for story generation and workflow replay
- CODE_OF_CONDUCT.md

### Fixed
- CI workflow Go version mismatch (was 1.22, now matches go.mod 1.25.0)
- DB test isolation — tests now use t.Setenv instead of os.Setenv
- Capture test helper now uses temp directory instead of ~/.contextflow
- Version string synced from v0.1.10 to v0.1.15
- Removed duplicated shortDir() function

## [v0.1.15] - 2026-04-19

### Added
- DB-backed workflow save/load for export
- Confirm flag for delete command

### Fixed
- Stats average duration calculation
- Duration conversion in recording

## [v0.1.14] - 2026-04-19

### Fixed
- Shell hook --shell flag handling
- Uninstall command improvements

## [v0.1.13] - 2026-04-19

### Fixed
- `cf init` now properly passes --shell flag

## [v0.1.12] - 2026-04-19

### Added
- Atuin history import command
- Install command for quick setup

## [v0.1.11] - 2026-04-19

### Changed
- Code cleanup and removal of stale TODO comments

## [v0.1.10] - 2026-04-19

### Added
- Duration string parsing for story command
- Import history wired to database
- Replay wired to database
- Export wired to database
- GetWorkflowByID database method
- Tag and delete wired to database
- UpdateWorkflowName and DeleteWorkflow database methods

### Changed
- Full database-backed workflow operations

## [v0.1.9] - 2026-04-18

### Added
- Commands reference section in README
- Uninstall command

## [v0.1.8] - 2026-04-18

### Changed
- Cleaned up unused export imports

## [v0.1.7] - 2026-04-18

### Added
- Shell integration for fish shell
- `cf init` command with shell flags

## [v0.1.6] - 2026-04-17

### Added
- Shell hook generation (bash/zsh)
- Command recording via PROMPT_COMMAND/precmd
- SQLite-backed storage
- Fuzzy search TUI via Bubble Tea
- Workflow detection engine
- Workflow replay (interactive + dry-run)
- Export as shell script or Markdown
- Usage statistics
- Secret filtering for tokens/passwords
- Local-first, zero dependency design

[Unreleased]: https://github.com/Luv-Goel/contextflow/compare/v0.1.15...HEAD
[v0.1.15]: https://github.com/Luv-Goel/contextflow/compare/v0.1.14...v0.1.15
[v0.1.14]: https://github.com/Luv-Goel/contextflow/compare/v0.1.13...v0.1.14
[v0.1.13]: https://github.com/Luv-Goel/contextflow/compare/v0.1.12...v0.1.13
[v0.1.12]: https://github.com/Luv-Goel/contextflow/compare/v0.1.11...v0.1.12
[v0.1.11]: https://github.com/Luv-Goel/contextflow/compare/v0.1.10...v0.1.11
[v0.1.10]: https://github.com/Luv-Goel/contextflow/compare/v0.1.9...v0.1.10
[v0.1.9]: https://github.com/Luv-Goel/contextflow/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/Luv-Goel/contextflow/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/Luv-Goel/contextflow/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/Luv-Goel/contextflow/releases/tag/v0.1.6
