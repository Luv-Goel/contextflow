package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/Luv-Goel/contextflow/internal/db"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("86"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	exitStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	repoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
)

// SearchModel is the Bubble Tea model for the search TUI.
type SearchModel struct {
	input      textinput.Model
	commands   []db.Command
	filtered   []db.Command
	cursor     int
	selected   string
	quitting   bool
	printOnly  bool
	width      int
	height     int
}

// NewSearchModel creates a search model pre-loaded with commands.
func NewSearchModel(commands []db.Command, printOnly bool) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search commands..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	m := SearchModel{
		input:     ti,
		commands:  commands,
		filtered:  commands,
		printOnly: printOnly,
	}
	return m
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if len(m.filtered) > 0 {
				m.selected = m.filtered[m.cursor].Command
			}
			return m, tea.Quit

		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown, tea.KeyCtrlN:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filtered = filterCommands(m.commands, m.input.Value())
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m, cmd
}

func (m SearchModel) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder

	// Header
	sb.WriteString(titleStyle.Render("  ContextFlow ") + dimStyle.Render(" ctrl+r search  ↑/↓ navigate  enter select  esc quit") + "\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  %d/%d commands", len(m.filtered), len(m.commands))) + "\n\n")

	// Input
	sb.WriteString("  " + m.input.View() + "\n\n")

	// Results — show up to (height - 8) items
	maxItems := 10
	if m.height > 0 {
		maxItems = m.height - 8
		if maxItems < 5 {
			maxItems = 5
		}
	}

	start := 0
	if m.cursor >= maxItems {
		start = m.cursor - maxItems + 1
	}
	end := start + maxItems
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		cmd := m.filtered[i]
		line := fmt.Sprintf("  %-60s", truncate(cmd.Command, 60))
		meta := ""
		if cmd.GitRepo != "" {
			repo := repoName(cmd.GitRepo)
			meta = repoStyle.Render(" " + repo)
		} else if cmd.Directory != "" {
			meta = dimStyle.Render(" " + shortDir(cmd.Directory))
		}

		if i == m.cursor {
			sb.WriteString(selectedStyle.Render(line) + meta + "\n")
		} else {
			sb.WriteString(line + meta + "\n")
		}
	}

	if len(m.filtered) == 0 {
		sb.WriteString(dimStyle.Render("  No results.\n"))
	}

	sb.WriteString("\n" + exitStyle.Render("  esc") + dimStyle.Render(" to cancel"))
	return sb.String()
}

// Selected returns the command the user picked (or "" if cancelled).
func (m SearchModel) Selected() string {
	return m.selected
}

// filterCommands applies fuzzy filtering to a command list.
func filterCommands(commands []db.Command, query string) []db.Command {
	if query == "" {
		return commands
	}
	strs := make([]string, len(commands))
	for i, c := range commands {
		strs[i] = c.Command
	}
	matches := fuzzy.Find(query, strs)
	result := make([]db.Command, 0, len(matches))
	for _, m := range matches {
		result = append(result, commands[m.Index])
	}
	return result
}

func repoName(remote string) string {
	parts := strings.Split(remote, "/")
	if len(parts) >= 2 {
		name := parts[len(parts)-1]
		return strings.TrimSuffix(name, ".git")
	}
	return remote
}

func shortDir(dir string) string {
	parts := strings.Split(dir, "/")
	if len(parts) > 2 {
		return "~/" + strings.Join(parts[len(parts)-2:], "/")
	}
	return dir
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
