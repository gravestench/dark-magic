package raylibRenderer

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// SetResidencyDebug toggles the owner-thread diagnostics overlay.
func (s *Service) SetResidencyDebug(enabled bool) { s.residencyDebug.Store(enabled) }

// SetTextureUploadBudget sets optional warm-upload work allowed per frame.
// Demand uploads remain unbounded because they are required to draw the scene.
func (s *Service) SetTextureUploadBudget(bytes uint64) {
	if bytes == 0 {
		bytes = 1
	}
	s.textureUploadBudget.Store(bytes)
}

// SetTextureCacheBudget requests a resident-texture budget change. The cache
// applies it on the graphics-owner thread because evictions unload GPU handles.
func (s *Service) SetTextureCacheBudget(bytes uint64) {
	if bytes == 0 {
		bytes = 1
	}
	s.textureCacheBudget.Store(bytes)
}

func (s *Service) applyTextureCacheBudget() error {
	if s.cache == nil {
		return nil
	}
	wanted := s.textureCacheBudget.Load()
	if wanted == 0 || uint64(s.cache.GetBudget()) == wanted {
		return nil
	}
	return s.cache.SetBudget(int(wanted))
}

type ResidencyDiagnostics struct {
	Entries, Weight, Budget            int
	Hits, Misses, Evictions            uint64
	Uploads, UploadBytes, UploadBudget uint64
	WarmPending                        int
	WarmPendingBytes                   uint64
	FrameDrawCalls, FrameNodesVisited  uint64
	FrameSubtreesCulled, FrameUpdates  uint64
}

func (s *Service) ResidencyDiagnostics(composer *render.Composer) ResidencyDiagnostics {
	result := ResidencyDiagnostics{
		Uploads: s.textureUploads.Load(), UploadBytes: s.textureUploadBytes.Load(),
		UploadBudget: s.textureUploadBudget.Load(),
	}
	backend := s.BackendDiagnostics()
	result.FrameDrawCalls = backend.LastFrameDrawCalls
	result.FrameNodesVisited = backend.LastFrameNodesVisited
	result.FrameSubtreesCulled = backend.LastFrameSubtreesCulled
	result.FrameUpdates = backend.LastFrameTextureUpdates
	if s.cache != nil {
		stats := s.cache.Diagnostics()
		result.Entries, result.Weight, result.Budget = stats.Entries, stats.Weight, stats.Budget
		result.Hits, result.Misses, result.Evictions = stats.Hits, stats.Misses, stats.Evictions
	}
	if composer != nil {
		stats := composer.Diagnostics()
		result.WarmPending, result.WarmPendingBytes = stats.WarmPending, stats.WarmPendingBytes
	}
	return result
}

func (s *Service) drawResidencyDebug(composer *render.Composer) {
	if !s.residencyDebug.Load() {
		return
	}
	stats := s.ResidencyDiagnostics(composer)
	lines := []string{
		fmt.Sprintf("frame draws=%d  nodes=%d  culled=%d  texture updates=%d", stats.FrameDrawCalls, stats.FrameNodesVisited, stats.FrameSubtreesCulled, stats.FrameUpdates),
		fmt.Sprintf("GPU textures  resident=%d  %.1f/%.1f MiB", stats.Entries, float64(stats.Weight)/(1024*1024), float64(stats.Budget)/(1024*1024)),
		fmt.Sprintf("cache hits=%d  misses=%d  evictions=%d", stats.Hits, stats.Misses, stats.Evictions),
		fmt.Sprintf("lifetime uploads=%d  traffic=%.1f MiB  warm budget=%.1f MiB/frame", stats.Uploads, float64(stats.UploadBytes)/(1024*1024), float64(stats.UploadBudget)/(1024*1024)),
		fmt.Sprintf("warm queue=%d  %.1f MiB", stats.WarmPending, float64(stats.WarmPendingBytes)/(1024*1024)),
	}
	rl.DrawRectangle(8, 8, 510, int32(18*len(lines)+12), rl.Fade(rl.Black, 0.82))
	for index, line := range lines {
		rl.DrawText(line, 14, int32(14+index*18), 14, rl.Lime)
	}
}
