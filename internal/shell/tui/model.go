// Package tui presents a shared shell session in a renderer-free terminal.
package tui

import (
	"context"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/gravestench/dark-magic/internal/shell"
)

type evaluationMsg shell.Entry
type refreshMsg struct{}
type viewMode uint8

const (
	viewLua viewMode = iota
	viewLogs
)

// Model is the Bubble Tea adapter for a renderer-independent shell Session.
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

// NewModel prepares an interactive terminal model without starting or taking
// ownership of a terminal lifecycle.
func NewModel(session *shell.Session) Model {
	return newModel(context.Background(), session)
}

// newModel binds the program context used by asynchronous submissions and
// initializes editor and viewport state before the first Bubble Tea frame.
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

	model := Model{
		ctx:     ctx,
		session: session,
		input:   input,
		output:  output,
		history: len(session.History()),
	}
	model.refreshTranscript()

	return model
}

// Init starts textarea blinking and periodic log refresh without performing any
// terminal I/O before Bubble Tea owns the program.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, refreshLogs())
}

// Update applies terminal events in priority order: window/view lifecycle,
// submission/completion, transcript refresh, then focused component updates.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.resize()
	case tea.KeyPressMsg:
		updated, command, handled := m.handleKey(message)

		m = updated
		if handled {
			return m, command
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

	return m.updateComponents(message)
}

// handleKey processes modal commands before textarea/viewport components see
// the key, preventing shortcuts from also editing or scrolling the wrong view.
func (m Model) handleKey(message tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch message.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit, true
	case "f1":
		m.view = viewLua
		m.input.Focus()
		m.resize()

		return m, nil, true
	case "f2":
		m.view = viewLogs
		m.input.Blur()
		m.resize()

		return m, nil, true
	case "ctrl+s", "alt+enter":
		if m.view != viewLua || m.busy || strings.TrimSpace(m.input.Value()) == "" {
			return m, nil, true
		}

		source := m.input.Value()
		m.busy = true
		m.status = "evaluating"
		m.candidates = nil

		return m, func() tea.Msg {
			return evaluationMsg(m.session.Submit(m.ctx, source))
		}, true
	case "tab", "shift+tab":
		if m.view != viewLua {
			return m, nil, false
		}

		m.complete(message.String() == "shift+tab")

		return m, nil, true
	case "alt+up":
		m.moveHistory(-1)

		return m, nil, true
	case "alt+down":
		m.moveHistory(1)

		return m, nil, true
	default:
		return m, nil, false
	}
}

// updateComponents forwards ordinary messages only to the active editor and to
// the always-visible output viewport, batching their asynchronous commands.
func (m Model) updateComponents(message tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	if m.view == viewLua {
		var command tea.Cmd

		m.input, command = m.input.Update(message)
		commands = append(commands, command)
	}

	var command tea.Cmd

	m.output, command = m.output.Update(message)
	commands = append(commands, command)

	return m, tea.Batch(commands...)
}

// Run owns the Bubble Tea terminal lifecycle but not the supplied shell session,
// allowing callers to close shared evaluator resources in their own order.
func Run(ctx context.Context, session *shell.Session, input io.Reader, output io.Writer) error {
	program := tea.NewProgram(
		newModel(ctx, session),
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)

	_, err := program.Run()

	return err
}

// resize allocates terminal rows differently for Lua editing and log-only views,
// then rebuilds wrapping against the new viewport width.
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

// refreshLogs schedules periodic revision checks instead of rebuilding content
// every frame when neither transcript nor process logs changed.
func refreshLogs() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}
