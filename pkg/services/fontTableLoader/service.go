package fontTableLoader

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *zerolog.Logger
	mpq    mpqLoader.Dependency
}

func (s *Service) DependenciesResolved() bool {
	if s.mpq == nil {
		return false
	}

	if len(s.mpq.Archives()) != 11 {
		time.Sleep(time.Second)
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}

func (s *Service) Init(rt runtime.R) {

}

func (s *Service) Name() string {
	return "GPL Loader"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}
