package raylibRenderer

import (
	"context"
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

	audioMu      sync.Mutex
	audioBackend *raylibAudioBackend
}

// OnFrame registers work that must run on the renderer thread, immediately
// before scene graph updates. Raylib window and input calls belong here.
func (s *Service) OnFrame(callback func()) {
	_ = s.SubscribeFrame(callback)
}

// SubscribeFrame registers renderer-thread work and returns a safe idempotent
// cancellation function.
func (s *Service) SubscribeFrame(callback func()) func() {
	if callback == nil {
		return func() {}
	}
	var active atomic.Bool
	active.Store(true)
	wrapper := func() {
		if active.Load() {
			callback()
		}
	}
	s.frameMux.Lock()
	s.frameCallbacks = append(s.frameCallbacks, wrapper)
	s.frameSnapshot.Store(append([]func(){}, s.frameCallbacks...))
	s.frameMux.Unlock()
	return func() { active.Store(false) }
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.mesh.Events().On(servicemesh.EventServiceMeshShutdownInitiated, func(_ ...any) {
		cancel()
	})

	for {
		defer func() {
			if err := recover(); err != nil {
				//s.logger.Error("init renderer", "error", err)
			}
		}()

		mainthread.Call(func() {
			if err := s.Start(context.Background()); err != nil {
				s.logger.Error("starting renderer", "error", err)
				return
			}
			_ = s.Run(ctx)
			_ = s.Stop(context.Background())
			s.mesh.Shutdown()
		})

		break
	}
}
