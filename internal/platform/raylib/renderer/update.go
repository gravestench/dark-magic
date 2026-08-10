package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) update() {
	if snapshot := s.frameSnapshot.Load(); snapshot != nil {
		for _, callback := range snapshot.([]func()) {
			callback()
		}
	}
	s.rootNode.UpdateWorldMatrix(rl.MatrixIdentity(), false)
}
