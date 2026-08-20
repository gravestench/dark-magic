package raylibRenderer

import (
	"encoding/json"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/cache"
)

// Config contains native-window, logical-resolution, and GPU cache settings. It intentionally contains no raylib types,
// allowing composition roots to decode it before a native owner thread exists.
type Config struct {
	Window struct {
		Title            string
		Width, Height    int
		Fullscreen       bool
		Borderless       bool
		Resizable        bool
		ShowSystemCursor bool
	}
	Resolution struct {
		Width, Height int
		Fit           string
		Native        bool
	}
	Cache struct {
		BudgetMB int
	}
}

// DefaultConfigData serializes typed defaults for configuration-file generation.
// Config contains only supported JSON values, so marshaling cannot fail; the signature preserves the provider contract.
func (s *Service) DefaultConfigData() []byte {
	cfg := DefaultConfig()

	data, _ := json.MarshalIndent(&cfg, "", "\t")

	return data
}

// DefaultConfig returns a complete renderer configuration without requiring the legacy configuration service. Logical
// and initial window dimensions match so default input mapping requires no scaling.
func DefaultConfig() Config {
	var cfg Config

	cfg.Window.Title = "Dark Magic"
	cfg.Window.Width = 800
	cfg.Window.Height = 600
	cfg.Window.Resizable = true
	cfg.Resolution.Width = 800
	cfg.Resolution.Height = 600
	cfg.Resolution.Fit = "contain"
	cfg.Cache.BudgetMB = 512

	return cfg
}

// Configure explicitly supplies renderer configuration and creates its GPU texture cache. The eviction handler unloads
// native textures, which is why later budget application and cache clearing must occur on the renderer owner thread.
func (s *Service) Configure(config Config) {
	s.config = &config
	textureCache := cache.New(s.CacheBudget())
	textureCache.SetEvictionHandler(func(value interface{}) {
		if texture, ok := value.(rl.Texture2D); ok && texture.ID != 0 {
			rl.UnloadTexture(texture)
		}
	})
	s.FlushCache(textureCache)
	s.textureCacheBudget.Store(uint64(s.CacheBudget()))

	// Warm uploads are optional, but a non-zero default lets the composer incrementally populate textures without
	// allowing one frame to consume an unbounded upload budget.
	if s.textureUploadBudget.Load() == 0 {
		s.textureUploadBudget.Store(16 * 1024 * 1024)
	}
}
