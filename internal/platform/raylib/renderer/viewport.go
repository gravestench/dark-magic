package raylibRenderer

import rl "github.com/gen2brain/raylib-go/raylib"

// SubscribeViewport reports initial and changed render dimensions from the
// renderer owner thread. It never calls user code after cancellation.
func (s *Service) SubscribeViewport(callback func(width, height int)) func() {
	if callback == nil {
		return func() {}
	}
	lastWidth, lastHeight := -1, -1
	return s.SubscribeFrame(func() {
		width, height := int(rl.GetRenderWidth()), int(rl.GetRenderHeight())
		if width <= 0 || height <= 0 || width == lastWidth && height == lastHeight {
			return
		}
		lastWidth, lastHeight = width, height
		callback(width, height)
	})
}
