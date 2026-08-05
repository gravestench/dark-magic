package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) update() {
	s.frameMux.RLock()
	callbacks := append([]func(){}, s.frameCallbacks...)
	s.frameMux.RUnlock()
	for _, callback := range callbacks {
		callback()
	}
	s.rootNode.update()
	s.rootNode.UpdateWorldMatrix(rl.MatrixIdentity())
}
