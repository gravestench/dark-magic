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

func (*raylibRenderer) Name() string { return "raylib" }

func (r *raylibRenderer) BackendDiagnostics() any { return r.Service.BackendDiagnostics() }

func (r *raylibRenderer) CacheDiagnostics() any { return r.Service.CacheDiagnostics() }

func (r *raylibRenderer) AttachAudio(mixer *audio.Mixer) error {
	if r.nativeAudio {
		return r.Service.AttachAudio(mixer)
	}
	r.Service.SubscribeFrame(func() {
		if err := mixer.Drain(discardAudioBackend{}); err != nil {
			r.Service.Logger().Error("draining muted Raylib audio", "error", err)
		}
	})
	return nil
}

// New constructs the default Raylib backend without exposing it to clientapp.
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
		if err := service.ConfigurePaletteQuantization(options.Content, options.PalettePath); err != nil {
			return nil, err
		}
	}
	renderer := &raylibRenderer{Service: service, nativeAudio: options.NativeAudio}
	input := raylibinput.New(service)
	input.SetLogger(options.Logger.With("component", "input"))
	return &Bundle{Renderer: renderer, Input: input}, nil
}

// NewConsole constructs the Raylib developer-console adapter.
func NewConsole(options ConsoleOptions) Console {
	return raylibshell.New(options.Session, options.Settings)
}

var (
	_ Renderer = (*raylibRenderer)(nil)
	_ Input    = (*raylibinput.Service)(nil)
	_ Console  = (*raylibshell.Overlay)(nil)
)
