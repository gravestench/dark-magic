package raylibRenderer

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/faiface/mainthread"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/cache"
)

type Service struct {
	mesh   servicemesh.Mesh
	logger *slog.Logger

	config *Config
	cache  *cache.Cache

	cameras map[string]*rl.Camera2D

	rootNode Renderable

	isInit         atomic.Bool
	frameMux       sync.Mutex
	frameCallbacks []func()
	frameSnapshot  atomic.Value

	compositionMu      sync.Mutex
	composition        *rendercore.Composer
	compositionBackend *compositionBackend
}

// OnFrame registers work that must run on the renderer thread, immediately
// before scene graph updates. Raylib window and input calls belong here.
func (s *Service) OnFrame(callback func()) {
	if callback == nil {
		return
	}
	s.frameMux.Lock()
	s.frameCallbacks = append(s.frameCallbacks, callback)
	s.frameSnapshot.Store(append([]func(){}, s.frameCallbacks...))
	s.frameMux.Unlock()
}

func (s *Service) DependenciesResolved() bool {
	return s.config != nil
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	// noop
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.cameras = make(map[string]*rl.Camera2D)
	s.rootNode = s.NewRenderable()
	s.rootNode.Disable() // dont render

	go s.initRenderer()
}

func (s *Service) IsInit() bool {
	return s.isInit.Load()
}

func (s *Service) Name() string {
	return "Renderer"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	return true
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) initRenderer() {
	title := s.config.Window.Title
	width := s.config.Window.Width
	height := s.config.Window.Height

	rl.SetTraceLogCallback(func(level int, msg string) {
		switch level {
		case 0, 1, 2, 3:
			s.logger.Debug(msg)
		case 4:
			s.logger.Error(msg)
			panic(msg)
		}
	})

	var serviceMeshShuttingDown bool

	s.mesh.Events().On(servicemesh.EventServiceMeshShutdownInitiated, func(_ ...any) {
		serviceMeshShuttingDown = true
	})

	for {
		defer func() {
			if err := recover(); err != nil {
				//s.logger.Error("init renderer", "error", err)
			}
		}()

		mainthread.Call(func() {
			rl.InitWindow(int32(width), int32(height), title)
			rl.InitAudioDevice()
			rl.SetTargetFPS(60)
			rl.HideCursor()
			s.isInit.Store(true)

			for !rl.WindowShouldClose() && !serviceMeshShuttingDown {
				rl.BeginDrawing()
				rl.ClearBackground(rl.Black)
				rl.BeginMode2D(*s.GetDefaultCamera())
				s.update()
				s.render()
				rl.EndMode2D()

				rl.EndDrawing()
			}

			rl.CloseWindow()
			s.mesh.Shutdown()
		})

		break
	}
}
