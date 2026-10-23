package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Command represents a recorded shell command.
type Command struct {
	ID         int64
	Command    string
	Directory  string
	GitRepo    string
	GitBranch  string
	ExitCode   int
	DurationMs int64
	SessionID  string
	Hostname   string
	RecordedAt time.Time
}

// RecordCommand inserts a new command into the database.
func (db *DB) RecordCommand(c Command) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO commands
			(command, directory, git_repo, git_branch, exit_code, duration_ms, session_id, hostname, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Command, c.Directory, c.GitRepo, c.GitBranch,
		c.ExitCode, c.DurationMs, c.SessionID, c.Hostname,
		c.RecordedAt.Unix(),
	)
	if err != nil {
		return 0, fmt.Errorf("record command: %w", err)
	}
	return res.LastInsertId()
}

// SearchCommands returns commands matching a query string (fuzzy via LIKE).
func (db *DB) SearchCommands(query string, limit int) ([]Command, error) {
	rows, err := db.Query(`
		SELECT id, command, directory, git_repo, git_branch,
		       exit_code, duration_ms, session_id, hostname, recorded_at
		FROM commands
		WHERE command LIKE ?
		ORDER BY recorded_at DESC, id DESC
		LIMIT ?`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search commands: %w", err)
	}
	defer rows.Close()
	return scanCommands(rows)
}

// RecentCommands returns the most recent commands, optionally filtered by git repo.
func (db *DB) RecentCommands(gitRepo string, limit int) ([]Command, error) {
	var rows *sql.Rows
	var err error
	if gitRepo != "" {
		rows, err = db.Query(`
			SELECT id, command, directory, git_repo, git_branch,
			       exit_code, duration_ms, session_id, hostname, recorded_at
			FROM commands
			WHERE git_repo = ?
			ORDER BY recorded_at DESC, id DESC
			LIMIT ?`, gitRepo, limit)
	} else {
		rows, err = db.Query(`
			SELECT id, command, directory, git_repo, git_branch,
			       exit_code, duration_ms, session_id, hostname, recorded_at
			FROM commands
			ORDER BY recorded_at DESC, id DESC
			LIMIT ?`, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("recent commands: %w", err)
	}
	defer rows.Close()
	return scanCommands(rows)
}

func scanCommands(rows *sql.Rows) ([]Command, error) {
	var cmds []Command
	for rows.Next() {
		var c Command
		var ts int64
		if err := rows.Scan(
			&c.ID, &c.Command, &c.Directory, &c.GitRepo, &c.GitBranch,
			&c.ExitCode, &c.DurationMs, &c.SessionID, &c.Hostname, &ts,
		); err != nil {
			return nil, err
		}
		c.RecordedAt = time.Unix(ts, 0)
		cmds = append(cmds, c)
	}
	return cmds, rows.Err()
}

// Workflow represents a grouped set of commands.
type Workflow struct {
	ID        int64
	Name      string
	GitRepo   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Commands  []Command
}

// Duration returns the total time spent on this workflow's commands,
// computed as the sum of each command's duration_ms. Falls back to the
// wall-clock span (last recorded_at minus first) when durations are absent.
func (w *Workflow) Duration() time.Duration {
	var total int64
	for _, c := range w.Commands {
		total += c.DurationMs
	}
	if total > 0 {
		return time.Duration(total) * time.Millisecond
	}
	if len(w.Commands) > 1 {
		return w.Commands[len(w.Commands)-1].RecordedAt.Sub(w.Commands[0].RecordedAt)
	}
	return 0
}

// CommandCount returns the number of commands in this workflow.
func (w *Workflow) CommandCount() int {
	return len(w.Commands)
}

// SuccessCount returns the number of commands that exited with code 0.
func (w *Workflow) SuccessCount() int {
	var n int
	for _, c := range w.Commands {
		if c.ExitCode == 0 {
			n++
		}
	}
	return n
}

// FailureCount returns the number of commands that exited non-zero.
func (w *Workflow) FailureCount() int {
	var n int
	for _, c := range w.Commands {
		if c.ExitCode != 0 {
			n++
		}
	}
	return n
}

// ListWorkflows returns all workflows with their commands.
func (db *DB) ListWorkflows(limit int) ([]Workflow, error) {
	rows, err := db.Query(`
		SELECT id, name, git_repo, created_at, updated_at
		FROM workflows
		ORDER BY updated_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var w Workflow
		var name sql.NullString
		var createdAt, updatedAt int64
		if err := rows.Scan(&w.ID, &name, &w.GitRepo, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		w.Name = name.String
		w.CreatedAt = time.Unix(createdAt, 0)
		w.UpdatedAt = time.Unix(updatedAt, 0)

		cmds, err := db.workflowCommands(w.ID)
		if err != nil {
			return nil, err
		}
		w.Commands = cmds
		workflows = append(workflows, w)
	}
	return workflows, rows.Err()
}

func (db *DB) workflowCommands(workflowID int64) ([]Command, error) {
	rows, err := db.Query(`
		SELECT c.id, c.command, c.directory, c.git_repo, c.git_branch,
		       c.exit_code, c.duration_ms, c.session_id, c.hostname, c.recorded_at
		FROM commands c
		JOIN workflow_commands wc ON wc.command_id = c.id
		WHERE wc.workflow_id = ?
		ORDER BY wc.position ASC`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCommands(rows)
}

// SaveWorkflow persists a detected workflow.
func (db *DB) SaveWorkflow(w Workflow) (int64, error) {
	now := time.Now().Unix()
	res, err := db.Exec(`
		INSERT INTO workflows (name, git_repo, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		nullString(w.Name), w.GitRepo, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("save workflow: %w", err)
	}
	wID, _ := res.LastInsertId()
	for i, cmd := range w.Commands {
		if _, err := db.Exec(`
			INSERT INTO workflow_commands (workflow_id, command_id, position)
			VALUES (?, ?, ?)`, wID, cmd.ID, i); err != nil {
			return 0, fmt.Errorf("save workflow command: %w", err)
		}
	}
	return wID, nil
}

// GetCommandsSince returns commands recorded within the given duration.
func (db *DB) GetCommandsSince(d time.Duration) ([]Command, error) {
	since := time.Now().Add(-d).Unix()
	rows, err := db.Query(`
		SELECT id, command, directory, git_repo, git_branch,
		       exit_code, duration_ms, session_id, hostname, recorded_at
		FROM commands
		WHERE recorded_at >= ?
		ORDER BY recorded_at ASC`, since)
	if err != nil {
		return nil, fmt.Errorf("commands since: %w", err)
	}
	defer rows.Close()
	return scanCommands(rows)
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// UpdateWorkflowName updates a workflow's name.
func (db *DB) UpdateWorkflowName(id int64, name string) error {
	_, err := db.Exec(`
		UPDATE workflows SET name = ?, updated_at = ?
		WHERE id = ?`, name, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("update workflow name: %w", err)
	}
	return nil
}

// DeleteWorkflow deletes a workflow and its commands.
func (db *DB) DeleteWorkflow(id int64) error {
	_, err := db.Exec(`DELETE FROM workflows WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	return nil
}

// GetWorkflowByID returns a workflow by ID with its commands.
func (db *DB) GetWorkflowByID(id int64) (Workflow, error) {
	var w Workflow
	var name sql.NullString
	var createdAt, updatedAt int64
	err := db.QueryRow(`
		SELECT id, name, git_repo, created_at, updated_at
		FROM workflows WHERE id = ?`, id).
		Scan(&w.ID, &name, &w.GitRepo, &createdAt, &updatedAt)
	if err != nil {
		return w, fmt.Errorf("get workflow: %w", err)
	}
	w.Name = name.String
	w.CreatedAt = time.Unix(createdAt, 0)
	w.UpdatedAt = time.Unix(updatedAt, 0)

	// Get commands
	rows, err := db.Query(`
		SELECT c.id, c.command, c.directory, c.git_repo, c.git_branch,
		       c.exit_code, c.duration_ms, c.session_id, c.hostname, c.recorded_at
		FROM workflow_commands wc
		JOIN commands c ON wc.command_id = c.id
		WHERE wc.workflow_id = ?
		ORDER BY wc.position`, id)
	if err != nil {
		return w, fmt.Errorf("get workflow commands: %w", err)
	}
	defer rows.Close()
	w.Commands, err = scanCommands(rows)
	return w, err
}
