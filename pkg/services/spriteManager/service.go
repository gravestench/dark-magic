package spriteManager

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
)

type Service struct {
	logger *slog.Logger
	config *Config

	assets assetLoader.Dependency

	spriteCache *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "Sprite Manager"
}

func (s *Service) Ready() bool {
	if s.assets == nil {
		return false
	}

	if s.config == nil {
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
