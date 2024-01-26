package screenManager

import (
	"log/slog"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/gui"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	logger *slog.Logger

	dc6      dc6Loader.Dependency
	pl2      pl2Loader.Dependency
	renderer raylibRenderer.Dependency
	input    input.Dependency
	gui      gui.Dependency

	currentMode string
	modals      []ModalGameUI
	rootNode    raylibRenderer.Renderable
	cursorNode  raylibRenderer.Renderable
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.modals = make([]ModalGameUI, 0)
	s.rootNode = s.renderer.NewRenderable()
	s.rootNode.Disable()

	s.initCursorAnimation()

	for _, service := range mesh.Services() {
		s.attemptBindService(service)
	}

	s.initUpdateLoop()
}

func (s *Service) attemptBindService(service servicemesh.Service) {
	candidate, ok := service.(ModalGameUI)
	if !ok {
		return
	}

	s.modals = append(s.modals, candidate)
}

func (s *Service) initCursor() {
	const (
		cursor      = paths.CursorDefault
		paletteAct1 = paths.PaletteTransformAct1
	)

	anim, err := s.gui.NewAnimationDC6(cursor, paletteAct1)
	for err != nil {
		anim, err = s.gui.NewAnimationDC6(cursor, paletteAct1)
	}

	s.cursorNode = anim.Renderable()
	s.cursorNode.SetZIndex(99999)
	s.cursorNode.SetOrigin(0, 0)

	s.cursorNode.OnUpdate(func() {
		s.cursorNode.SetPosition(func() (x, y float32) {
			xInt, yInt := s.input.MouseCursorState()
			return float32(xInt), float32(yInt)
		}())
	})
}

func (s *Service) initUpdateLoop() {
	ticker := time.NewTicker(time.Second / 24)
	go func() {
		for {
			<-ticker.C
			for _, modal := range s.modals {
				go modal.Update()
			}
		}
	}()
}

func (s *Service) Name() string {
	return "Modal Game UI"
}

func (s *Service) Ready() bool {
	if s.dc6 == nil {
		return false
	}

	if s.pl2 == nil {
		return false
	}

	if s.renderer == nil {
		return false
	}

	if s.input == nil {
		return false
	}

	if s.gui == nil {
		return false
	}

	if !s.gui.Ready() {
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
