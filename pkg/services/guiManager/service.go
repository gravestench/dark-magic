package guiManager

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
)

type Service struct {
	logger *slog.Logger

	dc6      dc6Loader.Dependency
	mpq      mpqLoader.Dependency
	sprite   spriteManager.Dependency
	renderer raylibRenderer.Dependency
	input    input.Dependency

	inputState struct {
		keys []int
	}
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "GUI"
}

func (s *Service) Ready() bool {
	if s.mpq == nil {
		return false
	}

	if s.dc6 == nil {
		return false
	}

	if s.sprite == nil {
		return false
	}

	if s.renderer == nil {
		return false
	}

	if s.input == nil {
		return false
	}

	return true
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
