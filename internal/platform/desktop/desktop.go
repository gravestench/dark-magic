// Package desktop selects the native interactive client backend at build time.
//
// The default build uses Raylib. Building with the ebitengine tag selects the
// Ebitengine implementation. Gameplay and Lua presentation code depend only on
// the backend-neutral contracts in this file.
package desktop

import (
	"context"
	"io/fs"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/shell"
)

// Options contains backend-independent window and logical-surface policy.
type Options struct {
	Content              fs.FS
	PalettePath          string
	WindowTitle          string
	WindowWidth          int
	WindowHeight         int
	LogicalWidth         int
	LogicalHeight        int
	ViewportFit          string
	BorderlessFullscreen bool
	ShowSystemCursor     bool
	NativeAudio          bool
	Logger               *slog.Logger
}

// DefaultOptions returns the ordinary 800x600 desktop client policy.
func DefaultOptions() Options {
	return Options{
		WindowTitle:      "Dark Magic",
		WindowWidth:      800,
		WindowHeight:     600,
		LogicalWidth:     800,
		LogicalHeight:    600,
		ViewportFit:      "contain",
		ShowSystemCursor: false,
		NativeAudio:      true,
		Logger:           slog.Default(),
	}
}

// Renderer is the complete native surface required by the client composition
// root. Native textures, shaders, windows, and event-loop types do not cross it.
type Renderer interface {
	Start(context.Context) error
	Stop(context.Context) error
	Run(context.Context) error
	Name() string
	SubscribeFrame(func()) func()
	SubscribePostFrame(func()) func()
	SubscribeOverlay(func()) func()
	SubscribeViewport(func(width, height int)) func()
	AttachComposer(*render.Composer) error
	AttachAudio(*audio.Mixer) error
	CaptureScreenshot(string) error
	WindowSize() (width, height int)
	Resolution() (width, height int)
	ScreenToGame(x, y int) (gameX, gameY int, inside bool)
	SetWindowTitle(string)
	SetResidencyDebug(bool)
	SetTextureUploadBudget(uint64)
	SetTextureCacheBudget(uint64)
	BackendDiagnostics() any
	CacheDiagnostics() any
}

// Input publishes one backend-neutral snapshot per native frame.
type Input interface {
	Start(context.Context) error
	Stop(context.Context) error
	Snapshot() inputstate.Frame
}

// Console keeps shell policy independent from its native drawing adapter.
type Console interface {
	LoadFont() error
	Handle(context.Context, inputstate.Frame) bool
	Draw(width, height int)
	Close()
}

// Bundle owns the renderer and its input sampler.
type Bundle struct {
	Renderer Renderer
	Input    Input
}

// ConsoleOptions contains the shared shell objects used by native adapters.
type ConsoleOptions struct {
	Session  *shell.Session
	Settings *shell.Settings
}

type discardAudioBackend struct{}

func (discardAudioBackend) Apply(audio.Command) error { return nil }
