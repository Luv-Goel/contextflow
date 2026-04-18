# ContextFlow 🧠

> **Shell history that understands your workflows, not just your commands.**

[![CI](https://github.com/Luv-Goel/contextflow/actions/workflows/ci.yml/badge.svg)](https://github.com/Luv-Goel/contextflow/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Luv-Goel/contextflow)](https://goreportcard.com/report/github.com/Luv-Goel/contextflow)
[![Go version](https://img.shields.io/github/go-mod/go-version/Luv-Goel/contextflow)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Luv-Goel/contextflow?include_prereleases)](https://github.com/Luv-Goel/contextflow/releases)

---

Every developer has typed `history | grep` in desperation. You know a command exists somewhere — you just can't find the *sequence* around it. What else did you run? In what directory? For which project?

**ContextFlow** remembers the workflow, not just the command.

---

## Features

- 🔍 **Fuzzy search TUI** — beautiful Ctrl+R replacement with project context
- 🧠 **Workflow detection** — auto-groups related commands by git repo + session
- 🔁 **Replay** — step through any past workflow interactively  
- 📤 **Export** — workflows as `.sh` scripts or Markdown runbooks
- 📊 **Stats** — most-used commands, busiest repos, session analytics
- 📥 **Import** — bring in your existing `~/.bash_history` / `~/.zsh_history`
- 🔒 **Secret filtering** — tokens, passwords, and API keys are never recorded
- 📁 **Project-aware** — history linked to git repos, not just directories
- 💾 **Local-first** — everything in `~/.contextflow/history.db` (SQLite), nothing uploaded

---

## Install

**One-liner (Linux & macOS):**

```bash
curl -fsSL https://raw.githubusercontent.com/Luv-Goel/contextflow/main/scripts/install.sh | bash
```

**go install (recommended):**

```bash
go install github.com/Luv-Goel/contextflow/cmd/cf@latest
```

**Homebrew:**

```bash
brew install Luv-Goel/tap/contextflow
```

**Build from source:**

```bash
git clone https://github.com/Luv-Goel/contextflow
cd contextflow
go build -o ~/bin/cf ./cmd/cf
```

---

## Setup

Add to your shell config and restart:

```bash
# bash — add to ~/.bashrc
eval "$(cf init bash)"

# zsh — add to ~/.zshrc
eval "$(cf init zsh)"

# fish — add to ~/.config/fish/config.fish
cf init fish | source
```

Then press **Ctrl+R** to search, or run `cf workflows` to explore.

**Import your existing history:**

```bash
cf import   # auto-detects ~/.bash_history and ~/.zsh_history
```

---

## Usage

```bash
# Search (Ctrl+R replacement — fuzzy TUI)
cf search
cf search "docker"

# Browse auto-detected workflows
cf workflows

# Step through a workflow interactively
cf replay 42

# Preview without executing
cf replay 42 --dry-run

# Export as shell script
cf export 42 > setup.sh

# Export as Markdown runbook
cf export 42 --format md > RUNBOOK.md

# Give a workflow a name
cf tag 42 "my docker setup"

# Usage statistics
cf stats

# Delete a workflow
cf delete 42

# Import existing shell history
cf import
cf import ~/.zsh_history
```

---

## How Workflow Detection Works

ContextFlow automatically groups commands into workflows when they are:

1. Recorded in the **same terminal session**
2. Within **30 minutes** of each other  
3. In the **same git repository** or directory

So if you ran `docker build`, `docker run`, `docker ps`, and `curl localhost:8080` in one session — that's a workflow. Name it "docker-setup" and replay it any time.

**No manual tagging required.** Just work normally.

---

## Privacy & Security

- All data stored in `~/.contextflow/history.db` — never leaves your machine
- **Secret filtering** — commands matching patterns for passwords, tokens, API keys, and credentials are recorded as `[redacted]` automatically
- No telemetry, no analytics, no accounts, no network calls

---

## Why Not Atuin?

| Feature | Atuin | ContextFlow |
|---------|-------|-------------|
| Fuzzy search TUI | ✅ | ✅ |
| Cross-machine sync | ✅ | — (planned) |
| E2E encrypted sync | ✅ | — |
| **Workflow detection** | ❌ | ✅ |
| **Workflow replay** | ❌ | ✅ |
| **Export as script/runbook** | ❌ | ✅ |
| **Git repo awareness** | ❌ | ✅ |
| **Secret filtering** | ❌ | ✅ |
| **Usage stats** | partial | ✅ |
| **History import** | ✅ | ✅ |
| Single binary | ✅ | ✅ |
| Local-first | ✅ | ✅ |

They're complementary. **Atuin** is best for sync + search across machines. **ContextFlow** is best for understanding and replaying what you actually did. Use both — Atuin import is in v0.2.

---

## Roadmap

- [x] v0.1 — Shell hooks (bash/zsh/fish), fuzzy search TUI, workflow detection, replay, export, stats, import, secret filtering
- [ ] v0.2 — Atuin import, `cf share` (Gist export), improved workflow naming, Fish history import
- [ ] v0.3 — Natural language search, team snippet library, VS Code extension

---

## Contributing

PRs and issues are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
git clone https://github.com/Luv-Goel/contextflow
cd contextflow
go test ./...
```

---

## License

MIT © [Luv-Goel](https://github.com/Luv-Goel)

---

*Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) 🧋 · Powered by SQLite · Zero cloud dependencies*
