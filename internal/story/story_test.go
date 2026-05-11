package story

import (
	"strings"
	"testing"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

// fakeDB implements the command provider for testing
type fakeDB struct {
	commands []db.Command
	err      error
}

func (f *fakeDB) GetCommandsSince(since time.Duration) ([]db.Command, error) {
	return f.commands, f.err
}

func TestGenerate_NoCommands(t *testing.T) {
	result := Generate(&fakeDB{}, 24*time.Hour)
	if !strings.Contains(result, "No commands recorded") {
		t.Errorf("expected 'No commands recorded', got: %s", result)
	}
}

func TestGenerate_WithCommands(t *testing.T) {
	fdb := &fakeDB{
		commands: []db.Command{
			{Command: "git push origin main", GitRepo: "https://github.com/user/repo", ExitCode: 0, RecordedAt: time.Now().Add(-1 * time.Hour)},
			{Command: "npm run build", GitRepo: "https://github.com/user/repo", ExitCode: 0, RecordedAt: time.Now().Add(-30 * time.Minute)},
			{Command: "docker ps", ExitCode: 0, RecordedAt: time.Now().Add(-15 * time.Minute)},
		},
	}
	result := Generate(fdb, 24*time.Hour)
	if strings.Contains(result, "No commands recorded") {
		t.Errorf("expected narrative, got 'no commands': %s", result)
	}
	if !strings.Contains(result, "3 commands") {
		t.Errorf("expected '3 commands' in output, got: %s", result)
	}
	if !strings.Contains(result, "git") {
		t.Errorf("expected git mention in narrative, got: %s", result)
	}
}

func TestGenerate_GitDominance(t *testing.T) {
	cmds := make([]db.Command, 15)
	base := time.Now().Add(-2 * time.Hour)
	for i := 0; i < 15; i++ {
		cmds[i] = db.Command{
			Command:    "git status",
			GitRepo:    "https://github.com/user/repo",
			RecordedAt: base.Add(time.Duration(i) * 5 * time.Minute),
		}
	}
	fdb := &fakeDB{commands: cmds}
	result := Generate(fdb, 24*time.Hour)
	if !strings.Contains(result, "Git was your companion") {
		t.Errorf("expected 'Git was your companion' verdict, got: %s", result)
	}
}

func TestGenerate_RegretsDetected(t *testing.T) {
	fdb := &fakeDB{
		commands: []db.Command{
			{Command: "rm -rf node_modules", RecordedAt: time.Now().Add(-1 * time.Hour)},
			{Command: "git reset --hard HEAD~1", RecordedAt: time.Now().Add(-30 * time.Minute)},
			{Command: "echo oops", RecordedAt: time.Now().Add(-15 * time.Minute)},
		},
	}
	result := Generate(fdb, 24*time.Hour)
	if !strings.Contains(result, "deleted more than you committed") {
		t.Errorf("expected 'deleted more than you committed' verdict, got: %s", result)
	}
}

func TestGenerate_SuccessDetected(t *testing.T) {
	fdb := &fakeDB{
		commands: []db.Command{
			{Command: "git push origin main", RecordedAt: time.Now().Add(-2 * time.Hour)},
			{Command: "npm run deploy", RecordedAt: time.Now().Add(-1 * time.Hour)},
			{Command: "git push --tags", RecordedAt: time.Now().Add(-30 * time.Minute)},
			{Command: "gh release create v1.0", RecordedAt: time.Now().Add(-15 * time.Minute)},
		},
	}
	result := Generate(fdb, 24*time.Hour)
	if !strings.Contains(result, "You shipped things") {
		t.Errorf("expected 'You shipped things' verdict, got: %s", result)
	}
}

func TestGenerate_SearchesDetected(t *testing.T) {
	cmds := make([]db.Command, 8)
	base := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 8; i++ {
		cmds[i] = db.Command{
			Command:    "grep something",
			RecordedAt: base.Add(time.Duration(i) * time.Minute),
		}
	}
	fdb := &fakeDB{commands: cmds}
	result := Generate(fdb, 24*time.Hour)
	if !strings.Contains(result, "So many searches") {
		t.Errorf("expected 'So many searches' verdict, got: %s", result)
	}
}
