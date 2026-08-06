// Package tui presents a shared shell session in a renderer-free terminal.
package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gravestench/dark-magic/internal/shell"
)

var (
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D69A2D")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#817A6D"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B5E"))
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3A7"))
	panelStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6E5228"))
)

type evaluationMsg shell.Entry

// Model is the Charmbracelet adapter for a renderer-independent shell Session.
type Model struct {
	ctx         context.Context
	session     *shell.Session
	input       textarea.Model
	output      viewport.Model
	width       int
	height      int
	busy        bool
	history     int
	candidates  []shell.Candidate
	candidateAt int
	status      string
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

func (m Model) Init() tea.Cmd { return textarea.Blink }

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.resize()
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "ctrl+s", "alt+enter":
			if m.busy || strings.TrimSpace(m.input.Value()) == "" {
				return m, nil
			}
			source := m.input.Value()
			m.busy, m.status = true, "evaluating"
			m.candidates = nil
			return m, func() tea.Msg { return evaluationMsg(m.session.Submit(m.ctx, source)) }
		case "tab", "shift+tab":
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
	}

	var commands []tea.Cmd
	var command tea.Cmd
	m.input, command = m.input.Update(message)
	commands = append(commands, command)
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
	status := dimStyle.Render("Ctrl-S run  Enter newline  Tab complete  Alt-↑/↓ history  Ctrl-Q quit")
	if m.status != "" {
		status += "  " + accentStyle.Render(m.status)
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		capabilities,
		panelStyle.Width(max(1, m.width-2)).Render(m.output.View()),
		panelStyle.Width(max(1, m.width-2)).Render(m.input.View()),
		status,
	)
	return tea.NewView(content)
}

// Run owns the terminal lifecycle but not the supplied shell session.
func Run(ctx context.Context, session *shell.Session, input io.Reader, output io.Writer) error {
	_, err := tea.NewProgram(newModel(ctx, session), tea.WithContext(ctx), tea.WithInput(input), tea.WithOutput(output)).Run()
	return err
}

func (m *Model) resize() {
	width := max(20, m.width-6)
	m.output.SetWidth(width)
	m.output.SetHeight(max(4, m.height-13))
	m.input.SetWidth(width)
	m.refreshTranscript()
}

func (m *Model) refreshTranscript() {
	entries := m.session.Transcript()
	lines := make([]string, 0, len(entries)*2+1)
	if len(entries) == 0 {
		lines = append(lines, dimStyle.Render("No commands have been evaluated in this scope."))
	}
	for _, entry := range entries {
		lines = append(lines, accentStyle.Render("❯ ")+entry.Source)
		if entry.Error != "" {
			lines = append(lines, errorStyle.Render(entry.Error))
		} else if entry.Result.Text != "" {
			lines = append(lines, valueStyle.Render(entry.Result.Text))
		}
	}
	m.output.SetContent(strings.Join(lines, "\n"))
	m.output.GotoBottom()
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
