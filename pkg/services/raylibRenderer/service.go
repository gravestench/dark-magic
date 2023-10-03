package raylibRenderer

import (
	"sort"

	"github.com/faiface/mainthread"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	logger *zerolog.Logger

	cfg configFile.Dependency

	renderableLayers []IsRenderableLayer

	init bool
}

func (s *Service) WindowSize() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

func (s *Service) Resolution() (width, height int) {
	return rl.GetRenderWidth(), rl.GetRenderHeight()
}

func (s *Service) Init(rt runtime.Runtime) {
	s.initRenderer()

	for _, service := range rt.Services() {
		if candidate, ok := service.(IsRenderableLayer); ok {
			s.renderableLayers = append(s.renderableLayers, candidate)
		}
	}
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

func (s *Service) IsInit() bool {
	return s.init
}

func (s *Service) initRenderer() {
	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	title := cfg.Group(groupKeyWindow).GetString(keyWindowTitle)
	width := cfg.Group(groupKeyWindow).GetInt(keyWindowWidth)
	height := cfg.Group(groupKeyWindow).GetInt(keyWindowHeight)

	mainthread.Call(func() {
		rl.InitWindow(int32(width), int32(height), title)
		rl.SetTargetFPS(24)
		s.init = true

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

func (s *Service) drawRenderableLayers() {
	sort.Slice(s.renderableLayers, func(i, j int) bool {
		return s.renderableLayers[i].LayerIndex() > s.renderableLayers[j].LayerIndex()
	})

	for _, layer := range s.renderableLayers {
		tx := rl.LoadTextureFromImage(rl.NewImageFromImage(layer.LayerImage()))
		x, y := layer.Position()
		rl.DrawTextureEx(tx, rl.Vector2{float32(x), float32(y)}, layer.Rotation(), layer.Scale(), rl.NewColor(255, 255, 255, uint8(layer.Opacity()*255)))
	}
}
