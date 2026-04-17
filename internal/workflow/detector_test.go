package workflow

import (
	"testing"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

func makeCmd(id int64, cmd, dir, repo, session string, t time.Time, exit int) db.Command {
	return db.Command{
		ID:         id,
		Command:    cmd,
		Directory:  dir,
		GitRepo:    repo,
		SessionID:  session,
		RecordedAt: t,
		ExitCode:   exit,
	}
}

func TestDetect_EmptyInput(t *testing.T) {
	workflows := Detect(nil)
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows for nil input, got %d", len(workflows))
	}
	workflows = Detect([]db.Command{})
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows for empty input, got %d", len(workflows))
	}
}

func TestDetect_BelowMinSize(t *testing.T) {
	now := time.Now()
	cmds := []db.Command{
		makeCmd(1, "git status", "/repo", "https://github.com/x/y", "s1", now, 0),
		makeCmd(2, "git log", "/repo", "https://github.com/x/y", "s1", now.Add(1*time.Minute), 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows (below min size 3), got %d", len(workflows))
	}
}

func TestDetect_SingleWorkflow(t *testing.T) {
	base := time.Now()
	cmds := []db.Command{
		makeCmd(1, "git clone https://github.com/x/y", "/tmp", "https://github.com/x/y", "s1", base, 0),
		makeCmd(2, "cd y", "/tmp/y", "https://github.com/x/y", "s1", base.Add(1*time.Minute), 0),
		makeCmd(3, "npm install", "/tmp/y", "https://github.com/x/y", "s1", base.Add(2*time.Minute), 0),
		makeCmd(4, "npm run dev", "/tmp/y", "https://github.com/x/y", "s1", base.Add(3*time.Minute), 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	if len(workflows[0].Commands) != 4 {
		t.Errorf("expected 4 commands in workflow, got %d", len(workflows[0].Commands))
	}
}

func TestDetect_ChronologicalOrder(t *testing.T) {
	base := time.Now()
	// Commands passed in DESC order (as from DB)
	cmds := []db.Command{
		makeCmd(4, "npm run dev", "/repo", "https://github.com/x/y", "s1", base.Add(3*time.Minute), 0),
		makeCmd(3, "npm install", "/repo", "https://github.com/x/y", "s1", base.Add(2*time.Minute), 0),
		makeCmd(2, "git checkout main", "/repo", "https://github.com/x/y", "s1", base.Add(1*time.Minute), 0),
		makeCmd(1, "git clone ...", "/repo", "https://github.com/x/y", "s1", base, 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	w := workflows[0]
	// Should be in chronological order: id 1, 2, 3, 4
	for i := 0; i < len(w.Commands)-1; i++ {
		if w.Commands[i].ID > w.Commands[i+1].ID {
			t.Errorf("commands out of order at index %d: id %d before id %d",
				i, w.Commands[i].ID, w.Commands[i+1].ID)
		}
	}
}

func TestDetect_SeparateSessionsSplitWorkflows(t *testing.T) {
	base := time.Now()
	cmds := []db.Command{
		makeCmd(1, "docker build .", "/app", "https://github.com/x/y", "s1", base, 0),
		makeCmd(2, "docker run myapp", "/app", "https://github.com/x/y", "s1", base.Add(1*time.Minute), 0),
		makeCmd(3, "docker ps", "/app", "https://github.com/x/y", "s1", base.Add(2*time.Minute), 0),
		// Different session — should be a new workflow
		makeCmd(4, "git pull", "/app", "https://github.com/x/y", "s2", base.Add(3*time.Minute), 0),
		makeCmd(5, "npm test", "/app", "https://github.com/x/y", "s2", base.Add(4*time.Minute), 0),
		makeCmd(6, "npm run build", "/app", "https://github.com/x/y", "s2", base.Add(5*time.Minute), 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 2 {
		t.Fatalf("expected 2 workflows (different sessions), got %d", len(workflows))
	}
}

func TestDetect_TimeGapSplitsWorkflow(t *testing.T) {
	base := time.Now()
	cmds := []db.Command{
		makeCmd(1, "docker build .", "/app", "", "s1", base, 0),
		makeCmd(2, "docker run myapp", "/app", "", "s1", base.Add(1*time.Minute), 0),
		makeCmd(3, "docker ps", "/app", "", "s1", base.Add(2*time.Minute), 0),
		// 45-minute gap — should split
		makeCmd(4, "git status", "/app", "", "s1", base.Add(47*time.Minute), 0),
		makeCmd(5, "git pull", "/app", "", "s1", base.Add(48*time.Minute), 0),
		makeCmd(6, "git push", "/app", "", "s1", base.Add(49*time.Minute), 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 2 {
		t.Fatalf("expected 2 workflows (time gap), got %d", len(workflows))
	}
}

func TestDetect_SameTimetampUsesIDOrder(t *testing.T) {
	// All same timestamp — ID should determine order
	base := time.Now().Truncate(time.Second)
	cmds := []db.Command{
		makeCmd(3, "npm run dev", "/repo", "https://g.com/x/y", "s1", base, 0),
		makeCmd(1, "git clone ...", "/repo", "https://g.com/x/y", "s1", base, 0),
		makeCmd(4, "open localhost:3000", "/repo", "https://g.com/x/y", "s1", base, 0),
		makeCmd(2, "npm install", "/repo", "https://g.com/x/y", "s1", base, 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	ids := []int64{}
	for _, c := range workflows[0].Commands {
		ids = append(ids, c.ID)
	}
	for i := 0; i < len(ids)-1; i++ {
		if ids[i] > ids[i+1] {
			t.Errorf("commands out of ID order at index %d: %v", i, ids)
		}
	}
}

func TestAutoName_SkipsNoisyCommands(t *testing.T) {
	base := time.Now()
	cmds := []db.Command{
		makeCmd(1, "cd /app", "/app", "", "s1", base, 0),
		makeCmd(2, "ls -la", "/app", "", "s1", base.Add(10*time.Second), 0),
		makeCmd(3, "docker build -t myapp .", "/app", "", "s1", base.Add(20*time.Second), 0),
	}
	name := autoName(cmds)
	if name != "docker build" {
		t.Errorf("expected autoName to skip cd/ls and return 'docker build', got %q", name)
	}
}

func TestDetect_NoisyCommandsExcluded(t *testing.T) {
	base := time.Now()
	// Only 2 meaningful + 1 noise in window — should not form workflow
	cmds := []db.Command{
		makeCmd(1, "ls", "/tmp", "", "s1", base, 0),
		makeCmd(2, "pwd", "/tmp", "", "s1", base.Add(1*time.Minute), 0),
	}
	workflows := Detect(cmds)
	if len(workflows) != 0 {
		t.Errorf("expected no workflow from 2 noise commands, got %d", len(workflows))
	}
}
