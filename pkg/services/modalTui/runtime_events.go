package modalTui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/faiface/mainthread"
	"github.com/gravestench/runtime"
)

func (s *Service) OnRuntimeRunLoopInitiated(_ ...interface{}) {
	go mainthread.Call(func() {
		p := tea.NewProgram(&s.modalUiModel)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	})
}

func (s *Service) OnServiceAdded(args ...any) {
	if len(args) < 1 {
		return
	}

	service, ok := args[0].(runtime.Service)
	if !ok {
		return
	}

	s.attemptBindService(service)
}
