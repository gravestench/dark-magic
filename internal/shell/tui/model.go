// Package tui presents a shared shell session in a renderer-free terminal.
package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gravestench/dark-magic/internal/shell"
)

var (
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#D69A2D")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#817A6D"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B5E"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3A7"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8B84A"))
	panelStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6E5228"))
)

type evaluationMsg shell.Entry
type refreshMsg struct{}
type viewMode uint8

const (
	viewLua viewMode = iota
	viewLogs
)

// Model is the Charmbracelet adapter for a renderer-independent shell Session.
type Model struct {
	ctx              context.Context
	session          *shell.Session
	input            textarea.Model
	output           viewport.Model
	width            int
	height           int
	busy             bool
	history          int
	candidates       []shell.Candidate
	candidateAt      int
	status           string
	timelineRevision uint64
	view             viewMode
}

// NewModel prepares an interactive terminal model without starting a terminal.
func NewModel(session *shell.Session) Model {
	return newModel(context.Background(), session)
}

func newModel(ctx context.Context, session *shell.Session) Model {
	input := textarea.New()
	input.Placeholder = "Lua expression or statement"
	input.Prompt = "❯ "
	input.SetHeight(4)
	input.ShowLineNumbers = true
	input.Focus()

	output := viewport.New(viewport.WithWidth(76), viewport.WithHeight(12))
	output.SoftWrap = true
	output.FillHeight = true
	model := Model{ctx: ctx, session: session, input: input, output: output, history: len(session.History())}
	model.refreshTranscript()
	return model
}

func (m Model) Init() tea.Cmd { return tea.Batch(textarea.Blink, refreshLogs()) }

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.resize()
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "f1":
			m.view = viewLua
			m.input.Focus()
			m.resize()
			return m, nil
		case "f2":
			m.view = viewLogs
			m.input.Blur()
			m.resize()
			return m, nil
		case "ctrl+s", "alt+enter":
			if m.view != viewLua || m.busy || strings.TrimSpace(m.input.Value()) == "" {
				return m, nil
			}
			source := m.input.Value()
			m.busy, m.status = true, "evaluating"
			m.candidates = nil
			return m, func() tea.Msg { return evaluationMsg(m.session.Submit(m.ctx, source)) }
		case "tab", "shift+tab":
			if m.view != viewLua {
				break
			}
			m.complete(message.String() == "shift+tab")
			return m, nil
		case "alt+up":
			m.moveHistory(-1)
			return m, nil
		case "alt+down":
			m.moveHistory(1)
			return m, nil
		}
		m.candidates = nil
	case evaluationMsg:
		m.busy = false
		m.status = "ready"
		m.input.Reset()
		m.history = len(m.session.History())
		m.refreshTranscript()
		return m, nil
	case refreshMsg:
		if m.session.TimelineRevision() != m.timelineRevision {
			m.refreshTranscript()
		}
		return m, refreshLogs()
	}

	var commands []tea.Cmd
	var command tea.Cmd
	if m.view == viewLua {
		m.input, command = m.input.Update(message)
		commands = append(commands, command)
	}
	m.output, command = m.output.Update(message)
	commands = append(commands, command)
	return m, tea.Batch(commands...)
}

func (m Model) View() tea.View {
	policy := m.session.Policy()
	mode := "read-only"
	if policy.Mutable {
		mode = "mutable"
	}
	header := accentStyle.Render("DARK MAGIC SHELL") + "  " +
		dimStyle.Render(fmt.Sprintf("target %s  session %s  policy %s (%s)", m.session.Target(), m.session.ID(), policy.Name, mode))
	capabilities := dimStyle.Render("capabilities: " + strings.Join(policy.Capabilities, ", "))
	statusText := "F1 Lua  F2 Logs  Ctrl-S run  Enter newline  Tab complete  Ctrl-Q quit"
	if m.view == viewLogs {
		statusText = "F1 Lua  F2 Logs  arrows/PgUp/PgDn scroll  Ctrl-Q quit"
	}
	status := dimStyle.Render(statusText)
	if m.status != "" {
		status += "  " + accentStyle.Render(m.status)
	}
	sections := []string{header, capabilities, renderTabs(m.view), panelStyle.Width(max(1, m.width-2)).Render(m.output.View())}
	if m.view == viewLua {
		if candidates := m.renderCandidates(); candidates != "" {
			sections = append(sections, panelStyle.Width(max(1, m.width-2)).Render(candidates))
		}
		sections = append(sections, panelStyle.Width(max(1, m.width-2)).Render(m.input.View()))
	}
	sections = append(sections, status)
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tea.NewView(content)
}

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

func renderTabs(view viewMode) string {
	lua, logs := dimStyle.Render("[F1 LUA]"), dimStyle.Render("[F2 LOGS]")
	if view == viewLua {
		lua = accentStyle.Render("[F1 LUA]")
	} else {
		logs = accentStyle.Render("[F2 LOGS]")
	}
	return lua + "  " + logs
}

// Run owns the terminal lifecycle but not the supplied shell session.
func Run(ctx context.Context, session *shell.Session, input io.Reader, output io.Writer) error {
	_, err := tea.NewProgram(newModel(ctx, session), tea.WithContext(ctx), tea.WithInput(input), tea.WithOutput(output)).Run()
	return err
}

func (m *Model) resize() {
	width := max(20, m.width-6)
	m.output.SetWidth(width)
	reserved := 20
	if m.view == viewLogs {
		reserved = 8
	}
	m.output.SetHeight(max(4, m.height-reserved))
	m.input.SetWidth(width)
	m.refreshTranscript()
}

func (m *Model) refreshTranscript() {
	events := m.session.TranscriptTimeline()
	if m.view == viewLogs {
		events = m.session.LogTimeline()
	}
	m.timelineRevision = m.session.TimelineRevision()
	lines := make([]string, 0, len(events)+1)
	if len(events) == 0 {
		empty := "No commands have been evaluated in this scope."
		if m.view == viewLogs {
			empty = "No application logs have been captured."
		}
		lines = append(lines, dimStyle.Render(empty))
	}
	for _, event := range events {
		switch event.Kind {
		case "motd":
			lines = append(lines, accentStyle.Render(event.Text))
		case "command":
			lines = append(lines, accentStyle.Render("❯ ")+event.Text)
		case "value":
			lines = append(lines, valueStyle.Render(event.Text))
		case "error", "log-error":
			lines = append(lines, errorStyle.Render(event.Text))
		case "log-warn":
			lines = append(lines, warningStyle.Render(event.Text))
		case "log-debug":
			lines = append(lines, dimStyle.Render(event.Text))
		default:
			lines = append(lines, event.Text)
		}
	}
	m.output.SetContent(strings.Join(lines, "\n"))
	m.output.GotoBottom()
}

func refreshLogs() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return refreshMsg{} })
}

func (m *Model) complete(reverse bool) {
	if len(m.candidates) == 0 {
		originalToken := completionToken(m.input.Value())
		candidates, err := m.session.Complete(m.ctx, m.input.Value())
		if err != nil {
			m.status = err.Error()
			return
		}
		m.candidates = candidates
		m.candidateAt = -1
		prefix := shell.SharedPrefix(candidates)
		if prefix != "" {
			m.replaceToken(prefix)
		}
		if len(candidates) > 1 && prefix != originalToken {
			m.status = fmt.Sprintf("%d candidates", len(candidates))
			return
		}
	}
	if len(m.candidates) == 0 {
		m.status = "no completions"
		return
	}
	if reverse {
		m.candidateAt = (m.candidateAt - 1 + len(m.candidates)) % len(m.candidates)
	} else {
		m.candidateAt = (m.candidateAt + 1) % len(m.candidates)
	}
	candidate := m.candidates[m.candidateAt]
	m.replaceToken(candidate.Value)
	m.status = fmt.Sprintf("%d/%d %s", m.candidateAt+1, len(m.candidates), candidate.Detail)
}

func (m *Model) replaceToken(value string) {
	source := m.input.Value()
	token := completionToken(source)
	m.input.SetValue(strings.TrimSuffix(source, token) + value)
}

func (m *Model) moveHistory(delta int) {
	history := m.session.History()
	if len(history) == 0 {
		return
	}
	m.history = max(0, min(len(history), m.history+delta))
	if m.history == len(history) {
		m.input.Reset()
	} else {
		m.input.SetValue(history[m.history])
	}
	m.candidates = nil
}

func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !(current == '_' || current == '.' || current >= 'a' && current <= 'z' || current >= 'A' && current <= 'Z' || current >= '0' && current <= '9')
	})
	return source[index+1:]
}
