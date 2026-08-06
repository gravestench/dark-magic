package raylibRenderer

import (
	"encoding/json"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "raylib_renderer.json"
}

type Config struct {
	Window struct {
		Title         string
		Width, Height int
		Fullscreen    bool
		Borderless    bool
	}
	Resolution struct {
		Width, Height int
	}
	Cache struct {
		BudgetMB int
	}
}

func (s *Service) DefaultConfigData() []byte {
	cfg := DefaultConfig()

	data, _ := json.MarshalIndent(&cfg, "", "\t")

	return data
}

// DefaultConfig returns a complete renderer configuration without requiring
// the legacy configuration service.
func DefaultConfig() Config {
	var cfg Config
	cfg.Window.Title = "Dark Magic"
	cfg.Window.Width = 800
	cfg.Window.Height = 600
	cfg.Resolution.Width = 800
	cfg.Resolution.Height = 600
	cfg.Cache.BudgetMB = 100
	return cfg
}

// Configure explicitly supplies renderer configuration and cache ownership.
func (s *Service) Configure(config Config) {
	s.config = &config
	textureCache := cache.New(s.CacheBudget())
	textureCache.SetEvictionHandler(func(value interface{}) {
		if texture, ok := value.(rl.Texture2D); ok && texture.ID != 0 {
			rl.UnloadTexture(texture)
		}
	})
	s.FlushCache(textureCache)
}

func (s *Service) IngestConfig(config *configManager.ConfigHandle) error {
	data, err := config.Data()
	if err != nil {
		return fmt.Errorf("getting config data: %v", err)
	}

	if err = json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("unmarshalling config: %v", err)
	}

	return nil
}
