package raylibRenderer

import (
	"github.com/faiface/mainthread"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	logger *zerolog.Logger

	cfg   configFile.Dependency
	cache *cache.Cache

	cameras map[string]*rl.Camera2D

	rootNode Renderable

	isInit bool
}

func (s *Service) Init(rt runtime.Runtime) {
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

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) initRenderer() {
	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	title := cfg.Group(groupKeyWindow).GetString(keyWindowTitle)
	width := cfg.Group(groupKeyWindow).GetInt(keyWindowWidth)
	height := cfg.Group(groupKeyWindow).GetInt(keyWindowHeight)

	s.isInit = true

	rl.SetTraceLogCallback(func(level int, msg string) {
		switch level {
		case 0:
			s.logger.Trace().Msg(msg)
		case 1:
			s.logger.Debug().Msg(msg)
		case 2:
			s.logger.Info().Msg(msg)
		case 3:
			s.logger.Warn().Msg(msg)
		case 4:
			s.logger.Fatal().Msg(msg)
		}
	})

	mainthread.Call(func() {
		rl.InitWindow(int32(width), int32(height), title)
		rl.SetTargetFPS(60)
		rl.HideCursor()

		for !rl.WindowShouldClose() {
			rl.BeginDrawing()
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(*s.GetDefaultCamera())
			s.update()
			s.render()
			rl.EndMode2D()

			rl.EndDrawing()
		}

		rl.CloseWindow()
	})
}
