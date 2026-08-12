package raylibRenderer

import (
	"log/slog"
	"sync"
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/cache"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// Service owns the Raylib window, GPU resources, audio device, and native owner
// thread. Other packages submit backend-neutral state or subscribe callbacks;
// they never receive Raylib handles or call native APIs directly.
type Service struct {
	logger *slog.Logger

	config *Config
	cache  *cache.Cache

	camera   rl.Camera2D
	rootNode *node

	isInit             atomic.Bool
	frameMux           sync.Mutex
	frameCallbacks     []func()
	frameSnapshot      atomic.Value
	postFrameMux       sync.Mutex
	postFrameCallbacks []func()
	postFrameSnapshot  atomic.Value
	overlayMux         sync.Mutex
	overlayCallbacks   []func()
	overlaySnapshot    atomic.Value

	compositionMu      sync.Mutex
	composition        *render.Composer
	compositionBackend *compositionBackend

	audioMu      sync.Mutex
	audioBackend *raylibAudioBackend

	paletteQuantizer        *paletteQuantizer
	gameTarget              rl.RenderTexture2D
	frames                  atomic.Uint64
	drawCalls               atomic.Uint64
	nodesVisited            atomic.Uint64
	subtreesCulled          atomic.Uint64
	textureUpdates          atomic.Uint64
	textureUploads          atomic.Uint64
	textureUploadBytes      atomic.Uint64
	textureUploadBudget     atomic.Uint64
	textureCacheBudget      atomic.Uint64
	residencyDebug          atomic.Bool
	lastFrameDrawCalls      atomic.Uint64
	lastFrameNodesVisited   atomic.Uint64
	lastFrameSubtreesCulled atomic.Uint64
	lastFrameTextureUpdates atomic.Uint64
	lastFrameCompositionNS  atomic.Uint64
	lastFrameRenderNS       atomic.Uint64
	lastFrameUploadNS       atomic.Uint64
	textureUploadNS         atomic.Uint64
}

// BackendDiagnostics reports native-adapter work without exposing Raylib
// handles. Values are cumulative for the renderer lifetime.
type BackendDiagnostics struct {
	Frames, DrawCalls, NodesVisited, SubtreesCulled, TextureUpdates uint64
	LastFrameDrawCalls, LastFrameNodesVisited                       uint64
	LastFrameSubtreesCulled, LastFrameTextureUpdates                uint64
	LastFrameCompositionNS, LastFrameRenderNS, LastFrameUploadNS    uint64
	TextureUploadNS                                                 uint64
}

// BackendDiagnostics returns lock-free cumulative counters suitable for overlays.
func (s *Service) BackendDiagnostics() BackendDiagnostics {
	return BackendDiagnostics{
		Frames: s.frames.Load(), DrawCalls: s.drawCalls.Load(),
		NodesVisited: s.nodesVisited.Load(), SubtreesCulled: s.subtreesCulled.Load(),
		TextureUpdates:          s.textureUpdates.Load(),
		LastFrameDrawCalls:      s.lastFrameDrawCalls.Load(),
		LastFrameNodesVisited:   s.lastFrameNodesVisited.Load(),
		LastFrameSubtreesCulled: s.lastFrameSubtreesCulled.Load(),
		LastFrameTextureUpdates: s.lastFrameTextureUpdates.Load(),
		LastFrameCompositionNS:  s.lastFrameCompositionNS.Load(),
		LastFrameRenderNS:       s.lastFrameRenderNS.Load(),
		LastFrameUploadNS:       s.lastFrameUploadNS.Load(),
		TextureUploadNS:         s.textureUploadNS.Load(),
	}
}

// SubscribeOverlay registers owner-thread drawing after scene composition and
// display quantization but before the back buffer is presented.
func (s *Service) SubscribeOverlay(callback func()) func() {
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
	s.overlayMux.Lock()
	s.overlayCallbacks = append(s.overlayCallbacks, wrapper)
	s.overlaySnapshot.Store(append([]func(){}, s.overlayCallbacks...))
	s.overlayMux.Unlock()
	return func() { active.Store(false) }
}

func (s *Service) runOverlays() {
	snapshot := s.overlaySnapshot.Load()
	if snapshot == nil {
		return
	}
	for _, callback := range snapshot.([]func()) {
		callback()
	}
}

// SubscribePostFrame registers owner-thread work after the fully composed frame
// has been presented. Screenshot and visual-inspection tools belong here.
func (s *Service) SubscribePostFrame(callback func()) func() {
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
	s.postFrameMux.Lock()
	s.postFrameCallbacks = append(s.postFrameCallbacks, wrapper)
	s.postFrameSnapshot.Store(append([]func(){}, s.postFrameCallbacks...))
	s.postFrameMux.Unlock()
	return func() { active.Store(false) }
}

func (s *Service) runPostFrame() {
	snapshot := s.postFrameSnapshot.Load()
	if snapshot == nil {
		return
	}
	for _, callback := range snapshot.([]func()) {
		callback()
	}
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
