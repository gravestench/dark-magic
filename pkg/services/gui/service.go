package gui

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
)

type Service struct {
	logger *slog.Logger

	dc6      dc6Loader.Dependency
	pl2      pl2Loader.Dependency
	mpq      mpqLoader.Dependency
	sprite   spriteManager.Dependency
	renderer raylibRenderer.Dependency
	input    input.Dependency

	config *configFile.Config

	inputState struct {
		keys []int
	}

	animations map[uuid.UUID]animation
	lastUpdate time.Time
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.animations = make(map[uuid.UUID]animation)
	go s.animationUpdateLoop()
}

func (s *Service) Name() string {
	return "GUI"
}

func (s *Service) Ready() bool {
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

	if s.pl2 == nil {
		return false
	}

	if s.config == nil {
		return false
	}

	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
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

func (s *Service) animationUpdateLoop() {
	fps := s.config.Group(groupGUI).GetInt(keyFPS)
	frameDuration := time.Second / time.Duration(fps)
	ticker := time.NewTicker(frameDuration)

	for {
		<-ticker.C
		s.updateAnimations()
		s.lastUpdate = time.Now()
	}
}

func (s *Service) updateAnimations() {
	for _, anim := range s.animations {
		anim.Advance(time.Since(s.lastUpdate))
	}
}
