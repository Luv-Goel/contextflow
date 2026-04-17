package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Luv-Goel/contextflow/internal/db"
)

func sampleWorkflows() []db.Workflow {
	now := time.Now()
	cmds := []db.Command{
		{ID: 1, Command: "git fetch", RecordedAt: now},
		{ID: 2, Command: "git rebase origin/main", RecordedAt: now},
		{ID: 3, Command: "go build ./...", RecordedAt: now},
	}
	return []db.Workflow{
		{
			ID:        1,
			Name:      "deploy",
			GitRepo:   "https://github.com/org/repo.git",
			CreatedAt: now,
			UpdatedAt: now,
			Commands:  cmds,
		},
		{
			ID:        2,
			Name:      "test-run",
			GitRepo:   "",
			CreatedAt: now,
			UpdatedAt: now,
			Commands:  cmds[:2],
		},
		{
			ID:        3,
			Name:      "build",
			GitRepo:   "",
			CreatedAt: now,
			UpdatedAt: now,
			Commands:  cmds[2:],
		},
	}
}

func TestNewWorkflowsModel_EmptyList(t *testing.T) {
	m := NewWorkflowsModel([]db.Workflow{})

	if len(m.workflows) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(m.workflows))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.cursor)
	}
	if m.selected != nil {
		t.Error("expected selected=nil initially")
	}
	if m.quitting {
		t.Error("expected quitting=false initially")
	}
}

func TestNewWorkflowsModel_WithWorkflows(t *testing.T) {
	wfs := sampleWorkflows()
	m := NewWorkflowsModel(wfs)

	if len(m.workflows) != len(wfs) {
		t.Errorf("expected %d workflows, got %d", len(wfs), len(m.workflows))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.cursor)
	}
}

func TestWorkflowsModel_DownMovesCursor(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(WorkflowsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after Down, got %d", m.cursor)
	}
}

func TestWorkflowsModel_UpAtTopStaysZero(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(WorkflowsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay 0 when pressing Up at top, got %d", m.cursor)
	}
}

func TestWorkflowsModel_DownAndUpNavigation(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	// Down twice
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(WorkflowsModel)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(WorkflowsModel)

	if m.cursor != 2 {
		t.Fatalf("expected cursor=2, got %d", m.cursor)
	}

	// Up once
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(WorkflowsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after Up, got %d", m.cursor)
	}
}

func TestWorkflowsModel_DownClampedAtBottom(t *testing.T) {
	wfs := sampleWorkflows()
	m := NewWorkflowsModel(wfs)

	for i := 0; i < len(wfs)+10; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(WorkflowsModel)
	}

	if m.cursor != len(wfs)-1 {
		t.Errorf("expected cursor clamped at %d, got %d", len(wfs)-1, m.cursor)
	}
}

func TestWorkflowsModel_EscapeSetsQuitting(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(WorkflowsModel)

	if !m.quitting {
		t.Error("expected quitting=true after Escape")
	}
}

func TestWorkflowsModel_CtrlCSetsQuitting(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = result.(WorkflowsModel)

	if !m.quitting {
		t.Error("expected quitting=true after Ctrl+C")
	}
}

func TestWorkflowsModel_EnterSetsSelected(t *testing.T) {
	wfs := sampleWorkflows()
	m := NewWorkflowsModel(wfs)

	// cursor is at 0
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(WorkflowsModel)

	if m.selected == nil {
		t.Fatal("expected selected to be non-nil after Enter")
	}
	if m.selected.ID != wfs[0].ID {
		t.Errorf("expected selected.ID=%d, got %d", wfs[0].ID, m.selected.ID)
	}
	if m.selected.Name != wfs[0].Name {
		t.Errorf("expected selected.Name=%q, got %q", wfs[0].Name, m.selected.Name)
	}
}

func TestWorkflowsModel_EnterOnEmptyList(t *testing.T) {
	m := NewWorkflowsModel([]db.Workflow{})

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(WorkflowsModel)

	// Should not panic; selected stays nil
	if m.selected != nil {
		t.Error("expected selected=nil when pressing Enter on empty list")
	}
}

func TestWorkflowsModel_EnterSelectsAtCursor(t *testing.T) {
	wfs := sampleWorkflows()
	m := NewWorkflowsModel(wfs)

	// Move to second workflow
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(WorkflowsModel)

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(WorkflowsModel)

	if m.selected == nil {
		t.Fatal("expected selected to be non-nil")
	}
	if m.selected.ID != wfs[1].ID {
		t.Errorf("expected selected workflow ID=%d, got %d", wfs[1].ID, m.selected.ID)
	}
}

func TestWorkflowsModel_Selected(t *testing.T) {
	wfs := sampleWorkflows()
	m := NewWorkflowsModel(wfs)

	if m.Selected() != nil {
		t.Error("expected Selected()=nil initially")
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(WorkflowsModel)

	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected non-nil Selected() after Enter")
	}
	if sel.ID != wfs[0].ID {
		t.Errorf("expected Selected().ID=%d, got %d", wfs[0].ID, sel.ID)
	}
}

func TestWorkflowsModel_CtrlNMovesDown(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = result.(WorkflowsModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after Ctrl+N, got %d", m.cursor)
	}
}

func TestWorkflowsModel_CtrlPMovesUp(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	// Move down first
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(WorkflowsModel)

	// Ctrl+P moves up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = result.(WorkflowsModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after Ctrl+P, got %d", m.cursor)
	}
}

func TestWorkflowsModel_WindowSizeMsg(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(WorkflowsModel)

	if m.width != 120 {
		t.Errorf("expected width=120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height=40, got %d", m.height)
	}
}

func TestWorkflowsModel_ViewEmptyList(t *testing.T) {
	m := NewWorkflowsModel([]db.Workflow{})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty View() even for empty workflow list")
	}
}

func TestWorkflowsModel_ViewWithWorkflows(t *testing.T) {
	m := NewWorkflowsModel(sampleWorkflows())
	view := m.View()
	if view == "" {
		t.Error("expected non-empty View() with workflows")
	}
}
