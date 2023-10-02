package modalTui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	rt     runtime.Runtime
	logger *zerolog.Logger
	cfg    configFile.Dependency
	isInit bool
	mux    sync.Mutex
	modalUiModel
}

func (s *Service) Init(rt runtime.Runtime) {
	s.rt = rt
	s.modalUiModel.modals = make(map[string]tea.Model)

	dir := s.cfg.ConfigDirectory()
	logPath := expandHomeDirectory(filepath.Join(dir, "output.log"))

	redirect, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Fatal().Msgf("Error opening or creating the file: %v", err)
	}

	rt.SetLogDestination(redirect)

	clearScreen()

	// bubbletea kicked off using runtime event handler

	for _, service := range rt.Services() {
		go s.attemptBindService(service)
	}

	s.isInit = true
}

func (s *Service) OnShutdown() {
	s.rt.SetLogDestination(os.Stdout)

	dir := s.cfg.ConfigDirectory()
	logPath := filepath.Join(dir, "output.log")

	s.logger.Info().Msgf("see %q for log output", logPath)
}

func (s *Service) Name() string {
	return "Modal TUI"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func clearScreen() {
	// ANSI escape code to clear the screen
	fmt.Print("\033[H\033[2J")
}

func (s *Service) attemptBindService(service runtime.Service) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for !s.isInit {
		time.Sleep(time.Second)
	}

	tui, ok := service.(HasModalTextUserInterface)
	if !ok {
		return
	}

	name, modal := tui.ModalTui()

	s.modals[name] = modal

	if s.selectedModal == "" {
		s.selectedModal = name
	}

	if current, exists := s.modals[s.selectedModal]; exists {
		current.Update(tea.Msg(1))
	}
}

func expandHomeDirectory(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = strings.Replace(path, "~", homeDir, 1)
		}
	}
	return path
}
