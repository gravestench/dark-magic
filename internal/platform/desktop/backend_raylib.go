//go:build !ebitengine

package desktop

import (
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/audio"
	raylibinput "github.com/gravestench/dark-magic/internal/platform/raylib/input"
	raylibrenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
	raylibshell "github.com/gravestench/dark-magic/internal/shell/raylib"
)

type raylibRenderer struct {
	*raylibrenderer.Service
	nativeAudio bool
}

// Name identifies the selected renderer without exposing the embedded Raylib service type.
func (*raylibRenderer) Name() string { return "raylib" }

// BackendDiagnostics delegates to the native service so the desktop-neutral interface retains Raylib's snapshot.
func (r *raylibRenderer) BackendDiagnostics() any { return r.Service.BackendDiagnostics() }

// CacheDiagnostics delegates to the native service without widening the desktop package's backend-neutral contract.
func (r *raylibRenderer) CacheDiagnostics() any { return r.Service.CacheDiagnostics() }

// AttachAudio either gives commands to Raylib's native backend or drains them into a sink every frame. Muted mode must
// still drain commands so mixer ownership advances instead of accumulating an unbounded pending queue.
func (r *raylibRenderer) AttachAudio(mixer *audio.Mixer) error {
	if r.nativeAudio {
		return r.Service.AttachAudio(mixer)
	}

	r.SubscribeFrame(func() {
		if err := mixer.Drain(discardAudioBackend{}); err != nil {
			r.Service.Logger().Error("draining muted Raylib audio", "error", err)
		}
	})

	return nil
}

// New translates backend-neutral desktop options into Raylib configuration and wires input to the same renderer;
// callers receive no Raylib-specific types.
func New(options Options) (*Bundle, error) {
	if options.Logger == nil {
		options.Logger = slog.Default()
	}

	service := &raylibrenderer.Service{}
	service.SetLogger(options.Logger)

	config := raylibrenderer.DefaultConfig()
	if options.WindowTitle != "" {
		config.Window.Title = options.WindowTitle
	}

	config.Window.Width = options.WindowWidth
	config.Window.Height = options.WindowHeight
	config.Window.Borderless = options.BorderlessFullscreen
	config.Window.ShowSystemCursor = options.ShowSystemCursor
	config.Resolution.Width = options.LogicalWidth
	config.Resolution.Height = options.LogicalHeight
	config.Resolution.Fit = options.ViewportFit
	service.Configure(config)

	if options.PalettePath != "" {
		if options.Content == nil {
			return nil, fmt.Errorf("raylib backend: palette content is required")
		}

		if err := service.ConfigurePaletteQuantization(
			options.Content,
			options.PalettePath,
		); err != nil {
			return nil, err
		}
	}

	renderer := &raylibRenderer{Service: service, nativeAudio: options.NativeAudio}
	input := raylibinput.New(service)
	input.SetLogger(options.Logger.With("component", "input"))

	return &Bundle{Renderer: renderer, Input: input}, nil
}

// NewConsole shares the existing shell session and settings with Raylib's native overlay rather than creating a
// second command history.
func NewConsole(options ConsoleOptions) Console {
	return raylibshell.New(options.Session, options.Settings)
}

var (
	_ Renderer = (*raylibRenderer)(nil)
	_ Input    = (*raylibinput.Service)(nil)
	_ Console  = (*raylibshell.Overlay)(nil)
)
