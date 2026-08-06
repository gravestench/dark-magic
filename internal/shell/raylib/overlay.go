// Package raylib presents a shared shell session inside the graphical client.
package raylibshell

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/shell"
	"golang.org/x/image/font/gofont/gomono"
)

const (
	fontSize   = float32(18)
	lineHeight = int32(23)
)

// Overlay adapts the shared session to Raylib input and immediate-mode debug
// drawing. Evaluation remains serialized by Session and its Lua runtime.
type Overlay struct {
	session     *shell.Session
	open        bool
	busy        bool
	input       string
	history     int
	candidates  []shell.Candidate
	candidateAt int
	finished    chan shell.Entry
	font        rl.Font
	fontLoaded  bool
}

func New(session *shell.Session) *Overlay {
	return &Overlay{session: session, history: len(session.History()), finished: make(chan shell.Entry, 1)}
}

func (o *Overlay) Open() bool { return o.open }

// LoadFont creates the GPU atlas on the renderer owner thread. Go Mono is
// embedded by the existing x/image dependency, so console typography is stable
// across platforms and does not depend on user-installed fonts.
func (o *Overlay) LoadFont() error {
	if o.fontLoaded {
		return nil
	}
	codepoints := make([]rune, 0, 224)
	for current := rune(32); current <= 255; current++ {
		codepoints = append(codepoints, current)
	}
	o.font = rl.LoadFontFromMemory(".ttf", gomono.TTF, 36, codepoints)
	if !rl.IsFontValid(o.font) {
		return fmt.Errorf("raylib shell: load embedded Go Mono font")
	}
	rl.SetTextureFilter(o.font.Texture, rl.FilterBilinear)
	o.fontLoaded = true
	return nil
}

// Close releases the font atlas before the renderer destroys its GPU context.
func (o *Overlay) Close() {
	if !o.fontLoaded {
		return
	}
	rl.UnloadFont(o.font)
	o.fontLoaded = false
}

// Handle consumes one stable input frame and reports whether scene input must
// be suppressed for this frame.
func (o *Overlay) Handle(ctx context.Context, frame inputstate.Frame) bool {
	select {
	case <-o.finished:
		o.busy = false
	default:
	}
	if frame.Actions["shell_toggle"].Pressed {
		o.open = !o.open
		o.resetCompletion()
		return true
	}
	if !o.open {
		return false
	}
	if frame.Actions["cancel"].Pressed {
		o.open = false
		return true
	}
	if frame.Actions["backspace"].Pressed {
		o.input = trimLastRune(o.input)
		o.resetCompletion()
	}
	if frame.Actions["delete"].Pressed {
		o.input = ""
		o.resetCompletion()
	}
	if frame.Actions["tab"].Pressed {
		if !o.busy {
			o.complete(ctx, frame.Actions["shift"].Down || frame.Actions["shift"].Pressed)
		}
	} else if frame.Actions["up"].Pressed {
		o.moveHistory(-1)
	} else if frame.Actions["down"].Pressed {
		o.moveHistory(1)
	}
	if text := strings.ReplaceAll(strings.ReplaceAll(frame.Text, "`", ""), "~", ""); text != "" {
		o.input += text
		o.resetCompletion()
	}
	if frame.Actions["confirm"].Pressed {
		if frame.Actions["shift"].Down || frame.Actions["shift"].Pressed {
			o.input += "\n"
		} else {
			o.submit(ctx)
		}
	}
	return true
}

func (o *Overlay) Draw(width, height int) {
	if !o.open {
		return
	}
	panelHeight := int32(max(220, height*3/5))
	rl.DrawRectangle(0, 0, int32(width), panelHeight, rl.NewColor(8, 7, 6, 238))
	rl.DrawRectangleLinesEx(rl.NewRectangle(0, 0, float32(width), float32(panelHeight)), 2, rl.NewColor(176, 119, 38, 255))
	policy := o.session.Policy()
	mode := "read-only"
	if policy.Mutable {
		mode = "mutable"
	}
	o.drawText("DARK MAGIC LUA SHELL", 16, 12, 22, rl.NewColor(222, 163, 58, 255))
	o.drawText(fmt.Sprintf("target %s | policy %s (%s)", o.session.Target(), policy.Name, mode), 16, 40, 16, rl.Gray)

	transcriptTop := int32(68)
	promptTop := panelHeight - 58
	available := int((promptTop-transcriptTop)/lineHeight) - 1
	lines := wrapTranscript(transcriptLines(o.session.Transcript()), max(12, (width-36)/11))
	if len(lines) > available {
		lines = lines[len(lines)-available:]
	}
	for index, line := range lines {
		color := rl.LightGray
		if line.error {
			color = rl.NewColor(255, 104, 88, 255)
		} else if line.result {
			color = rl.NewColor(125, 211, 167, 255)
		}
		o.drawText(line.text, 18, transcriptTop+int32(index)*lineHeight, fontSize, color)
	}
	status := ""
	if o.busy {
		status = " [evaluating]"
	}
	o.drawText("> "+o.input+status, 16, promptTop, fontSize, rl.RayWhite)
	o.drawText("` close  Enter run  Shift+Enter newline  Tab complete  Up/Down history  Esc close", 16, panelHeight-26, 14, rl.Gray)
}

func (o *Overlay) drawText(text string, x, y int32, size float32, tint rl.Color) {
	if o.fontLoaded {
		rl.DrawTextEx(o.font, text, rl.NewVector2(float32(x), float32(y)), size, 1, tint)
		return
	}
	rl.DrawText(text, x, y, int32(size), tint)
}

func (o *Overlay) submit(ctx context.Context) {
	source := strings.TrimSpace(o.input)
	if source == "" || o.busy {
		return
	}
	o.busy = true
	o.input = ""
	o.resetCompletion()
	go func() { o.finished <- o.session.Submit(ctx, source) }()
}

func (o *Overlay) complete(ctx context.Context, reverse bool) {
	if len(o.candidates) == 0 {
		original := completionToken(o.input)
		candidates, err := o.session.Complete(ctx, o.input)
		if err != nil || len(candidates) == 0 {
			return
		}
		o.candidates, o.candidateAt = candidates, -1
		prefix := shell.SharedPrefix(candidates)
		if prefix != "" {
			o.replaceToken(prefix)
		}
		if len(candidates) > 1 && prefix != original {
			return
		}
	}
	if reverse {
		o.candidateAt = (o.candidateAt - 1 + len(o.candidates)) % len(o.candidates)
	} else {
		o.candidateAt = (o.candidateAt + 1) % len(o.candidates)
	}
	o.replaceToken(o.candidates[o.candidateAt].Value)
}

func (o *Overlay) replaceToken(value string) {
	token := completionToken(o.input)
	o.input = strings.TrimSuffix(o.input, token) + value
}

func (o *Overlay) moveHistory(delta int) {
	history := o.session.History()
	if len(history) == 0 {
		return
	}
	o.history = max(0, min(len(history), o.history+delta))
	if o.history == len(history) {
		o.input = ""
	} else {
		o.input = history[o.history]
	}
	o.resetCompletion()
}

func (o *Overlay) resetCompletion() {
	o.candidates = nil
	o.candidateAt = -1
}

func trimLastRune(value string) string {
	_, size := utf8.DecodeLastRuneInString(value)
	if size == 0 {
		return value
	}
	return value[:len(value)-size]
}

func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !(current == '_' || current == '.' || current >= 'a' && current <= 'z' || current >= 'A' && current <= 'Z' || current >= '0' && current <= '9')
	})
	return source[index+1:]
}

type transcriptLine struct {
	text          string
	result, error bool
}

func transcriptLines(entries []shell.Entry) []transcriptLine {
	lines := make([]transcriptLine, 0, len(entries)*2)
	for _, entry := range entries {
		lines = append(lines, transcriptLine{text: "> " + strings.ReplaceAll(entry.Source, "\n", " ")})
		if entry.Error != "" {
			lines = append(lines, transcriptLine{text: entry.Error, error: true})
		} else if entry.Result.Text != "" {
			lines = append(lines, transcriptLine{text: entry.Result.Text, result: true})
		}
	}
	return lines
}

func wrapTranscript(lines []transcriptLine, columns int) []transcriptLine {
	var wrapped []transcriptLine
	for _, line := range lines {
		runes := []rune(line.text)
		for len(runes) > columns {
			wrapped = append(wrapped, transcriptLine{text: string(runes[:columns]), result: line.result, error: line.error})
			runes = runes[columns:]
		}
		wrapped = append(wrapped, transcriptLine{text: string(runes), result: line.result, error: line.error})
	}
	return wrapped
}
