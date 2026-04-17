package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS commands (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	command     TEXT    NOT NULL,
	directory   TEXT,
	git_repo    TEXT,
	git_branch  TEXT,
	exit_code   INTEGER DEFAULT 0,
	duration_ms INTEGER DEFAULT 0,
	session_id  TEXT,
	hostname    TEXT,
	recorded_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_commands_recorded_at ON commands(recorded_at);
CREATE INDEX IF NOT EXISTS idx_commands_git_repo    ON commands(git_repo);
CREATE INDEX IF NOT EXISTS idx_commands_session_id  ON commands(session_id);

CREATE TABLE IF NOT EXISTS workflows (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT,
	git_repo   TEXT,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_commands (
	workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
	command_id  INTEGER NOT NULL REFERENCES commands(id)  ON DELETE CASCADE,
	position    INTEGER NOT NULL,
	PRIMARY KEY (workflow_id, command_id)
);
`

// DB wraps a sql.DB with helpers.
type DB struct {
	*sql.DB
}

// DataDir returns the ContextFlow data directory (~/.contextflow).
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".contextflow")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create data directory: %w", err)
	}
	return dir, nil
}

// Open opens (or creates) the SQLite database and applies the schema.
func Open() (*DB, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dir, "history.db")
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	if _, err := sqlDB.Exec(schema); err != nil {
		return nil, fmt.Errorf("cannot apply schema: %w", err)
	}
	return &DB{sqlDB}, nil
}
