package configFile

import (
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (s *Service) ModalTui() (name string, model tea.Model) {
	return s.Name(), &tui{Service: s}
}

type tui struct {
	*Service
}

func (m *tui) Init() tea.Cmd {
	return nil
}

func (m *tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *tui) View() string {
	list := make(map[string]string)

	longestServiceName := 0
	longestConfigName := 0

	for _, service := range m.rt.Services() {
		if candidate, ok := service.(HasConfig); ok {
			name := service.Name()
			path := filepath.Join(m.ConfigDirectory(), candidate.ConfigFileName())

			if len(name) > longestServiceName {
				longestServiceName = len(name)
			}

			if len(path) > longestConfigName {
				longestConfigName = len(path)
			}

			list[name] = path
		}
	}

	styleHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef7aef"))

	rAlign := lipgloss.NewStyle().
		Width(longestServiceName+2).
		Padding(0, 1).
		Align(lipgloss.Right)

	lAlign := lipgloss.NewStyle().
		Width(longestConfigName+2).
		Padding(0, 1).
		Align(lipgloss.Left)

	output := rAlign.Render("Service Name") + lAlign.Render("Filepath")
	output = styleHeader.Render(output) + "\r\n"

	names := make([]string, 0)
	for name := range list {
		names = append(names, name)
	}

	sort.Strings(names)

	for idx, name := range names {
		output += rAlign.Render(name) + lAlign.Render(list[name])

		if idx != len(names)-1 {
			output += "\r\n"
		}
	}

	return output
}
