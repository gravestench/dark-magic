package guiManager

import (
	"fmt"
	"image"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

var _ managesTileSprites = &Service{}

func (s *Service) NewTileSprite(config TileSpriteConfig) (ts *TileSprite, err error) {
	if len(config.Tiles) < 1 {
		return nil, fmt.Errorf("no tiles")
	}

	if len(config.TileConfigs) < 1 {
		return nil, fmt.Errorf("no tile configs")
	}

	node := s.renderer.NewRenderable()

	if config.Parent != nil {
		node.SetParent(config.Parent)
	}

	ts = &TileSprite{
		service:    s,
		config:     config,
		Renderable: node,
	}

	if err = ts.setupChildNodes(); err != nil {
		return nil, fmt.Errorf("setting up child nodes: %v", err)
	}

	return
}

type TileSpriteConfig struct {
	Parent      raylibRenderer.Renderable
	Tiles       []image.Image
	TileConfigs []TileConfig // tile indexes, top-left to bottom-right
}

type TileSprite struct {
	service    *Service
	config     TileSpriteConfig
	Renderable raylibRenderer.Renderable
}

type TileConfig struct {
	Index    int
	Position image.Point
}

func (ts *TileSprite) setupChildNodes() error {
	for index, tcfg := range ts.config.TileConfigs {
		if tcfg.Index >= len(ts.config.TileConfigs) {
			return fmt.Errorf("invalid tile config: index %d: %+v", index, tcfg)
		}

		tile := ts.config.Tiles[tcfg.Index]

		tNode := ts.service.renderer.NewRenderable()
		tNode.SetOrigin(0, 0)
		tNode.SetImage(tile)

		x, y := float32(tcfg.Position.X), float32(tcfg.Position.Y)
		tNode.SetPosition(x, y)
		tNode.SetParent(ts.config.Parent)
	}

	return nil
}
