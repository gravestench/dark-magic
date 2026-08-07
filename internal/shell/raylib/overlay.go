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

type viewMode uint8

const (
	viewLua viewMode = iota
	viewLogs
)

// Overlay adapts the shared session to Raylib input and immediate-mode debug
// drawing. Evaluation remains serialized by Session and its Lua runtime.
type Overlay struct {
	session         *shell.Session
	open            bool
	busy            bool
	input           string
	cursor          int
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
	displayView     viewMode
	view            viewMode
	logOffset       int
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
	if o.view == viewLogs {
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
		return true
	}
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
	o.drawText("DARK MAGIC CONSOLE", 16, offsetY+12, 22, fade(rl.NewColor(222, 163, 58, 255), opacity))
	o.drawText(fmt.Sprintf("target %s | policy %s (%s)", o.session.Target(), policy.Name, mode), 16, offsetY+40, 16, fade(rl.Gray, opacity))
	o.drawTabs(offsetY, opacity)

	transcriptTop := offsetY + 94
	promptTop := offsetY + panelHeight - 58
	contentBottom := promptTop
	completion := o.completionLines()
	if o.view == viewLua {
		contentBottom -= int32(len(completion) * 18)
	}
	if o.view == viewLogs {
		contentBottom = offsetY + panelHeight - 34
	}
	available := max(1, int((contentBottom-transcriptTop)/lineHeight)-1)
	lines := o.timeline(width)
	end := len(lines)
	if o.view == viewLogs {
		o.logOffset = min(o.logOffset, max(0, len(lines)-available))
		end = max(0, len(lines)-o.logOffset)
	}
	start := max(0, end-available)
	if start < end {
		lines = lines[start:end]
	} else {
		lines = nil
	}
	for index, line := range lines {
		color := rl.LightGray
		if line.error {
			color = rl.NewColor(255, 104, 88, 255)
		} else if line.warning {
			color = rl.NewColor(245, 190, 75, 255)
		} else if line.notice {
			color = rl.NewColor(224, 183, 92, 255)
		} else if line.result {
			color = rl.NewColor(125, 211, 167, 255)
		} else if line.dim {
			color = rl.Gray
		}
		o.drawText(line.text, 18, transcriptTop+int32(index)*lineHeight, fontSize, fade(color, opacity))
	}
	if o.view == viewLua {
		for index, line := range completion {
			color := rl.Gray
			if line.selected {
				color = rl.NewColor(222, 163, 58, 255)
			}
			o.drawText(line.text, 24, contentBottom+int32(index*18), 14, fade(color, opacity))
		}
		status := ""
		if o.busy {
			status = " [evaluating]"
		}
		o.drawText("> "+o.inputWithCaret()+status, 16, promptTop, fontSize, fade(rl.RayWhite, opacity))
		o.drawText("F1 Lua  F2 Logs  ` close  Enter run  Shift+Enter newline  Tab complete", 16, offsetY+panelHeight-26, 14, fade(rl.Gray, opacity))
	} else {
		o.drawText("F1 Lua  F2 Logs  Up/Down scroll  PgUp/PgDn page  ` or Esc close", 16, offsetY+panelHeight-26, 14, fade(rl.Gray, opacity))
	}
}

type completionLine struct {
	text     string
	selected bool
}

func (o *Overlay) completionLines() []completionLine {
	if len(o.candidates) == 0 {
		return nil
	}
	limit := min(5, len(o.candidates))
	start := 0
	if o.candidateAt >= limit {
		start = min(len(o.candidates)-limit, o.candidateAt-limit/2)
	}
	lines := make([]completionLine, 0, limit)
	for index := start; index < start+limit; index++ {
		candidate := o.candidates[index]
		lines = append(lines, completionLine{
			text: fmt.Sprintf("%s  %s", candidate.Value, candidate.Detail), selected: index == o.candidateAt,
		})
	}
	return lines
}

func (o *Overlay) drawTabs(offsetY int32, opacity float64) {
	luaColor, logColor := rl.Gray, rl.Gray
	if o.view == viewLua {
		luaColor = rl.NewColor(222, 163, 58, 255)
	} else {
		logColor = rl.NewColor(222, 163, 58, 255)
	}
	o.drawText("[F1 LUA]", 16, offsetY+68, 17, fade(luaColor, opacity))
	o.drawText("[F2 LOGS]", 116, offsetY+68, 17, fade(logColor, opacity))
}

func (o *Overlay) timeline(width int) []transcriptLine {
	columns := max(12, (width-36)/11)
	revision := o.session.TimelineRevision()
	if revision != o.displayRevision || columns != o.displayColumns || o.view != o.displayView {
		events := o.session.TranscriptTimeline()
		if o.view == viewLogs {
			events = o.session.LogTimeline()
		}
		o.displayLines = wrapTranscript(timelineLines(events), columns)
		o.displayRevision = revision
		o.displayColumns = columns
		o.displayView = o.view
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
	o.cursor = 0
	o.resetCompletion()
	go func() { o.finished <- o.session.Submit(ctx, source) }()
}

func (o *Overlay) complete(ctx context.Context, reverse bool) {
	if len(o.candidates) == 0 {
		runes := []rune(o.input)
		o.cursor = max(0, min(len(runes), o.cursor))
		prefix := string(runes[:o.cursor])
		original := completionToken(prefix)
		candidates, err := o.session.Complete(ctx, prefix)
		if err != nil || len(candidates) == 0 {
			return
		}
		o.candidates, o.candidateAt = candidates, -1
		shared := shell.SharedPrefix(candidates)
		if shared != "" {
			o.replaceToken(shared)
		}
		if len(candidates) > 1 && shared != original {
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
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))
	prefix := string(runes[:o.cursor])
	token := completionToken(prefix)
	before := []rune(strings.TrimSuffix(prefix, token))
	replacement := []rune(value)
	o.input = string(append(append(before, replacement...), runes[o.cursor:]...))
	o.cursor = len(before) + len(replacement)
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
	o.cursor = utf8.RuneCountInString(o.input)
	o.resetCompletion()
}

func (o *Overlay) resetCompletion() {
	o.candidates = nil
	o.candidateAt = -1
}

func (o *Overlay) insert(value string) {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))
	inserted := []rune(value)
	o.input = string(append(append(runes[:o.cursor:o.cursor], inserted...), runes[o.cursor:]...))
	o.cursor += len(inserted)
}

func (o *Overlay) backspace() {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))
	if o.cursor == 0 {
		return
	}
	o.input = string(append(runes[:o.cursor-1], runes[o.cursor:]...))
	o.cursor--
}

func (o *Overlay) delete() {
	runes := []rune(o.input)
	o.cursor = max(0, min(len(runes), o.cursor))
	if o.cursor == len(runes) {
		return
	}
	o.input = string(append(runes[:o.cursor], runes[o.cursor+1:]...))
}

func (o *Overlay) inputWithCaret() string {
	runes := []rune(o.input)
	cursor := max(0, min(len(runes), o.cursor))
	return string(append(append(runes[:cursor:cursor], '|'), runes[cursor:]...))
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
	dim, notice            bool
}

func timelineLines(events []shell.TimelineEvent) []transcriptLine {
	lines := make([]transcriptLine, 0, len(events))
	for _, event := range events {
		parts := strings.Split(event.Text, "\n")
		for index, part := range parts {
			line := transcriptLine{text: part}
			switch event.Kind {
			case "motd":
				line.notice = true
			case "command":
				if index == 0 {
					line.text = "> " + line.text
				}
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
	}
	return lines
}

func wrapTranscript(lines []transcriptLine, columns int) []transcriptLine {
	var wrapped []transcriptLine
	for _, line := range lines {
		runes := []rune(line.text)
		for len(runes) > columns {
			wrapped = append(wrapped, transcriptLine{
				text: string(runes[:columns]), result: line.result, error: line.error, warning: line.warning, dim: line.dim, notice: line.notice,
			})
			runes = runes[columns:]
		}
		wrapped = append(wrapped, transcriptLine{
			text: string(runes), result: line.result, error: line.error, warning: line.warning, dim: line.dim, notice: line.notice,
		})
	}
	return wrapped
}
