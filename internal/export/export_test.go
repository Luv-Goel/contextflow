package export

import (
	"strings"
	"testing"
	"time"

	"github.com/Luv-Goel/contextflow/internal/db"
)

func sampleWorkflow() db.Workflow {
	base := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	return db.Workflow{
		ID:        1,
		Name:      "docker setup",
		GitRepo:   "https://github.com/test/app",
		CreatedAt: base,
		UpdatedAt: base.Add(5 * time.Minute),
		Commands: []db.Command{
			{ID: 1, Command: "docker build -t myapp .", Directory: "/app", RecordedAt: base},
			{ID: 2, Command: "docker run -p 8080:8080 myapp", Directory: "/app", RecordedAt: base.Add(2 * time.Minute)},
			{ID: 3, Command: "curl localhost:8080/health", Directory: "/app", RecordedAt: base.Add(3 * time.Minute)},
		},
	}
}

func TestToShellScript_HasShebang(t *testing.T) {
	out := ToShellScript(sampleWorkflow())
	if !strings.HasPrefix(out, "#!/usr/bin/env bash") {
		t.Errorf("expected shebang, got: %s", out[:40])
	}
}

func TestToShellScript_ContainsAllCommands(t *testing.T) {
	out := ToShellScript(sampleWorkflow())
	for _, cmd := range []string{"docker build", "docker run", "curl localhost"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected output to contain %q", cmd)
		}
	}
}

func TestToShellScript_ContainsWorkflowName(t *testing.T) {
	out := ToShellScript(sampleWorkflow())
	if !strings.Contains(out, "docker setup") {
		t.Error("expected output to contain workflow name")
	}
}

func TestToShellScript_SetEUO(t *testing.T) {
	out := ToShellScript(sampleWorkflow())
	if !strings.Contains(out, "set -euo pipefail") {
		t.Error("expected set -euo pipefail in shell script")
	}
}

func TestToShellScript_UnnamedWorkflow(t *testing.T) {
	w := sampleWorkflow()
	w.Name = ""
	out := ToShellScript(w)
	if !strings.Contains(out, "workflow-1") {
		t.Errorf("expected 'workflow-1' for unnamed workflow, got: %s", out[:200])
	}
}

func TestToMarkdown_HasTitle(t *testing.T) {
	out := ToMarkdown(sampleWorkflow())
	if !strings.HasPrefix(out, "# docker setup") {
		t.Errorf("expected markdown title, got: %s", out[:40])
	}
}

func TestToMarkdown_ContainsCodeBlocks(t *testing.T) {
	out := ToMarkdown(sampleWorkflow())
	count := strings.Count(out, "```bash")
	if count != 3 {
		t.Errorf("expected 3 code blocks, got %d", count)
	}
}

func TestToMarkdown_ContainsAllCommands(t *testing.T) {
	out := ToMarkdown(sampleWorkflow())
	for _, cmd := range []string{"docker build", "docker run", "curl localhost"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected markdown to contain %q", cmd)
		}
	}
}

func TestToMarkdown_ContainsRepo(t *testing.T) {
	out := ToMarkdown(sampleWorkflow())
	if !strings.Contains(out, "github.com/test/app") {
		t.Error("expected markdown to contain repo URL")
	}
}

func TestToMarkdown_ContainsSteps(t *testing.T) {
	out := ToMarkdown(sampleWorkflow())
	for _, step := range []string{"### Step 1", "### Step 2", "### Step 3"} {
		if !strings.Contains(out, step) {
			t.Errorf("expected markdown to contain %q", step)
		}
	}
}
