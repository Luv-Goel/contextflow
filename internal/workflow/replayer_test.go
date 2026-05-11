package workflow

import (
	"testing"

	"github.com/Luv-Goel/contextflow/internal/db"
)

func TestReplay_EmptyWorkflow(t *testing.T) {
	w := db.Workflow{
		Name:     "test",
		Commands: []db.Command{},
	}
	err := Replay(w, DryRun)
	if err == nil {
		t.Fatal("expected error for empty workflow, got nil")
	}
	if err.Error() != `workflow "test" has no commands` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReplay_EmptyWorkflowNoName(t *testing.T) {
	w := db.Workflow{
		Commands: []db.Command{},
	}
	err := Replay(w, DryRun)
	if err == nil {
		t.Fatal("expected error for empty workflow, got nil")
	}
}

func TestReplay_DryRunDoesNotError(t *testing.T) {
	w := db.Workflow{
		Name: "dry run test",
		Commands: []db.Command{
			{Command: "echo hello", Directory: "/tmp"},
			{Command: "echo world", Directory: "/tmp"},
		},
	}
	// DryRun mode should not execute commands, so it should not error
	err := Replay(w, DryRun)
	if err != nil {
		t.Fatalf("Replay returned error: %v", err)
	}
}

func TestReplayModeValues(t *testing.T) {
	if Interactive != 0 {
		t.Errorf("expected Interactive to be 0, got %d", Interactive)
	}
	if DryRun != 1 {
		t.Errorf("expected DryRun to be 1, got %d", DryRun)
	}
}
