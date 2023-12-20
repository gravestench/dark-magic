package raylibRenderer

import (
	"github.com/gravestench/dark-magic/pkg/cache"
)

func (s *Service) CacheBudget() int {
	const (
		kb            = 1024
		mb            = 1024 * kb
		defaultBudget = 500 * mb
	)

	budget := s.config.Group(groupKeyCache).GetInt(keyCacheBudget)
	if budget <= 0 {
		return defaultBudget
	}

	return budget * mb
}

func (s *Service) FlushCache(newCache *cache.Cache) {
	s.cache = newCache
}
