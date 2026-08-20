package tui

import (
	"strings"

	"github.com/gravestench/dark-magic/internal/shell"
)

// refreshTranscript selects the active timeline, renders each semantic event,
// enforces the live line limit, and retains bottom-follow behavior.
func (m *Model) refreshTranscript() {
	events := m.session.TranscriptTimeline()
	if m.view == viewLogs {
		events = m.session.LogTimeline()
	}

	m.timelineRevision = m.session.TimelineRevision()
	lines := make([]string, 0, len(events)+1)

	if len(events) == 0 {
		lines = append(lines, m.emptyTimelineLine())
	}

	for _, event := range events {
		lines = append(lines, renderTimelineEvent(event)...)
	}

	if limit := m.session.Settings().Values().TranscriptLimit; len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	m.output.SetContent(strings.Join(lines, "\n"))
	m.output.GotoBottom()
}

// emptyTimelineLine distinguishes an unused Lua scope from an empty process-log
// buffer so operators know which source is currently selected.
func (m *Model) emptyTimelineLine() string {
	if m.view == viewLogs {
		return dimStyle.Render("No application logs have been captured.")
	}

	return dimStyle.Render("No commands have been evaluated in this scope.")
}

// renderTimelineEvent expands multi-line values while keeping commands, errors,
// logs, and MOTD text on their established semantic styles.
func renderTimelineEvent(event shell.TimelineEvent) []string {
	event.Text = strings.ReplaceAll(event.Text, "\t", "    ")

	switch event.Kind {
	case "motd":
		return []string{accentStyle.Render(event.Text)}
	case "command":
		return []string{accentStyle.Render("❯ ") + event.Text}
	case "value":
		return renderValueLines(event.Text)
	case "error", "log-error":
		return []string{errorStyle.Render(event.Text)}
	case "log-warn":
		return []string{warningStyle.Render(event.Text)}
	case "log-debug":
		return []string{dimStyle.Render(event.Text)}
	default:
		return []string{event.Text}
	}
}

// renderValueLines highlights Markdown-like headings and code indentation while
// leaving ordinary evaluator values on the result color.
func renderValueLines(text string) []string {
	parts := strings.Split(text, "\n")
	lines := make([]string, 0, len(parts))

	for _, line := range parts {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "#"):
			lines = append(lines, accentStyle.Render(line))
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(line, "  "):
			lines = append(lines, dimStyle.Render(line))
		default:
			lines = append(lines, valueStyle.Render(line))
		}
	}

	return lines
}
