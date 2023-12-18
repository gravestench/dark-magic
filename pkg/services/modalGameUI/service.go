package modalGameUI

import (
	"log/slog"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
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

	currentMode string
	modals      []ModalGameUI
	rootNode    raylibRenderer.Renderable
	cursorNode  raylibRenderer.Renderable
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.modals = make([]ModalGameUI, 0)
	s.rootNode = s.renderer.NewRenderable()
	s.rootNode.Disable()

	s.initCursor()

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
	s.cursorNode = s.renderer.NewRenderable()

	time.Sleep(time.Millisecond * 100)

	dc6Cursor, err := s.dc6.Load(paths.CursorDefault)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 100)
		dc6Cursor, err = s.dc6.Load(paths.CursorDefault)
	}

	act1, err := s.pl2.ExtractPaletteFromPl2(paths.PaletteTransformAct1)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 100)
		act1, err = s.pl2.ExtractPaletteFromPl2(paths.PaletteTransformAct1)
	}

	dc6Cursor.SetPalette(act1)

	d1 := dc6Cursor.Directions[0]
	frames := d1.Frames
	frame := 0
	forward := true
	s.cursorNode.SetImage(frames[frame].ToImageRGBA())

	t := time.Now()

	s.cursorNode.OnUpdate(func() {
		if time.Since(t) < time.Second/24 {
			return
		}

		t = time.Now()

		if frame == len(frames)-1 {
			forward = false
		} else if frame == 0 {
			forward = true
		}

		if forward {
			frame++
		} else {
			frame--
		}

		x, y := s.input.MouseCursorState()
		s.cursorNode.SetPosition(float32(x), float32(y))
		s.cursorNode.SetImage(frames[frame%len(frames)].ToImageRGBA())
	})
}

func (s *Service) initUpdateLoop() {
	ticker := time.NewTicker(time.Second / 24)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, modal := range s.modals {
					go modal.Update()
				}
			}
		}
	}()
}

func (s *Service) Name() string {
	return "Modal Game UI"
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
