package modalTui

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	modals        map[string]tea.Model
	selectedModal string
}

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// switching the current modal
		case tea.KeyCtrlOpenBracket.String():
			m.switchPreviousModal()

		case tea.KeyCtrlCloseBracket.String():
			m.switchNextModal()
		}
	}

	m.CurrentModal().Update(msg)

	return m, doTick()
}

func (m *model) View() string {
	current, exists := m.modals[m.selectedModal]
	if !exists {
		return (&defaultModel{}).View()
	}

	styleModal := lipgloss.NewStyle().
		BorderBottom(true).
		BorderTop(true).
		BorderStyle(lipgloss.RoundedBorder())

	view := m.headerLine() + "\r\n"
	view += styleModal.Render(current.View()) + "\r\n"
	view += m.footerLine()

	return view
}

func (m *model) headerLine() string {
	selectionLine := ""

	for idx, name := range m.GetModalNames() {
		if name == m.selectedModal {
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4"))

			name = style.Render(fmt.Sprintf("%s", name))
		}

		if idx == 0 {
			selectionLine = name
			continue
		}

		selectionLine = fmt.Sprintf("%s %s", selectionLine, name)
	}

	return selectionLine
}

func (m *model) footerLine() string {
	const (
		helpString = "( Previous ctrl+[ ) ( Next ctrl+] )"
	)

	return fmt.Sprintf("\r\n%s", helpString)
}

func (m *model) CurrentModal() tea.Model {
	if current, exists := m.modals[m.selectedModal]; exists {
		return current
	}

	return &defaultModel{}
}

func (m *model) CurrentModalName() string {
	return m.selectedModal
}

func (m *model) GetModalNames() (keys []string) {
	return m.getSortedModalNameList()
}

func (m *model) SwitchModal(name string) {
	if _, exists := m.modals[name]; exists {
		m.selectedModal = name
	}
}

func (m *model) getSortedModalNameList() (keys []string) {
	for key := range m.modals {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return
}

func (m *model) switchPreviousModal() {
	m.SwitchModal(m.previous(m.CurrentModalName()))
}

func (m *model) switchNextModal() {
	m.SwitchModal(m.next(m.CurrentModalName()))
}

func (m *model) previous(current string) string {
	keys := m.GetModalNames()

	for i, item := range keys {
		if item != current {
			continue
		}

		// If the current item is the first item in the list, wrap around to the last item.
		if i == 0 {
			return keys[len(keys)-1]
		}

		// Return the previous item in the list.
		return keys[i-1]
	}

	// If the current item is not found in the list, default to the first item.
	return keys[0]
}

func (m *model) next(current string) string {
	keys := m.GetModalNames()

	for i, item := range keys {
		if item == current {
			if i == len(keys)-1 {
				// If the current item is the last item in the list, wrap around to the first item.
				return keys[0]
			}
			// Return the next item in the list.
			return keys[i+1]
		}
	}
	// If the current item is not found in the list, default to the first item.
	return keys[0]
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second/4, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
