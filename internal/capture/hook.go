package capture

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// Record captures a command and writes it to the database.
// Called by shell hooks: cf record --cmd "..." --dir "..." --exit 0 --duration 123 --session "..."
func Record(command, directory string, exitCode int, durationMs int64, sessionID string) error {
	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	hostname, _ := os.Hostname()

	cmd := db.Command{
		Command:    command,
		Directory:  directory,
		GitRepo:    gitRemote(directory),
		GitBranch:  gitBranch(directory),
		ExitCode:   exitCode,
		DurationMs: durationMs,
		SessionID:  sessionID,
		Hostname:   hostname,
		RecordedAt: time.Now(),
	}

	_, err = database.RecordCommand(cmd)
	return err
}

// gitRemote returns the git remote origin URL for a directory, or "".
func gitRemote(dir string) string {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitBranch returns the current git branch for a directory, or "".
func gitBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// InitShell prints the shell hook script for the given shell.
// Usage: eval "$(cf init bash)"
func InitShell(shell string) (string, error) {
	cfPath, err := os.Executable()
	if err != nil {
		cfPath = "cf"
	}

	switch strings.ToLower(shell) {
	case "bash":
		return bashHook(cfPath), nil
	case "zsh":
		return zshHook(cfPath), nil
	case "fish":
		return fishHook(cfPath), nil
	default:
		return "", fmt.Errorf("unsupported shell %q — supported: bash, zsh, fish", shell)
	}
}

func bashHook(cfPath string) string {
	return fmt.Sprintf(`
# ContextFlow bash integration
__cf_session_id="%s"
__cf_cmd_start=0
__cf_last_cmd=""

__cf_preexec() {
    __cf_cmd_start=$(date +%%s%%3N)
    __cf_last_cmd="$1"
}

__cf_precmd() {
    local exit_code=$?
    if [[ -n "$__cf_last_cmd" && "$__cf_last_cmd" != "cf"* ]]; then
        local now
        now=$(date +%%s%%3N)
        local duration=$(( now - __cf_cmd_start ))
        %s record \
            --cmd "$__cf_last_cmd" \
            --dir "$PWD" \
            --exit "$exit_code" \
            --duration "$duration" \
            --session "$__cf_session_id" &>/dev/null &
    fi
    __cf_last_cmd=""
}

# Install hooks using bash-preexec if available, else manual PROMPT_COMMAND
if [[ $(type -t __bp_install) == "function" ]]; then
    preexec_functions+=(__cf_preexec)
    precmd_functions+=(__cf_precmd)
else
    trap '__cf_preexec "$BASH_COMMAND"' DEBUG
    PROMPT_COMMAND="__cf_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi

# Ctrl+R override
__cf_search() {
    local selected
    selected=$(%s search --print 2>/dev/null)
    if [[ -n "$selected" ]]; then
        READLINE_LINE="$selected"
        READLINE_POINT=${#READLINE_LINE}
    fi
}
bind -x '"\C-r": __cf_search'
`, sessionID(), cfPath, cfPath)
}

func zshHook(cfPath string) string {
	return fmt.Sprintf(`
# ContextFlow zsh integration
__cf_session_id="%s"
__cf_cmd_start=0
__cf_last_cmd=""

__cf_preexec() {
    __cf_cmd_start=$(($(date +%%s%%3N)))
    __cf_last_cmd="$1"
}

__cf_precmd() {
    local exit_code=$?
    if [[ -n "$__cf_last_cmd" && "$__cf_last_cmd" != cf* ]]; then
        local now=$(($(date +%%s%%3N)))
        local duration=$(( now - __cf_cmd_start ))
        %s record \
            --cmd "$__cf_last_cmd" \
            --dir "$PWD" \
            --exit "$exit_code" \
            --duration "$duration" \
            --session "$__cf_session_id" &>/dev/null &
    fi
    __cf_last_cmd=""
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec __cf_preexec
add-zsh-hook precmd __cf_precmd

# Ctrl+R override
__cf_search_widget() {
    local selected
    selected=$(%s search --print 2>/dev/null)
    if [[ -n "$selected" ]]; then
        BUFFER="$selected"
        CURSOR=${#BUFFER}
    fi
    zle reset-prompt
}
zle -N __cf_search_widget
bindkey "^R" __cf_search_widget
`, sessionID(), cfPath, cfPath)
}

func fishHook(cfPath string) string {
	return fmt.Sprintf(`
# ContextFlow fish integration
set -g __cf_session_id "%s"

function __cf_on_event --on-event fish_preexec
    set -g __cf_cmd_start (date +%%s%%3N)
    set -g __cf_last_cmd $argv[1]
end

function __cf_on_postexec --on-event fish_postexec
    set exit_code $status
    if test -n "$__cf_last_cmd"; and not string match -q 'cf*' $__cf_last_cmd
        set now (date +%%s%%3N)
        set duration (math $now - $__cf_cmd_start)
        %s record \
            --cmd "$__cf_last_cmd" \
            --dir "$PWD" \
            --exit "$exit_code" \
            --duration "$duration" \
            --session "$__cf_session_id" &>/dev/null &
    end
end
`, sessionID(), cfPath)
}

func sessionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
