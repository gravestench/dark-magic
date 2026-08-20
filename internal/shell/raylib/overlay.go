// Package raylib presents a shared shell session inside the graphical client.
package raylibshell

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/easing"
	"github.com/gravestench/dark-magic/internal/shell"
	"golang.org/x/image/font/gofont/gomono"
)

type viewMode uint8

const (
	viewLua viewMode = iota
	viewLogs
)

// Overlay adapts the shared session to raylib input and immediate-mode debug
// drawing. Evaluation remains serialized by Session and its evaluator.
type Overlay struct {
	session         *shell.Session
	settings        *shell.Settings
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
	displayLimit    int
	displayLines    []transcriptLine
	displayView     viewMode
	view            viewMode
	logOffset       int
	luaOffset       int
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

// New binds one shared session and optional settings override. The buffered
// completion channel lets one evaluation finish without blocking its goroutine
// until the next input frame polls it.
func New(session *shell.Session, configured ...*shell.Settings) *Overlay {
	var settings *shell.Settings
	if len(configured) > 0 {
		settings = configured[0]
	}

	if settings == nil {
		settings = session.Settings()
	}

	return &Overlay{
		session:  session,
		settings: settings,
		history:  len(session.History()),
		finished: make(chan shell.Entry, 1),
	}
}

// Open reports the requested modal state; progress may remain non-zero while a
// closed overlay finishes its exit animation and still captures scene input.
func (o *Overlay) Open() bool { return o.open }

// LoadFont creates the GPU atlas on the renderer owner thread. The embedded Go
// Mono bytes keep console typography independent of host-installed fonts.
func (o *Overlay) LoadFont() error {
	if o.fontLoaded {
		return nil
	}

	codepoints := make([]rune, 0, 224)
	for current := rune(32); current <= 255; current++ {
		codepoints = append(codepoints, current)
	}

	// Rasterize above the largest supported setting so live font increases do not
	// upscale a small glyph atlas.
	o.font = rl.LoadFontFromMemory(".ttf", gomono.TTF, 64, codepoints)
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

// presentation maps linear progress through separate position and opacity
// curves so opening can overshoot without making alpha overshoot.
func (o *Overlay) presentation() (position, opacity float64) {
	if o.open {
		return elasticEntrance(o.progress), quickFadeIn(o.progress)
	}

	return smoothExit(o.progress), quickFadeOut(o.progress)
}

// setOpen first advances the current animation to the event time, preventing a
// direction change from discarding elapsed progress.
func (o *Overlay) setOpen(open bool, now time.Time) {
	o.updateAnimation(now)
	o.open = open
	o.animationAt = now
}

// updateAnimation advances clamped progress using live animation speed. The
// first call establishes a timestamp so construction time cannot create a jump.
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

	speed := o.settings.Values().AnimationSpeed
	if o.open {
		o.progress = min(1, o.progress+float64(elapsed)/float64(openDuration)*speed)
	} else {
		o.progress = max(0, o.progress-float64(elapsed)/float64(closeDuration)*speed)
	}
}
