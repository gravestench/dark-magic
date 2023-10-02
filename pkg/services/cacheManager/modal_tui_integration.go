package cacheManager

import (
	"fmt"
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
	type CacheStats struct {
		ServiceName string
		Size        string
		Budget      string
		Saturation  float64
	}

	cacheStats := make([]CacheStats, 0)

	longestName := 0

	var keys []string
	for name := range m.caches {
		if len(name) > longestName {
			longestName = len(name)
		}

		keys = append(keys, name)
	}

	sort.Strings(keys)

	if len(m.caches) < 1 {
		return "no caches present"
	}

	for _, name := range keys {
		cache := m.caches[name]
		w := cache.GetWeight()
		b := cache.GetBudget()

		stats := CacheStats{
			ServiceName: name,
			Size:        formatBytes(w),
			Budget:      formatBytes(b),
			Saturation:  float64(w) / float64(b),
		}

		cacheStats = append(cacheStats, stats)
	}

	var output string

	styleHeader := lipgloss.NewStyle().Foreground(lipgloss.Color("#ef7aef"))

	rAlign := lipgloss.NewStyle().
		Width(longestName+2).
		Padding(0, 1).
		Align(lipgloss.Right)

	lAlign := lipgloss.NewStyle().
		Width(12).
		Padding(0, 1).
		Align(lipgloss.Left)

	header := rAlign.Render("Service Name") + lAlign.Render("Size") + lAlign.Render("Budget") + lAlign.Render("Saturation")
	output += styleHeader.Render(header)
	for _, stats := range cacheStats {
		red := uint8(stats.Saturation * 255)
		green := 255 - uint8(stats.Saturation*196)
		fgSaturation := fmt.Sprintf("#%X%X22", red, green)
		styleSaturation := lipgloss.NewStyle().Foreground(lipgloss.Color(fgSaturation))
		percentSaturation := fmt.Sprintf("%0.2f%%", stats.Saturation*100)

		output += "\r\n"
		output += rAlign.Render(stats.ServiceName) +
			lAlign.Render(stats.Size) +
			lAlign.Render(stats.Budget) +
			lAlign.Render(fmt.Sprintf("%7s", styleSaturation.Render(percentSaturation)))
	}

	return output
}
