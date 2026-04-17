package db

import (
	"os"
	"testing"
	"time"
)

func tempDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	os.Setenv("HOME", dir) // DataDir() uses UserHomeDir
	database, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestRecordAndSearch(t *testing.T) {
	db := tempDB(t)

	_, err := db.RecordCommand(Command{
		Command:    "git push origin main",
		Directory:  "/repo",
		GitRepo:    "https://github.com/x/y",
		SessionID:  "s1",
		RecordedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("RecordCommand: %v", err)
	}

	results, err := db.SearchCommands("git push", 10)
	if err != nil {
		t.Fatalf("SearchCommands: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Command != "git push origin main" {
		t.Errorf("unexpected command: %q", results[0].Command)
	}
}

func TestSearchReturnsEmpty(t *testing.T) {
	db := tempDB(t)
	results, err := db.SearchCommands("doesnotexist", 10)
	if err != nil {
		t.Fatalf("SearchCommands: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRecentCommands_Limit(t *testing.T) {
	db := tempDB(t)
	base := time.Now()
	for i := 0; i < 5; i++ {
		db.RecordCommand(Command{
			Command:    "echo hello",
			Directory:  "/tmp",
			SessionID:  "s1",
			RecordedAt: base.Add(time.Duration(i) * time.Second),
		})
	}
	results, err := db.RecentCommands("", 3)
	if err != nil {
		t.Fatalf("RecentCommands: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results (limit), got %d", len(results))
	}
}

func TestRecentCommands_FilterByRepo(t *testing.T) {
	db := tempDB(t)
	base := time.Now()
	db.RecordCommand(Command{Command: "git status", GitRepo: "https://github.com/x/y", SessionID: "s1", RecordedAt: base})
	db.RecordCommand(Command{Command: "docker ps", GitRepo: "https://github.com/a/b", SessionID: "s1", RecordedAt: base.Add(time.Second)})
	db.RecordCommand(Command{Command: "git log", GitRepo: "https://github.com/x/y", SessionID: "s1", RecordedAt: base.Add(2 * time.Second)})

	results, err := db.RecentCommands("https://github.com/x/y", 10)
	if err != nil {
		t.Fatalf("RecentCommands: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for repo filter, got %d", len(results))
	}
	for _, r := range results {
		if r.GitRepo != "https://github.com/x/y" {
			t.Errorf("unexpected repo: %q", r.GitRepo)
		}
	}
}

func TestSaveAndListWorkflows(t *testing.T) {
	db := tempDB(t)
	base := time.Now()

	// Record some commands first to get IDs
	id1, _ := db.RecordCommand(Command{Command: "npm install", Directory: "/app", SessionID: "s1", RecordedAt: base})
	id2, _ := db.RecordCommand(Command{Command: "npm run build", Directory: "/app", SessionID: "s1", RecordedAt: base.Add(time.Minute)})

	w := Workflow{
		Name:      "npm build",
		GitRepo:   "https://github.com/x/y",
		CreatedAt: base,
		UpdatedAt: base.Add(time.Minute),
		Commands: []Command{
			{ID: id1, Command: "npm install"},
			{ID: id2, Command: "npm run build"},
		},
	}

	wID, err := db.SaveWorkflow(w)
	if err != nil {
		t.Fatalf("SaveWorkflow: %v", err)
	}
	if wID == 0 {
		t.Fatal("expected non-zero workflow ID")
	}

	workflows, err := db.ListWorkflows(10)
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	if workflows[0].Name != "npm build" {
		t.Errorf("unexpected name: %q", workflows[0].Name)
	}
	if len(workflows[0].Commands) != 2 {
		t.Errorf("expected 2 commands in workflow, got %d", len(workflows[0].Commands))
	}
}

func TestGetStats(t *testing.T) {
	db := tempDB(t)
	base := time.Now()

	cmds := []string{"git push", "git push", "npm run build", "docker ps", "git push"}
	for i, cmd := range cmds {
		db.RecordCommand(Command{
			Command:    cmd,
			GitRepo:    "https://github.com/x/y",
			SessionID:  "s1",
			DurationMs: 100,
			RecordedAt: base.Add(time.Duration(i) * time.Second),
		})
	}

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalCommands != 5 {
		t.Errorf("expected 5 total commands, got %d", stats.TotalCommands)
	}
	if stats.UniqueCommands != 3 {
		t.Errorf("expected 3 unique commands, got %d", stats.UniqueCommands)
	}
	if len(stats.TopCommands) == 0 {
		t.Error("expected at least 1 top command")
	}
	if stats.TopCommands[0].Command != "git push" {
		t.Errorf("expected 'git push' as top command, got %q", stats.TopCommands[0].Command)
	}
	if stats.TopCommands[0].Count != 3 {
		t.Errorf("expected count 3 for git push, got %d", stats.TopCommands[0].Count)
	}
}
