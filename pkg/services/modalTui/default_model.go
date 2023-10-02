package modalTui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type defaultModel struct{}

func (d defaultModel) Init() tea.Cmd {
	return nil
}

func (d defaultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return d, nil
}

func (d defaultModel) View() string {
	return "no modals present"
}
