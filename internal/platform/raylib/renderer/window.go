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

// Resolution returns the fixed logical game surface unless native rendering was selected. Native mode follows the
// drawable window exactly, so the renderer has no final scaling step.
func (s *Service) Resolution() (width, height int) {
	if s.config != nil && s.config.Resolution.Native {
		if s.isInit.Load() {
			width, height = s.WindowSize()
			if width > 0 && height > 0 {
				return width, height
			}
		}
		return s.config.Window.Width, s.config.Window.Height
	}
	return s.config.Resolution.Width, s.config.Resolution.Height
}
