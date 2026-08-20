package raylibshell

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// Handle consumes one stable input frame and reports whether scene input must
// remain suppressed, including while a closed overlay animates off-screen.
func (o *Overlay) Handle(ctx context.Context, frame inputstate.Frame) bool {
	o.pollEvaluation()

	if frame.Actions["shell_toggle"].Pressed {
		o.setOpen(!o.open, time.Now())
		o.resetCompletion()

		return true
	}

	if !o.open {
		return o.progress > 0
	}

	if o.handleModalAction(frame) {
		return true
	}

	if o.view == viewLogs {
		o.navigateLogs(frame)

		return true
	}

	return o.handleLuaInput(ctx, frame)
}

// pollEvaluation clears the busy flag only from the renderer/input goroutine,
// avoiding direct Overlay mutation from the evaluation goroutine.
func (o *Overlay) pollEvaluation() {
	select {
	case <-o.finished:
		o.busy = false
	default:
	}
}

// handleModalAction applies close and tab-switch actions before view-specific
// input, preserving their priority when multiple actions arrive in one frame.
func (o *Overlay) handleModalAction(frame inputstate.Frame) bool {
	if frame.Actions["cancel"].Pressed {
		o.setOpen(false, time.Now())

		return true
	}

	if frame.Actions["shell_lua"].Pressed {
		o.view = viewLua
		o.resetCompletion()

		return true
	}

	if frame.Actions["shell_logs"].Pressed {
		o.view = viewLogs
		o.logOffset = 0
		o.resetCompletion()

		return true
	}

	return false
}

// navigateLogs updates only the log scrollback; Lua editor actions are ignored
// while the log tab owns the modal input frame.
func (o *Overlay) navigateLogs(frame inputstate.Frame) {
	switch {
	case frame.Actions["page_up"].Pressed:
		o.logOffset += 10
	case frame.Actions["page_down"].Pressed:
		o.logOffset = max(0, o.logOffset-10)
	case frame.Actions["up"].Pressed:
		o.logOffset++
	case frame.Actions["down"].Pressed:
		o.logOffset = max(0, o.logOffset-1)
	}
}

// handleLuaInput applies page scrolling, editing, completion/history, text, and
// submission in the established order so combined frame actions remain compatible.
func (o *Overlay) handleLuaInput(ctx context.Context, frame inputstate.Frame) bool {
	if frame.Actions["page_up"].Pressed {
		o.luaOffset += 10

		return true
	}

	if frame.Actions["page_down"].Pressed {
		o.luaOffset = max(0, o.luaOffset-10)

		return true
	}

	o.applyEditorActions(frame)
	o.applyCompletionOrHistory(ctx, frame)

	text := strings.ReplaceAll(strings.ReplaceAll(frame.Text, "`", ""), "~", "")
	if text != "" {
		o.insert(text)
		o.resetCompletion()
	}

	if frame.Actions["confirm"].Pressed {
		if frame.Actions["shift"].Down || frame.Actions["shift"].Pressed {
			o.insert("\n")
		} else {
			o.submit(ctx)
		}
	}

	return true
}

// applyEditorActions allows independent edit/navigation actions from the same
// frame and resets stale completion state after every input mutation.
func (o *Overlay) applyEditorActions(frame inputstate.Frame) {
	if frame.Actions["backspace"].Pressed {
		o.backspace()
		o.resetCompletion()
	}

	if frame.Actions["delete"].Pressed {
		o.delete()
		o.resetCompletion()
	}

	if frame.Actions["left"].Pressed {
		o.cursor = max(0, o.cursor-1)
		o.resetCompletion()
	}

	if frame.Actions["right"].Pressed {
		o.cursor = min(utf8.RuneCountInString(o.input), o.cursor+1)
		o.resetCompletion()
	}

	if frame.Actions["home"].Pressed {
		o.cursor = 0
		o.resetCompletion()
	}

	if frame.Actions["end"].Pressed {
		o.cursor = utf8.RuneCountInString(o.input)
		o.resetCompletion()
	}
}

// applyCompletionOrHistory gives Tab priority over Up/Down, matching the prior
// mutually exclusive input chain when a frame contains several actions.
func (o *Overlay) applyCompletionOrHistory(ctx context.Context, frame inputstate.Frame) {
	if frame.Actions["tab"].Pressed {
		if !o.busy {
			reverse := frame.Actions["shift"].Down || frame.Actions["shift"].Pressed
			o.complete(ctx, reverse)
		}

		return
	}

	if frame.Actions["up"].Pressed {
		o.moveHistory(-1)
	} else if frame.Actions["down"].Pressed {
		o.moveHistory(1)
	}
}
