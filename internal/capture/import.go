package capture

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// ImportResult summarises what was imported.
type ImportResult struct {
	Total    int
	Imported int
	Skipped  int // duplicates or secrets
}

// ImportHistoryFile imports commands from a shell history file (bash/zsh format).
// Lines starting with '#' (timestamps) are skipped; blank lines are skipped.
func ImportHistoryFile(path string, database *db.DB) (ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	var result ImportResult
	// Use a fixed past time, incrementing by 1s per command so ordering is preserved
	baseTime := time.Now().Add(-365 * 24 * time.Hour)
	i := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result.Total++

		cmd := db.Command{
			Command:    sanitize(line),
			Directory:  "",
			RecordedAt: baseTime.Add(time.Duration(i) * time.Second),
			SessionID:  "import",
		}

		if _, err := database.RecordCommand(cmd); err != nil {
			result.Skipped++
		} else {
			result.Imported++
		}
		i++
	}
	return result, scanner.Err()
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
