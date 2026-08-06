package raylibRenderer

import (
	"github.com/gravestench/dark-magic/internal/cache"
)

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

func (s *Service) FlushCache(newCache *cache.Cache) {
	s.cache = newCache
}

// CacheDiagnostics exposes aggregate cache health without native payloads.
func (s *Service) CacheDiagnostics() cache.Stats {
	if s.cache == nil {
		return cache.Stats{}
	}
	return s.cache.Diagnostics()
}
