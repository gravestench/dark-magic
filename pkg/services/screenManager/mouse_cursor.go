package screenManager

import (
	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/gui"
)

func (s *Service) initCursorAnimation() {
	cursor, err := s.gui.NewAnimationDC6(paths.CursorDefault, paths.PaletteTransformAct1)
	for err != nil {
		cursor, err = s.gui.NewAnimationDC6(paths.CursorDefault, paths.PaletteTransformAct1)
	}

	cursor.SetPlayMode(gui.PlayYoYo)
	cursor.SetLoopingMode(gui.Looping)
	cursor.SetLoopCount(gui.LoopForever)
	cursor.Play()
	cursor.Renderable().SetZIndex(999)

	cursor.Renderable().OnUpdate(func() {
		x, y := s.input.MouseCursorState()
		cursor.Renderable().SetPosition(float32(x), float32(y))
	})
}
