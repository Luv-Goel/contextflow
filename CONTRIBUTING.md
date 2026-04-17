# Contributing to ContextFlow

Thanks for considering a contribution! ContextFlow is a small project and every PR, bug report, and idea genuinely helps.

## Quick Start

```bash
git clone https://github.com/Luv-Goel/contextflow
cd contextflow
go build ./...
go test ./...
```

No external services, no accounts, no environment setup needed — it's all local SQLite.

## Project Structure

```
cmd/cf/          → CLI entry point (Cobra commands)
internal/
  capture/       → Shell hook generation + command recording + secret filtering
  db/            → SQLite schema, queries, stats
  workflow/      → Auto-detection and replay engine
  tui/           → Bubble Tea TUI components (search + workflow browser)
  export/        → Shell script + Markdown export
shell/           → Shell integration files (reference)
scripts/         → install.sh
```

## How to Add a Feature

1. Fork the repo and create a branch: `git checkout -b feat/my-thing`
2. Write the feature + tests
3. Run `go test ./...` and `go vet ./...`
4. Open a PR with a clear description of what it does and why

## Good First Issues

- Fish shell history import support
- `cf search` — add date range filter flags
- `cf workflows` — add filter by repo flag
- Improve secret detection patterns
- Windows PowerShell history import

## Bug Reports

Please include:
- Your shell (bash/zsh/fish) and OS
- The command you ran
- What you expected vs what happened
- Output of `cf version`

## Code Style

- Standard Go formatting (`gofmt`)
- Keep functions small and focused
- Add tests for new logic — especially in `workflow/` and `capture/`
- No external API calls, no telemetry, no cloud dependencies

## License

By contributing, you agree your work will be licensed under MIT.
