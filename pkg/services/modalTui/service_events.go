package modalTui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gravestench/servicemesh"
)

func (s *Service) OnServiceMeshRunLoopInitiated() {
	p := tea.NewProgram(&s.modalUiModel)
	if _, err := p.Run(); err != nil {
		s.logger.Error("running bubbletea TUI", "error", err)
		panic(err)
	}
}

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	s.attemptBindService(service)
}
