# ContextFlow 🧠

> **Shell history that understands your workflows, not just your commands.**

[![CI](https://github.com/Luv-Goel/contextflow/actions/workflows/ci.yml/badge.svg)](https://github.com/Luv-Goel/contextflow/actions)
[![Go version](https://img.shields.io/github/go-mod/go-version/Luv-Goel/contextflow)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Luv-Goel/contextflow)](https://github.com/Luv-Goel/contextflow/releases)

---

<!--
  TODO before launch: add demo GIF here
  Use `vhs` (https://github.com/charmbracelet/vhs) to record
-->

Every developer has typed `history | grep` in desperation. You know a command exists somewhere — you just can't find the *workflow* around it. What did you run before that? What directory were you in? What project?

**ContextFlow** remembers workflows, not just commands:

- 🔍 **Fuzzy search** your entire history with a beautiful TUI (`Ctrl+R` replacement)
- 🧠 **Workflow detection** — auto-groups related commands by project and time
- 🔁 **Replay** any past workflow step-by-step
- 📤 **Export** workflows as shell scripts or markdown runbooks
- 📁 **Project-aware** — history linked to git repos, not just directories

---

## Install

```bash
# One-liner
curl -fsSL https://raw.githubusercontent.com/Luv-Goel/contextflow/main/scripts/install.sh | bash
```

```bash
# Homebrew (coming soon)
brew install Luv-Goel/tap/contextflow
```

Then add to your shell:

```bash
# ~/.bashrc
eval "$(cf init bash)"

# ~/.zshrc
eval "$(cf init zsh)"

# ~/.config/fish/config.fish
cf init fish | source
```

Restart your shell and press **Ctrl+R** to try it.

---

## Usage

```bash
# Search history (Ctrl+R replacement)
cf search
cf search "docker"

# Browse auto-detected workflows
cf workflows

# Replay a workflow step-by-step
cf replay 42

# Dry run (preview without executing)
cf replay 42 --dry-run

# Export as shell script
cf export 42 > setup.sh

# Export as markdown runbook
cf export 42 --format md > RUNBOOK.md
```

---

## How workflow detection works

ContextFlow groups commands into workflows when they are:
- In the **same git repository** (or same directory)
- Within **30 minutes** of each other
- In the **same terminal session**

So your "docker setup", "deploy to prod", or "new project bootstrap" sequences are automatically recognized and replayable — without any manual tagging.

---

## Why not Atuin?

| Feature | Atuin | ContextFlow |
|---------|-------|-------------|
| Fuzzy search TUI | ✅ | ✅ |
| Cross-machine sync | ✅ | ❌ (planned) |
| E2E encrypted sync | ✅ | ❌ |
| **Workflow detection** | ❌ | ✅ |
| **Replay workflows** | ❌ | ✅ |
| **Export as script/runbook** | ❌ | ✅ |
| **Git repo awareness** | ❌ | ✅ |
| Single binary | ✅ | ✅ |
| Local-first | ✅ | ✅ |

**They're complementary.** Atuin excels at sync + search. ContextFlow adds workflow intelligence on top. (Atuin import coming in v0.2.)

---

## Data & Privacy

All data is stored locally in `~/.contextflow/history.db` (SQLite).  
Nothing is ever uploaded anywhere. No accounts, no servers, no telemetry.

---

## Roadmap

- [x] v0.1 — Shell hooks, fuzzy search TUI, workflow detection, replay, export
- [ ] v0.2 — Fish support, Atuin import, `cf stats`, manual workflow tagging
- [ ] v0.3 — Natural language search, team snippet sharing

---

## Contributing

PRs and issues welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) (coming soon).

```bash
git clone https://github.com/Luv-Goel/contextflow
cd contextflow
go build ./...
go test ./...
```

---

## License

MIT © [Luv-Goel](https://github.com/Luv-Goel)

---

*Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) 🧋*
