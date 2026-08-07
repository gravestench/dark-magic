package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// Snapshot translates the legacy backend state into stable logical actions.
// This adapter disappears when input is constructed directly by the new host.
func (s *Service) Snapshot() inputstate.Frame {
	x, y := s.MouseCursorState()
	pointer := actionState(s.MouseButtonState()[int32(rl.MouseButtonLeft)])
	confirm := actionState(s.KeyState(rl.KeyEnter))
	cancel := actionState(s.KeyState(rl.KeyEscape))
	space := actionState(s.KeyState(rl.KeySpace))
	var text []rune
	for character := rl.GetCharPressed(); character > 0; character = rl.GetCharPressed() {
		text = append(text, rune(character))
	}
	return inputstate.Frame{
		CursorX: float64(x),
		CursorY: float64(y),
		Text:    string(text),
		Actions: map[string]inputstate.ActionState{
			"shell_toggle":    actionState(s.KeyState(rl.KeyGrave)),
			"shell_lua":       actionState(s.KeyState(rl.KeyF1)),
			"shell_logs":      actionState(s.KeyState(rl.KeyF2)),
			"pointer_primary": pointer,
			"confirm":         confirm,
			"cancel":          cancel,
			"skip":            mergeActionStates(pointer, confirm, cancel, space),
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
			"delete":          actionState(s.KeyState(rl.KeyDelete)),
			"tab":             actionState(s.KeyState(rl.KeyTab)),
			"page_up":         actionState(s.KeyState(rl.KeyPageUp)),
			"page_down":       actionState(s.KeyState(rl.KeyPageDown)),
			"home":            actionState(s.KeyState(rl.KeyHome)),
			"end":             actionState(s.KeyState(rl.KeyEnd)),
			"shift": mergeActionStates(
				actionState(s.ModifierKeyState(rl.KeyLeftShift)),
				actionState(s.ModifierKeyState(rl.KeyRightShift)),
			),
		},
	}
}

func mergeActionStates(states ...inputstate.ActionState) inputstate.ActionState {
	var result inputstate.ActionState
	for _, state := range states {
		result.Down = result.Down || state.Down
		result.Pressed = result.Pressed || state.Pressed
		result.Released = result.Released || state.Released
	}
	return result
}

func actionState(state InputState) inputstate.ActionState {
	return inputstate.ActionState{
		Down:     state == StateDown,
		Pressed:  state == StatePressed,
		Released: state == StateReleased,
	}
}
