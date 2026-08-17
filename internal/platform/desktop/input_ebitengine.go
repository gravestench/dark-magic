//go:build ebitengine

package desktop

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

type ebitInput struct {
	renderer *ebitRenderer
	frame    inputstate.Frame
	stop     func()
}

func newEbitInput(renderer *ebitRenderer) *ebitInput { return &ebitInput{renderer: renderer} }

func (i *ebitInput) Start(context.Context) error {
	i.stop = i.renderer.SubscribeFrame(i.sample)
	return nil
}

func (i *ebitInput) Stop(context.Context) error {
	if i.stop != nil {
		i.stop()
		i.stop = nil
	}
	return nil
}

func (i *ebitInput) Snapshot() inputstate.Frame { return i.frame }

func (i *ebitInput) sample() {
	x, y := ebiten.CursorPosition()
	scrollX, scrollY := ebiten.Wheel()
	pointer := ebitMouseAction(ebiten.MouseButtonLeft)
	pointerSecondary := ebitMouseAction(ebiten.MouseButtonRight)
	confirm, cancel, space := ebitKeyAction(ebiten.KeyEnter), ebitKeyAction(ebiten.KeyEscape), ebitKeyAction(ebiten.KeySpace)
	i.frame = inputstate.Frame{
		CursorX: float64(x), CursorY: float64(y), ScrollX: scrollX, ScrollY: scrollY, Text: string(ebiten.AppendInputChars(nil)),
		Actions: map[string]inputstate.ActionState{
			"shell_toggle":      ebitKeyAction(ebiten.KeyGraveAccent),
			"shell_lua":         ebitKeyAction(ebiten.KeyF1),
			"shell_logs":        ebitKeyAction(ebiten.KeyF2),
			"debug_collision":   ebitKeyAction(ebiten.KeyF3),
			"debug_map_tiles":   ebitKeyAction(ebiten.KeyF4),
			"debug_origins":     ebitKeyAction(ebiten.KeyF5),
			"debug_combat":      ebitKeyAction(ebiten.KeyF6),
			"pointer_primary":   pointer,
			"pointer_secondary": pointerSecondary,
			"confirm":           confirm,
			"cancel":            cancel,
			"skip":              mergeEbitActions(pointer, confirm, cancel, space),
			"space":             space,
			"inventory":         ebitKeyAction(ebiten.KeyI),
			"character":         ebitKeyAction(ebiten.KeyC),
			"skills":            ebitKeyAction(ebiten.KeyT),
			"automap":           ebitKeyAction(ebiten.KeyTab),
			"help":              ebitKeyAction(ebiten.KeyH),
			"search":            ebitKeyAction(ebiten.KeyF),
			"quests":            ebitKeyAction(ebiten.KeyQ),
			"party":             ebitKeyAction(ebiten.KeyP),
			"options":           ebitKeyAction(ebiten.KeyO),
			"pause":             ebitKeyAction(ebiten.KeyPause),
			"up":                ebitRepeatKeyAction(ebiten.KeyArrowUp),
			"down":              ebitRepeatKeyAction(ebiten.KeyArrowDown),
			"left":              ebitRepeatKeyAction(ebiten.KeyArrowLeft),
			"right":             ebitRepeatKeyAction(ebiten.KeyArrowRight),
			"toggle_run":        ebitKeyAction(ebiten.KeyR),
			"swap_weapons":      ebitKeyAction(ebiten.KeyW),
			"backspace":         ebitRepeatKeyAction(ebiten.KeyBackspace),
			"delete":            ebitRepeatKeyAction(ebiten.KeyDelete),
			"tab":               ebitKeyAction(ebiten.KeyTab),
			"page_up":           ebitRepeatKeyAction(ebiten.KeyPageUp),
			"page_down":         ebitRepeatKeyAction(ebiten.KeyPageDown),
			"home":              ebitRepeatKeyAction(ebiten.KeyHome),
			"end":               ebitRepeatKeyAction(ebiten.KeyEnd),
			"shift":             mergeEbitActions(ebitKeyAction(ebiten.KeyShiftLeft), ebitKeyAction(ebiten.KeyShiftRight)),
		},
	}
}

func ebitKeyAction(key ebiten.Key) inputstate.ActionState {
	pressed := inpututil.IsKeyJustPressed(key)
	return inputstate.ActionState{Down: ebiten.IsKeyPressed(key) && !pressed, Pressed: pressed, Released: inpututil.IsKeyJustReleased(key)}
}

func ebitRepeatKeyAction(key ebiten.Key) inputstate.ActionState {
	state := ebitKeyAction(key)
	duration := inpututil.KeyPressDuration(key)
	state.Pressed = state.Pressed || duration >= 30 && (duration-30)%6 == 0
	if state.Pressed {
		state.Down = false
	}
	return state
}

func ebitMouseAction(button ebiten.MouseButton) inputstate.ActionState {
	pressed := inpututil.IsMouseButtonJustPressed(button)
	return inputstate.ActionState{Down: ebiten.IsMouseButtonPressed(button) && !pressed, Pressed: pressed, Released: inpututil.IsMouseButtonJustReleased(button)}
}

func mergeEbitActions(actions ...inputstate.ActionState) inputstate.ActionState {
	var result inputstate.ActionState
	for _, action := range actions {
		result.Down = result.Down || action.Down
		result.Pressed = result.Pressed || action.Pressed
		result.Released = result.Released || action.Released
	}
	return result
}

var _ Input = (*ebitInput)(nil)
