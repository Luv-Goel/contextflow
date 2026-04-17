package story

import (
	"fmt"
	"strings"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// Generate creates a narrative from the recorded command history
func Generate(database *db.DB, since time.Duration) string {
	commands, err := database.GetCommandsSince(since)
	if err != nil || len(commands) == 0 {
		return "No commands recorded in the last " + formatDuration(since)
	}

	// Group stats
	var (
		totalCommands    = len(commands)
		gitCommands   int
		buildCommands int
		failedCommands int
		searches     int
		byRepo      = make(map[string]int)
		byHour      = make(map[int]int)
		totalDuration time.Duration
		regrets     int
		successes   int
	)

	successIndicators := []string{"push", "deploy", "merge", "release"}

	for _, cmd := range commands {
		lower := strings.ToLower(cmd.Command)

		// Category detection
		if strings.HasPrefix(lower, "git ") || lower == "git" {
			gitCommands++
		}
		if strings.Contains(lower, "build") || strings.Contains(lower, "go ") ||
			strings.Contains(lower, "make ") || strings.Contains(lower, "npm ") {
			buildCommands++
		}
		if strings.Contains(lower, "grep") || strings.Contains(lower, "search") ||
			strings.Contains(lower, "cf search") {
			searches++
		}

		// Outcome tracking
		if cmd.ExitCode != 0 {
			failedCommands++
		}
		if strings.Contains(lower, "rm -rf") || strings.Contains(lower, "reset --hard") {
			regrets++
		}
		for _, s := range successIndicators {
			if strings.Contains(lower, s) {
				successes++
				break
			}
		}

		// Repo grouping
		if cmd.GitRepo != "" {
			byRepo[cmd.GitRepo]++
		}

		// Hour grouping
		hour := cmd.RecordedAt.Hour()
		byHour[hour]++

		totalDuration += time.Duration(cmd.DurationMs) * time.Millisecond
	}

	// Find busiest repo
	busiestRepo := ""
	busiestCount := 0
	for repo, count := range byRepo {
		if count > busiestCount {
			busiestRepo = repo
			busiestCount = count
		}
	}

	// Find busiest hour
	busiestHour := 0
	busiestHourCount := 0
	for hour, count := range byHour {
		if count > busiestHourCount {
			busiestHour = hour
			busiestHourCount = count
		}
	}

	// Build the narrative
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📟 **Shell Story — Last %s**\n\n", formatDuration(since)))
	b.WriteString(fmt.Sprintf("You ran %d commands. ", totalCommands))

	if gitCommands > 0 {
		b.WriteString(fmt.Sprintf("%d of them were git. ", gitCommands))
	}
	if failedCommands > 0 {
		pct := failedCommands * 100 / totalCommands
		b.WriteString(fmt.Sprintf("%d%% failed. ", pct))
	}

	b.WriteString("\n\n")

	// Context
	if busiestRepo != "" {
		repoName := busiestRepo
		if slash := strings.LastIndex(busiestRepo, "/"); slash >= 0 {
			repoName = busiestRepo[slash+1:]
		}
		b.WriteString(fmt.Sprintf("Most of your time was in `%s` (%d commands).\n", repoName, busiestCount))
	}

	// Timing
	if busiestHourCount > 0 {
		ampm := "AM"
		displayHour := busiestHour
		if displayHour >= 12 {
			ampm = "PM"
			if displayHour > 12 {
				displayHour -= 12
			}
		}
		b.WriteString(fmt.Sprintf("Your busiest hour was %d %s (%d commands).\n", displayHour, ampm, busiestHourCount))
	}

	// Total time
	if totalDuration > 0 {
		mins := int(totalDuration.Minutes())
		if mins >= 60 {
			b.WriteString(fmt.Sprintf("Total command time: ~%dh %dm.\n", mins/60, mins%60))
		} else {
			b.WriteString(fmt.Sprintf("Total command time: ~%d minutes.\n", mins))
		}
	}

	// Verdict
	b.WriteString("\n")
	if regrets > successes {
		b.WriteString("⚠️ You deleted more than you committed. A bold strategy.")
	} else if successes > regrets && successes > 3 {
		b.WriteString("🚀 You shipped things. That's what matters.")
	} else if buildCommands > 5 {
		b.WriteString("🔨 You were building. Compiling, iterating, ship it.")
	} else if searches > 5 {
		b.WriteString("🔍 So many searches. You forgot more than you learned.")
	} else if gitCommands > 10 {
		b.WriteString("📝 Git was your companion today. As always.")
	} else {
		b.WriteString("💤 A quiet day. Or maybe you're just getting started.")
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		return "day"
	} else if d >= time.Hour {
		return "hour"
	} else {
		return "session"
	}
}