package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) update() {
	s.rootNode.(hasMatrix).UpdateWorldMatrix(rl.MatrixIdentity())
}
