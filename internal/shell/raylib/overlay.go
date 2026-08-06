// Package raylib presents a shared shell session inside the graphical client.
package raylibshell

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/easing"
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
	session         *shell.Session
	open            bool
	busy            bool
	input           string
	history         int
	candidates      []shell.Candidate
	candidateAt     int
	finished        chan shell.Entry
	font            rl.Font
	fontLoaded      bool
	progress        float64
	animationAt     time.Time
	displayRevision uint64
	displayColumns  int
	displayLines    []transcriptLine
}

var (
	elasticEntrance = (&easing.ElasticOutEaseProvider{}).New([]float64{1, 0.32})
	smoothExit      = (&easing.CubicInOutEaseProvider{}).New(nil)
	quickFadeIn     = (&easing.CubicOutEaseProvider{}).New(nil)
	quickFadeOut    = (&easing.CubicInEaseProvider{}).New(nil)
)

const (
	openDuration  = 650 * time.Millisecond
	closeDuration = 280 * time.Millisecond
)

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
		o.setOpen(!o.open, time.Now())
		o.resetCompletion()
		return true
	}
	if !o.open {
		return o.progress > 0
	}
	if frame.Actions["cancel"].Pressed {
		o.setOpen(false, time.Now())
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
	o.updateAnimation(time.Now())
	if o.progress <= 0 {
		return
	}
	panelHeight := int32(max(220, height*3/5))
	positionProgress, opacity := o.presentation()
	offsetY := int32(-float64(panelHeight) * (1 - positionProgress))
	rl.DrawRectangle(0, offsetY, int32(width), panelHeight, fade(rl.NewColor(8, 7, 6, 238), opacity))
	rl.DrawRectangleLinesEx(rl.NewRectangle(0, float32(offsetY), float32(width), float32(panelHeight)), 2, fade(rl.NewColor(176, 119, 38, 255), opacity))
	policy := o.session.Policy()
	mode := "read-only"
	if policy.Mutable {
		mode = "mutable"
	}
	o.drawText("DARK MAGIC LUA SHELL", 16, offsetY+12, 22, fade(rl.NewColor(222, 163, 58, 255), opacity))
	o.drawText(fmt.Sprintf("target %s | policy %s (%s)", o.session.Target(), policy.Name, mode), 16, offsetY+40, 16, fade(rl.Gray, opacity))

	transcriptTop := offsetY + 68
	promptTop := offsetY + panelHeight - 58
	available := int((promptTop-transcriptTop)/lineHeight) - 1
	lines := o.timeline(width)
	if len(lines) > available {
		lines = lines[len(lines)-available:]
	}
	for index, line := range lines {
		color := rl.LightGray
		if line.error {
			color = rl.NewColor(255, 104, 88, 255)
		} else if line.warning {
			color = rl.NewColor(245, 190, 75, 255)
		} else if line.result {
			color = rl.NewColor(125, 211, 167, 255)
		} else if line.dim {
			color = rl.Gray
		}
		o.drawText(line.text, 18, transcriptTop+int32(index)*lineHeight, fontSize, fade(color, opacity))
	}
	status := ""
	if o.busy {
		status = " [evaluating]"
	}
	o.drawText("> "+o.input+status, 16, promptTop, fontSize, fade(rl.RayWhite, opacity))
	o.drawText("` close  Enter run  Shift+Enter newline  Tab complete  Up/Down history  Esc close", 16, offsetY+panelHeight-26, 14, fade(rl.Gray, opacity))
}

func (o *Overlay) timeline(width int) []transcriptLine {
	columns := max(12, (width-36)/11)
	revision := o.session.TimelineRevision()
	if revision != o.displayRevision || columns != o.displayColumns {
		o.displayLines = wrapTranscript(timelineLines(o.session.Timeline()), columns)
		o.displayRevision = revision
		o.displayColumns = columns
	}
	return o.displayLines
}

func (o *Overlay) presentation() (position, opacity float64) {
	if o.open {
		return elasticEntrance(o.progress), quickFadeIn(o.progress)
	}
	return smoothExit(o.progress), quickFadeOut(o.progress)
}

func (o *Overlay) setOpen(open bool, now time.Time) {
	o.updateAnimation(now)
	o.open = open
	o.animationAt = now
}

func (o *Overlay) updateAnimation(now time.Time) {
	if o.animationAt.IsZero() {
		o.animationAt = now
		return
	}
	elapsed := now.Sub(o.animationAt)
	o.animationAt = now
	if elapsed <= 0 {
		return
	}
	if o.open {
		o.progress = min(1, o.progress+float64(elapsed)/float64(openDuration))
	} else {
		o.progress = max(0, o.progress-float64(elapsed)/float64(closeDuration))
	}
}

func fade(color rl.Color, opacity float64) rl.Color {
	color.A = uint8(float64(color.A) * max(0, min(1, opacity)))
	return color
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
	text                   string
	result, error, warning bool
	dim                    bool
}

func timelineLines(events []shell.TimelineEvent) []transcriptLine {
	lines := make([]transcriptLine, 0, len(events))
	for _, event := range events {
		line := transcriptLine{text: strings.ReplaceAll(event.Text, "\n", " ")}
		switch event.Kind {
		case "command":
			line.text = "> " + line.text
		case "value":
			line.result = true
		case "error", "log-error":
			line.error = true
		case "log-warn":
			line.warning = true
		case "log-debug":
			line.dim = true
		}
		lines = append(lines, line)
	}
	return lines
}

func wrapTranscript(lines []transcriptLine, columns int) []transcriptLine {
	var wrapped []transcriptLine
	for _, line := range lines {
		runes := []rune(line.text)
		for len(runes) > columns {
			wrapped = append(wrapped, transcriptLine{
				text: string(runes[:columns]), result: line.result, error: line.error, warning: line.warning, dim: line.dim,
			})
			runes = runes[columns:]
		}
		wrapped = append(wrapped, transcriptLine{text: string(runes), result: line.result, error: line.error})
	}
	return wrapped
}
