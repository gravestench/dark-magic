package tblLoader

import (
	"github.com/gravestench/dark-magic/pkg/cache"
)

// CacheBudget implements cacheManager.HasCache
func (s *Service) CacheBudget() int {
	const (
		kb = 1024
		mb = 1024 * kb
	)

	budget := s.config.Group(s.Name()).GetInt(configKeyCacheBudgetMB)
	if budget == 0 {
		budget = 500
	}

	return budget * mb
}

// FlushCache implements cacheManager.HasCache
func (s *Service) FlushCache(newCache *cache.Cache) {
	s.cache = newCache
}
