package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#D69A2D")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#817A6D"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B5E"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3A7"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8B84A"))
	panelStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6E5228"))
)

// View composes target authority, tabs, transcript, optional Lua editor, and
// status into one terminal frame without mutating model state.
func (m Model) View() tea.View {
	header, capabilities := m.renderHeader()
	sections := []string{
		header,
		capabilities,
		renderTabs(m.view),
		panelStyle.Width(max(1, m.width-2)).Render(m.output.View()),
	}

	if m.view == viewLua {
		if candidates := m.renderCandidates(); candidates != "" {
			sections = append(sections, panelStyle.Width(max(1, m.width-2)).Render(candidates))
		}

		sections = append(sections, panelStyle.Width(max(1, m.width-2)).Render(m.input.View()))
	}

	sections = append(sections, m.renderStatus())
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return tea.NewView(content)
}

// renderHeader surfaces target, session, mutability, and capability scope so a
// restricted remote shell cannot be mistaken for a local developer shell.
func (m Model) renderHeader() (string, string) {
	policy := m.session.Policy()

	mode := "read-only"
	if policy.Mutable {
		mode = "mutable"
	}

	details := fmt.Sprintf(
		"target %s  session %s  policy %s (%s)",
		m.session.Target(),
		m.session.ID(),
		policy.Name,
		mode,
	)

	header := accentStyle.Render("DARK MAGIC SHELL") + "  " + dimStyle.Render(details)
	capabilities := dimStyle.Render("capabilities: " + strings.Join(policy.Capabilities, ", "))

	return header, capabilities
}

// renderStatus selects view-specific key hints and appends transient evaluation
// or completion feedback when present.
func (m Model) renderStatus() string {
	text := "F1 Lua  F2 Logs  Ctrl-S run  Enter newline  Tab complete  " +
		"arrows/PgUp/PgDn scroll  Ctrl-Q quit"
	if m.view == viewLogs {
		text = "F1 Lua  F2 Logs  arrows/PgUp/PgDn scroll  Ctrl-Q quit"
	}

	status := dimStyle.Render(text)
	if m.status != "" {
		status += "  " + accentStyle.Render(m.status)
	}

	return status
}

// renderCandidates presents at most five sorted Session candidates and marks the
// active cycling position without changing the candidate slice.
func (m Model) renderCandidates() string {
	if len(m.candidates) == 0 {
		return ""
	}

	limit := min(5, len(m.candidates))
	lines := make([]string, 0, limit)

	for index, candidate := range m.candidates[:limit] {
		line := fmt.Sprintf("  %-28s %s", candidate.Value, candidate.Detail)
		if index == m.candidateAt {
			line = accentStyle.Render("› " + strings.TrimSpace(line))
		} else {
			line = dimStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderTabs highlights the active view while preserving fixed F1/F2 labels.
func renderTabs(view viewMode) string {
	lua, logs := dimStyle.Render("[F1 LUA]"), dimStyle.Render("[F2 LOGS]")
	if view == viewLua {
		lua = accentStyle.Render("[F1 LUA]")
	} else {
		logs = accentStyle.Render("[F2 LOGS]")
	}

	return lua + "  " + logs
}
