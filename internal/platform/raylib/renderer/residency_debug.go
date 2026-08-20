package raylibRenderer

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// SetResidencyDebug toggles the owner-thread diagnostics overlay through an atomic flag, allowing tools to change it
// without calling native drawing APIs from their goroutine.
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

// applyTextureCacheBudget performs a requested cache resize on the graphics owner thread. Equal or unset budgets avoid
// unnecessary cache work, while SetBudget owns any resulting GPU eviction callbacks.
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

// ResidencyDiagnostics combines cache, upload, composer, and last-frame counters without exposing native GPU handles.
type ResidencyDiagnostics struct {
	Entries, Weight, Budget            int
	Hits, Misses, Evictions            uint64
	Uploads, UploadBytes, UploadBudget uint64
	WarmPending                        int
	WarmPendingBytes                   uint64
	FrameDrawCalls, FrameNodesVisited  uint64
	FrameSubtreesCulled, FrameUpdates  uint64
}

// ResidencyDiagnostics captures a best-effort point-in-time view. Individual counters may advance during collection,
// which is acceptable for diagnostics and avoids blocking frame work with a global lock.
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

// drawResidencyDebug renders diagnostic text in window space after the game surface, keeping it readable regardless of
// logical resolution or palette quantization.
func (s *Service) drawResidencyDebug(composer *render.Composer) {
	if !s.residencyDebug.Load() {
		return
	}

	stats := s.ResidencyDiagnostics(composer)
	lines := residencyDebugLines(stats)

	rl.DrawRectangle(8, 8, 510, int32(18*len(lines)+12), rl.Fade(rl.Black, 0.82))

	for index, line := range lines {
		rl.DrawText(line, 14, int32(14+index*18), 14, rl.Lime)
	}
}

// residencyDebugLines formats counters separately from native drawing, making units and field grouping explicit.
func residencyDebugLines(stats ResidencyDiagnostics) []string {
	lines := []string{
		fmt.Sprintf(
			"frame draws=%d  nodes=%d  culled=%d  texture updates=%d",
			stats.FrameDrawCalls,
			stats.FrameNodesVisited,
			stats.FrameSubtreesCulled,
			stats.FrameUpdates,
		),
		fmt.Sprintf(
			"GPU textures  resident=%d  %.1f/%.1f MiB",
			stats.Entries,
			mebibytes(float64(stats.Weight)),
			mebibytes(float64(stats.Budget)),
		),
		fmt.Sprintf("cache hits=%d  misses=%d  evictions=%d", stats.Hits, stats.Misses, stats.Evictions),
		fmt.Sprintf(
			"lifetime uploads=%d  traffic=%.1f MiB  warm budget=%.1f MiB/frame",
			stats.Uploads,
			mebibytes(float64(stats.UploadBytes)),
			mebibytes(float64(stats.UploadBudget)),
		),
		fmt.Sprintf(
			"warm queue=%d  %.1f MiB",
			stats.WarmPending,
			mebibytes(float64(stats.WarmPendingBytes)),
		),
	}

	return lines
}

// mebibytes converts byte counters to the binary unit used by renderer budgets and diagnostics.
func mebibytes(bytes float64) float64 {
	return bytes / (1024 * 1024)
}
