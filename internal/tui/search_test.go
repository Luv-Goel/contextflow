package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Luv-Goel/contextflow/internal/db"
)

func sampleCommands() []db.Command {
	now := time.Now()
	return []db.Command{
		{ID: 1, Command: "git status", Directory: "/repo", RecordedAt: now},
		{ID: 2, Command: "git diff", Directory: "/repo", RecordedAt: now},
		{ID: 3, Command: "go test ./...", Directory: "/repo", RecordedAt: now},
		{ID: 4, Command: "docker ps", Directory: "/home", RecordedAt: now},
		{ID: 5, Command: "ls -la", Directory: "/home", RecordedAt: now},
	}
}

func TestNewSearchModel_InitialState(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	if len(m.commands) != len(cmds) {
		t.Errorf("expected %d commands, got %d", len(cmds), len(m.commands))
	}
	if len(m.filtered) != len(cmds) {
		t.Errorf("expected %d filtered commands initially, got %d", len(cmds), len(m.filtered))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.cursor)
	}
	if m.selected != "" {
		t.Errorf("expected empty selected, got %q", m.selected)
	}
	if m.quitting {
		t.Error("expected quitting=false initially")
	}
	if m.printOnly != false {
		t.Error("expected printOnly=false")
	}
}

func TestNewSearchModel_PrintOnly(t *testing.T) {
	m := NewSearchModel(sampleCommands(), true)
	if !m.printOnly {
		t.Error("expected printOnly=true")
	}
}

func TestSearchModel_DownMovescursor(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)
	initial := m.cursor // 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := m.Update(msg)
	m = result.(SearchModel)

	if m.cursor != initial+1 {
		t.Errorf("expected cursor %d after Down, got %d", initial+1, m.cursor)
	}
}

func TestSearchModel_DownMultipleTimes(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	for i := 0; i < 3; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(SearchModel)
	}

	if m.cursor != 3 {
		t.Errorf("expected cursor=3 after 3 downs, got %d", m.cursor)
	}
}

func TestSearchModel_UpAtTopStaysZero(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)

	// cursor is already 0, press Up
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(SearchModel)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay 0 when pressing Up at top, got %d", m.cursor)
	}
}

func TestSearchModel_DownClampedAtBottom(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	// Press Down more times than there are items
	for i := 0; i < len(cmds)+5; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(SearchModel)
	}

	if m.cursor != len(cmds)-1 {
		t.Errorf("expected cursor clamped at %d, got %d", len(cmds)-1, m.cursor)
	}
}

func TestSearchModel_EscapeSetsQuitting(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(SearchModel)

	if !m.quitting {
		t.Error("expected quitting=true after Escape")
	}
}

func TestSearchModel_CtrlCSetsQuitting(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = result.(SearchModel)

	if !m.quitting {
		t.Error("expected quitting=true after Ctrl+C")
	}
}

func TestSearchModel_EnterSelectsCurrent(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	// Move down once so cursor=1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(SearchModel)

	// Press Enter
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SearchModel)

	if m.selected != cmds[1].Command {
		t.Errorf("expected selected=%q, got %q", cmds[1].Command, m.selected)
	}
}

func TestSearchModel_EnterOnEmptyList(t *testing.T) {
	m := NewSearchModel([]db.Command{}, false)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SearchModel)

	// Should not panic, selected stays empty
	if m.selected != "" {
		t.Errorf("expected empty selected on empty list, got %q", m.selected)
	}
}

func TestSearchModel_TypingFiltersCommands(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	// Simulate typing "git" — feed individual rune key messages
	for _, ch := range "git" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(SearchModel)
	}

	// After typing "git", filtered should only contain commands matching "git"
	// fuzzy match: "git status", "git diff" definitely match; others may not
	if len(m.filtered) == 0 {
		t.Error("expected some filtered results after typing 'git'")
	}
	for _, c := range m.filtered {
		// All filtered commands should be fuzzy-matchable with "git"
		_ = c // just ensure no panic iterating
	}
}

func TestSearchModel_FilterReducesResults(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	// Type something very specific
	for _, ch := range "docker" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(SearchModel)
	}

	if len(m.filtered) >= len(cmds) {
		t.Errorf("expected filtered to be smaller than full list after typing 'docker', got %d/%d", len(m.filtered), len(cmds))
	}
}

func TestSearchModel_CursorClampedAfterFilter(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	// Move cursor to bottom
	for i := 0; i < len(cmds)-1; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(SearchModel)
	}

	// Now type something that produces fewer results
	for _, ch := range "docker" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(SearchModel)
	}

	// cursor should be clamped within filtered length
	if len(m.filtered) > 0 && m.cursor >= len(m.filtered) {
		t.Errorf("cursor %d out of bounds for filtered list of length %d", m.cursor, len(m.filtered))
	}
}

func TestSearchModel_Selected(t *testing.T) {
	cmds := sampleCommands()
	m := NewSearchModel(cmds, false)

	if m.Selected() != "" {
		t.Errorf("expected empty Selected() initially, got %q", m.Selected())
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SearchModel)

	if m.Selected() != cmds[0].Command {
		t.Errorf("expected Selected()=%q, got %q", cmds[0].Command, m.Selected())
	}
}

func TestSearchModel_CtrlPMovesUp(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)

	// Move down first
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(SearchModel)
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after Down, got %d", m.cursor)
	}

	// Ctrl+P should move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = result.(SearchModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after Ctrl+P, got %d", m.cursor)
	}
}

func TestSearchModel_CtrlNMovesDown(t *testing.T) {
	m := NewSearchModel(sampleCommands(), false)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = result.(SearchModel)

	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after Ctrl+N, got %d", m.cursor)
	}
}
