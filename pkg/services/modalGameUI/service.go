package modalGameUI

import (
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	common.Service

	assets   assetLoader.Dependency
	renderer raylibRenderer.Dependency
	input    input.Dependency

	currentMode string
	modals      []ModalGameUI
	rootNode    raylibRenderer.Renderable
	cursorNode  raylibRenderer.Renderable
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	time.Sleep(time.Second * 5)
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
	dc6Cursor, err := s.assets.LoadDc6(paths.CursorDefault)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 100)
		dc6Cursor, err = s.assets.LoadDc6(paths.CursorDefault)
	}

	act1, err := s.assets.ExtractPaletteFromPl2(paths.PaletteTransformAct1)
	for err != nil {
		// TODO :: fix a race condition with the mpq loader
		time.Sleep(time.Millisecond * 100)
		act1, err = s.assets.ExtractPaletteFromPl2(paths.PaletteTransformAct1)
	}

	dc6Cursor.SetPalette(act1)

	d1 := dc6Cursor.Directions[0]
	frames := d1.Frames
	frameIdx := 0
	forward := true
	frame := frames[frameIdx]

	s.cursorNode = s.renderer.NewRenderable()
	s.cursorNode.SetZIndex(99999)
	s.cursorNode.SetOrigin(0, 0)
	s.cursorNode.SetImage(frame.ToImageRGBA())

	t := time.Now()

	s.cursorNode.OnUpdate(func() {
		if time.Since(t) < time.Second/24 {
			return
		}

		x, y := s.input.MouseCursorState()
		s.cursorNode.SetPosition(float32(x), float32(y))

		if frameIdx == len(frames)-1 {
			forward = false
		} else if frameIdx == 0 {
			forward = true
		}

		if forward {
			frameIdx++
		} else {
			frameIdx--
		}

		s.cursorNode.SetImage(frames[frameIdx%len(frames)].ToImageRGBA())

		t = time.Now()
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
	if s.assets == nil {
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
