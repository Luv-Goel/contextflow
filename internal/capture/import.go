package capture

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// ImportResult summarises what was imported.
type ImportResult struct {
	Total    int
	Imported int
	Skipped  int // secrets or errors
}

// ImportHistoryFile imports commands from a shell history file.
// Supports:
//   - Plain format (bash):  <command>
//   - Zsh extended format:  : <timestamp>:<duration>;<command>
func ImportHistoryFile(path string, database *db.DB) (ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	var result ImportResult
	// Fallback base time for entries without a timestamp
	baseTime := time.Now().Add(-365 * 24 * time.Hour)
	fallbackOffset := 0

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // handle long lines
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		cmd, ts := parseLine(line, baseTime, fallbackOffset)
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}
		result.Total++
		fallbackOffset++

		entry := db.Command{
			Command:    sanitize(cmd),
			Directory:  "",
			RecordedAt: ts,
			SessionID:  "import",
		}

		if _, err := database.RecordCommand(entry); err != nil {
			result.Skipped++
		} else {
			result.Imported++
		}
	}
	return result, scanner.Err()
}

// parseLine handles both plain bash history and zsh extended history format.
// Zsh extended format: ": <unix_ts>:<duration>;<command>"
func parseLine(line string, baseTime time.Time, offset int) (cmd string, ts time.Time) {
	// Zsh extended history: ": 1713000000:0;git push origin main"
	if strings.HasPrefix(line, ": ") {
		parts := strings.SplitN(line, ";", 2)
		if len(parts) == 2 {
			meta := strings.TrimPrefix(parts[0], ": ")
			metaParts := strings.SplitN(meta, ":", 2)
			if len(metaParts) == 2 {
				if epoch, err := strconv.ParseInt(strings.TrimSpace(metaParts[0]), 10, 64); err == nil {
					return strings.TrimSpace(parts[1]), time.Unix(epoch, 0)
				}
			}
		}
	}
	// Plain format — use baseTime + offset for stable ordering
	return strings.TrimSpace(line), baseTime.Add(time.Duration(offset) * time.Second)
}

// DetectHistoryFiles returns the likely shell history file paths for this system.
func DetectHistoryFiles() []string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".bash_history"),
		filepath.Join(home, ".zsh_history"),
		filepath.Join(home, ".local", "share", "fish", "fish_history"),
	}
	var found []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}
