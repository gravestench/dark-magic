package raylibRenderer

import (
	"github.com/gravestench/dark-magic/internal/cache"
)

// CacheBudget converts the human-facing MiB setting into bytes for the cache implementation. Non-positive values retain
// the historical 500 MiB fallback used before configuration is applied.
func (s *Service) CacheBudget() int {
	const (
		kb            = 1024
		mb            = 1024 * kb
		defaultBudget = 500 * mb
	)

	budget := s.config.Cache.BudgetMB
	if budget <= 0 {
		return defaultBudget
	}

	return budget * mb
}

// FlushCache transfers cache ownership to the renderer service. Lifecycle shutdown clears this exact instance so GPU
// eviction callbacks run before the native window disappears.
func (s *Service) FlushCache(newCache *cache.Cache) {
	s.cache = newCache
}

// CacheDiagnostics exposes aggregate cache health without native payloads. A missing cache reports an empty snapshot,
// keeping debug tools safe before renderer configuration.
func (s *Service) CacheDiagnostics() cache.Stats {
	if s.cache == nil {
		return cache.Stats{}
	}

	return s.cache.Diagnostics()
}
