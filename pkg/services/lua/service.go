package lua

import (
	"log/slog"
	"sync"
	"time"

	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	config        *configFile.Config
	logger        *slog.Logger
	state         *lua.LState
	events        *ee.EventEmitter
	mux           sync.Mutex
	boundServices map[string]any
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	if s.boundServices == nil {
		s.boundServices = make(map[string]any)
	}

	s.state = lua.NewState()
	s.bindLoggerToLuaEnvironment()

	for _, service := range mesh.Services() {
		go s.tryToExportToLuaEnvironment(service)
	}

	// wait for all siblings to be ready before we launch scripts
waitForReady:
	for {
		for _, service := range mesh.Services() {
			if !service.Ready() {
				continue waitForReady
			}

			if candidate, ok := service.(servicemesh.HasDependencies); ok {
				if !candidate.DependenciesResolved() {
					continue waitForReady
				}
			}

			break waitForReady // all deps resolved for all siblings
		}
	}

	initScriptPath := s.config.Group(s.Name()).GetString("init script")

	if err := s.runScript(initScriptPath); err != nil {
		s.logger.Error("running script", "error", err)
		time.Sleep(time.Second * 1)
	}
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
