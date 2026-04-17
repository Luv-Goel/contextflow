package workflow

import (
	"fmt"
	"strings"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

const (
	// Commands within this window (same session + repo) form a workflow.
	workflowWindow = 30 * time.Minute
	// Minimum commands to be considered a workflow.
	minWorkflowSize = 3
)

// Detect groups a slice of commands into workflows based on:
// - Same git repo (or same directory)
// - Time proximity (within workflowWindow)
// - Session continuity
func Detect(commands []db.Command) []db.Workflow {
	if len(commands) == 0 {
		return nil
	}

	// Sort by time ascending (commands usually come in desc from DB)
	sorted := make([]db.Command, len(commands))
	copy(sorted, commands)
	sortByTime(sorted)

	var workflows []db.Workflow
	var current []db.Command

	flush := func() {
		if len(current) >= minWorkflowSize {
			w := db.Workflow{
				Name:      autoName(current),
				GitRepo:   current[0].GitRepo,
				CreatedAt: current[0].RecordedAt,
				UpdatedAt: current[len(current)-1].RecordedAt,
				Commands:  current,
			}
			workflows = append(workflows, w)
		}
		current = nil
	}

	for i, cmd := range sorted {
		if i == 0 {
			current = append(current, cmd)
			continue
		}

		prev := sorted[i-1]
		gap := cmd.RecordedAt.Sub(prev.RecordedAt)
		sameRepo := cmd.GitRepo != "" && cmd.GitRepo == prev.GitRepo
		sameDir := cmd.Directory == prev.Directory
		sameSession := cmd.SessionID == prev.SessionID

		withinWindow := gap <= workflowWindow
		related := (sameRepo || sameDir) && withinWindow && sameSession

		if related {
			current = append(current, cmd)
		} else {
			flush()
			current = append(current, cmd)
		}
	}
	flush()

	return workflows
}

// autoName generates a human-readable name for a workflow based on the commands.
func autoName(cmds []db.Command) string {
	// Use the first "meaningful" command (not cd, ls, etc.)
	skippable := map[string]bool{
		"cd": true, "ls": true, "ll": true, "pwd": true,
		"clear": true, "echo": true, "cat": true,
	}

	for _, cmd := range cmds {
		parts := strings.Fields(cmd.Command)
		if len(parts) == 0 {
			continue
		}
		base := parts[0]
		if !skippable[base] {
			name := base
			if len(parts) > 1 {
				name = fmt.Sprintf("%s %s", base, parts[1])
			}
			// Truncate and clean
			if len(name) > 40 {
				name = name[:40]
			}
			return name
		}
	}
	// Fallback: use directory name
	if len(cmds) > 0 && cmds[0].Directory != "" {
		parts := strings.Split(cmds[0].Directory, "/")
		return parts[len(parts)-1]
	}
	return "workflow"
}

func sortByTime(cmds []db.Command) {
	// Simple insertion sort (lists are usually small)
	for i := 1; i < len(cmds); i++ {
		for j := i; j > 0 && cmds[j].RecordedAt.Before(cmds[j-1].RecordedAt); j-- {
			cmds[j], cmds[j-1] = cmds[j-1], cmds[j]
		}
	}
}
