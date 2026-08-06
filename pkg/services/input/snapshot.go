package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/inputcore"
)

// Snapshot translates the legacy backend state into stable logical actions.
// This adapter disappears when input is constructed directly by the new host.
func (s *Service) Snapshot() inputcore.Frame {
	x, y := s.MouseCursorState()
	return inputcore.Frame{
		CursorX: float64(x),
		CursorY: float64(y),
		Actions: map[string]inputcore.ActionState{
			"confirm":   actionState(s.KeyState(rl.KeyEnter)),
			"cancel":    actionState(s.KeyState(rl.KeyEscape)),
			"inventory": actionState(s.KeyState(rl.KeyI)),
			"up":        actionState(s.KeyState(rl.KeyUp)),
			"down":      actionState(s.KeyState(rl.KeyDown)),
			"left":      actionState(s.KeyState(rl.KeyLeft)),
			"right":     actionState(s.KeyState(rl.KeyRight)),
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
