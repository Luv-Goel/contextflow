package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Luv-Goel/contextflow/internal/db"
)

var (
	workflowTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	workflowHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	workflowDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	workflowSelStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("86"))
	cmdStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cmdDimStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// WorkflowsModel is the Bubble Tea model for browsing workflows.
type WorkflowsModel struct {
	workflows  []db.Workflow
	cursor     int
	selected   *db.Workflow
	quitting   bool
	width      int
	height     int
}

// NewWorkflowsModel creates a workflow browser model.
func NewWorkflowsModel(workflows []db.Workflow) WorkflowsModel {
	return WorkflowsModel{workflows: workflows}
}

func (m WorkflowsModel) Init() tea.Cmd { return nil }

func (m WorkflowsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.workflows) > 0 {
				w := m.workflows[m.cursor]
				m.selected = &w
			}
			return m, tea.Quit

		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown, tea.KeyCtrlN:
			if m.cursor < len(m.workflows)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m WorkflowsModel) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(workflowTitleStyle.Render("  ContextFlow ") + workflowDimStyle.Render(" workflows  ↑/↓ navigate  enter replay  esc quit") + "\n")
	sb.WriteString(workflowDimStyle.Render(fmt.Sprintf("  %d workflows detected\n\n", len(m.workflows))))

	if len(m.workflows) == 0 {
		sb.WriteString(workflowDimStyle.Render("  No workflows yet. Run some commands and come back!\n"))
		return sb.String()
	}

	maxItems := 8
	if m.height > 0 {
		maxItems = (m.height - 6) / 5
		if maxItems < 3 {
			maxItems = 3
		}
	}

	start := 0
	if m.cursor >= maxItems {
		start = m.cursor - maxItems + 1
	}
	end := start + maxItems
	if end > len(m.workflows) {
		end = len(m.workflows)
	}

	for i := start; i < end; i++ {
		w := m.workflows[i]
		name := w.Name
		if name == "" {
			name = fmt.Sprintf("workflow #%d", w.ID)
		}

		header := fmt.Sprintf("  📋 %s", name)
		if w.GitRepo != "" {
			header += workflowDimStyle.Render("  " + repoName(w.GitRepo))
		}
		header += workflowDimStyle.Render(fmt.Sprintf("  (%d cmds, %s)", len(w.Commands), w.UpdatedAt.Format("Jan 2")))

		if i == m.cursor {
			sb.WriteString(workflowSelStyle.Render(header) + "\n")
		} else {
			sb.WriteString(workflowHeaderStyle.Render(header) + "\n")
		}

		// Show first 3 commands as preview
		previewCount := 3
		if len(w.Commands) < previewCount {
			previewCount = len(w.Commands)
		}
		for j := 0; j < previewCount; j++ {
			sb.WriteString(cmdDimStyle.Render("    $ ") + cmdStyle.Render(truncate(w.Commands[j].Command, 70)) + "\n")
		}
		if len(w.Commands) > previewCount {
			sb.WriteString(cmdDimStyle.Render(fmt.Sprintf("    ... +%d more\n", len(w.Commands)-previewCount)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Selected returns the workflow the user chose (or nil if cancelled).
func (m WorkflowsModel) Selected() *db.Workflow {
	return m.selected
}
