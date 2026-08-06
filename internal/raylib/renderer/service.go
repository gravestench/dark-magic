package raylibRenderer

import (
	"log/slog"
	"sync"
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/cache"
)

type Service struct {
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

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
