package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) SetWindowTitle(title string) {
	rl.SetWindowTitle(title)
}

func (s *Service) WindowSize() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

func (s *Service) Resolution() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}
