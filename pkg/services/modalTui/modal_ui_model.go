package modalTui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modalUiModel struct {
	modals        map[string]tea.Model
	selectedModal string
}

func (m *modalUiModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *modalUiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *modalUiModel) View() string {
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

func (m *modalUiModel) headerLine() string {
	selectionLine := ""

	for idx, name := range m.GetModalNames() {
		if name == m.selectedModal {
			style := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
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

func (m *modalUiModel) footerLine() string {
	styleItem := lipgloss.NewStyle().Width(18).Align(lipgloss.Center)
	styleLabel := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#7D56F4")).Align(lipgloss.Right)
	styleHotkey := lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#7D56F4")).Bold(true).Align(lipgloss.Left)

	var footer string

	hotkeys := []string{
		"Previous::ctrl+[",
		"Next::ctrl+]",
	}

	for _, entry := range hotkeys {
		label := strings.Split(entry, "::")[0]
		hotkey := strings.Split(entry, "::")[1]
		label = styleLabel.Render(label)
		hotkey = styleHotkey.Render(hotkey)
		footer += styleItem.Render(label + hotkey)
	}

	return footer
}

func (m *modalUiModel) CurrentModal() tea.Model {
	if current, exists := m.modals[m.selectedModal]; exists {
		return current
	}

	return &defaultModel{}
}

func (m *modalUiModel) CurrentModalName() string {
	return m.selectedModal
}

func (m *modalUiModel) GetModalNames() (keys []string) {
	return m.getSortedModalNameList()
}

func (m *modalUiModel) SwitchModal(name string) {
	if _, exists := m.modals[name]; exists {
		m.selectedModal = name
	}
}

func (m *modalUiModel) getSortedModalNameList() (keys []string) {
	for key := range m.modals {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return
}

func (m *modalUiModel) switchPreviousModal() {
	m.SwitchModal(m.previous(m.CurrentModalName()))
}

func (m *modalUiModel) switchNextModal() {
	m.SwitchModal(m.next(m.CurrentModalName()))
}

func (m *modalUiModel) previous(current string) string {
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

func (m *modalUiModel) next(current string) string {
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
