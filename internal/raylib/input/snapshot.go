package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/inputcore"
)

// Snapshot translates the legacy backend state into stable logical actions.
// This adapter disappears when input is constructed directly by the new host.
func (s *Service) Snapshot() inputcore.Frame {
	x, y := s.MouseCursorState()
	var text []rune
	for character := rl.GetCharPressed(); character > 0; character = rl.GetCharPressed() {
		text = append(text, rune(character))
	}
	return inputcore.Frame{
		CursorX: float64(x),
		CursorY: float64(y),
		Text:    string(text),
		Actions: map[string]inputcore.ActionState{
			"pointer_primary": actionState(s.MouseButtonState()[rl.MouseLeftButton]),
			"confirm":         actionState(s.KeyState(rl.KeyEnter)),
			"cancel":          actionState(s.KeyState(rl.KeyEscape)),
			"inventory":       actionState(s.KeyState(rl.KeyI)),
			"character":       actionState(s.KeyState(rl.KeyC)),
			"skills":          actionState(s.KeyState(rl.KeyT)),
			"automap":         actionState(s.KeyState(rl.KeyTab)),
			"options":         actionState(s.KeyState(rl.KeyO)),
			"pause":           actionState(s.KeyState(rl.KeyP)),
			"up":              actionState(s.KeyState(rl.KeyUp)),
			"down":            actionState(s.KeyState(rl.KeyDown)),
			"left":            actionState(s.KeyState(rl.KeyLeft)),
			"right":           actionState(s.KeyState(rl.KeyRight)),
			"backspace":       actionState(s.KeyState(rl.KeyBackspace)),
		},
	}
}

func actionState(state InputState) inputcore.ActionState {
	return inputcore.ActionState{
		Down:     state == StateDown,
		Pressed:  state == StatePressed,
		Released: state == StateReleased,
	}
}
