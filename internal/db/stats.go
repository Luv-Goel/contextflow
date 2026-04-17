package db

import "fmt"

// Stats holds usage analytics.
type Stats struct {
	TotalCommands   int
	UniqueCommands  int
	TotalWorkflows  int
	TopCommands     []CommandFreq
	TopRepos        []RepoFreq
	AvgDurationMs   int64
	TotalSessions   int
}

// CommandFreq is a command + how often it was used.
type CommandFreq struct {
	Command string
	Count   int
}

// RepoFreq is a git repo + how many commands were run in it.
type RepoFreq struct {
	Repo  string
	Count int
}

// GetStats returns aggregate usage statistics.
func (db *DB) GetStats() (Stats, error) {
	var s Stats

	if err := db.QueryRow(`SELECT COUNT(*), COUNT(DISTINCT command) FROM commands`).
		Scan(&s.TotalCommands, &s.UniqueCommands); err != nil {
		return s, fmt.Errorf("stats totals: %w", err)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM workflows`).Scan(&s.TotalWorkflows); err != nil {
		return s, fmt.Errorf("stats workflows: %w", err)
	}

	if err := db.QueryRow(`SELECT COUNT(DISTINCT session_id) FROM commands WHERE session_id != '' AND session_id != 'import'`).
		Scan(&s.TotalSessions); err != nil {
		return s, fmt.Errorf("stats sessions: %w", err)
	}

	if err := db.QueryRow(`SELECT COALESCE(AVG(duration_ms), 0) FROM commands WHERE duration_ms > 0`).
		Scan(&s.AvgDurationMs); err != nil {
		return s, fmt.Errorf("stats avg duration: %w", err)
	}

	// Top 10 commands by frequency
	rows, err := db.Query(`
		SELECT command, COUNT(*) as cnt
		FROM commands
		WHERE command NOT IN ('ls','ll','pwd','clear','cd','echo','cat','man','history')
		  AND command != '[redacted — command contained potential secret]'
		GROUP BY command
		ORDER BY cnt DESC
		LIMIT 10`)
	if err != nil {
		return s, fmt.Errorf("stats top commands: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cf CommandFreq
		if err := rows.Scan(&cf.Command, &cf.Count); err != nil {
			return s, err
		}
		s.TopCommands = append(s.TopCommands, cf)
	}

	// Top 10 repos by command count
	repoRows, err := db.Query(`
		SELECT git_repo, COUNT(*) as cnt
		FROM commands
		WHERE git_repo != ''
		GROUP BY git_repo
		ORDER BY cnt DESC
		LIMIT 10`)
	if err != nil {
		return s, fmt.Errorf("stats top repos: %w", err)
	}
	defer repoRows.Close()
	for repoRows.Next() {
		var rf RepoFreq
		if err := repoRows.Scan(&rf.Repo, &rf.Count); err != nil {
			return s, err
		}
		s.TopRepos = append(s.TopRepos, rf)
	}

	return s, nil
}
