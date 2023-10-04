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

	renderableLayers []IsRenderableLayer

	isInit bool
}

func (s *Service) WindowSize() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

func (s *Service) Resolution() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

func (s *Service) Init(rt runtime.Runtime) {
	go s.initRenderer()

	for _, service := range rt.Services() {
		if candidate, ok := service.(IsRenderableLayer); ok {
			s.renderableLayers = append(s.renderableLayers, candidate)
		}
	}
}

func (s *Service) IsInit() bool {
	return s.isInit
}

func (s *Service) OnServiceAdded(args ...any) {
	if len(args) < 1 {
		return
	}

	service, ok := args[0].(runtime.Service)
	if !ok {
		return
	}

	layer, ok := service.(IsRenderableLayer)
	if !ok {
		return
	}

	s.renderableLayers = append(s.renderableLayers, layer)
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
			s.drawRenderableLayers()
			rl.EndDrawing()
		}

		rl.CloseWindow()
	})
}

func (s *Service) SetWindowTitle(title string) {
	rl.SetWindowTitle(title)
}
