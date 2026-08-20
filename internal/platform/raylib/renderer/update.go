package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// update runs owner-thread callbacks before propagating scene transforms. Input and window updates therefore affect the
// same frame whose world matrices are calculated immediately afterward.
func (s *Service) update() {
	runCallbackSnapshot(s.frameSnapshot.Load())

	s.rootNode.UpdateWorldMatrix(rl.MatrixIdentity(), false)
}
