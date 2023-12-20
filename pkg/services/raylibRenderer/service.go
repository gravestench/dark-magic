package raylibRenderer

import (
	"log/slog"

	"github.com/faiface/mainthread"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	mesh   servicemesh.Mesh
	logger *slog.Logger

	config *configFile.Config
	cache  *cache.Cache

	cameras map[string]*rl.Camera2D

	rootNode Renderable

	isInit bool
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.cameras = make(map[string]*rl.Camera2D)
	s.rootNode = s.NewRenderable()
	s.rootNode.Disable() // dont render

	go s.initRenderer()
}

func (s *Service) IsInit() bool {
	return s.isInit
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
	title := s.config.Group(groupKeyWindow).GetString(keyWindowTitle)
	width := s.config.Group(groupKeyWindow).GetInt(keyWindowWidth)
	height := s.config.Group(groupKeyWindow).GetInt(keyWindowHeight)

	s.isInit = true

	rl.SetTraceLogCallback(func(level int, msg string) {
		switch level {
		case 0, 1:
			s.logger.Debug(msg)
		case 2:
			s.logger.Info(msg)
		case 3:
			s.logger.Warn(msg)
		case 4:
			s.logger.Error(msg)
			panic(msg)
		}
	})

	var serviceMeshShuttingDown bool

	s.mesh.Events().On(servicemesh.EventServiceMeshShutdownInitiated, func(_ ...any) {
		serviceMeshShuttingDown = true
	})

	mainthread.Call(func() {
		rl.InitWindow(int32(width), int32(height), title)
		rl.SetTargetFPS(60)
		rl.HideCursor()

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
}
