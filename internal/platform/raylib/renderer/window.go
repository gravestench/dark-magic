package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// SetWindowTitle changes the native title on the renderer owner thread; callers should schedule it through OnFrame when
// originating from another goroutine.
func (s *Service) SetWindowTitle(title string) {
	rl.SetWindowTitle(title)
}

// WindowSize returns the current drawable size, which may differ from configuration after a resizable window changes.
func (s *Service) WindowSize() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

// Resolution returns the fixed logical game surface. Presentation scales this surface into WindowSize without changing
// simulation or input coordinates.
func (s *Service) Resolution() (width, height int) {
	return s.config.Resolution.Width, s.config.Resolution.Height
}
