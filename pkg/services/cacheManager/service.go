package cacheManager

import (
	"sort"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/cache"
)

type cacheLookup = map[string]*cache.Cache

type Service struct {
	runtime runtime.Runtime
	logger  *zerolog.Logger
	caches  cacheLookup
}

func (s *Service) Init(rt runtime.Runtime) {
	s.runtime = rt

	s.FlushAllCaches()
}

func (s *Service) Name() string {
	return "Cache Manager"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) FlushAllCaches() {
	s.caches = make(cacheLookup)

	for _, service := range s.runtime.Services() {
		s.tryToFlushCacheForService(service)
	}
}

func (s *Service) tryToFlushAllCaches(rt runtime.Runtime) {
	for _, service := range rt.Services() {
		s.tryToFlushCacheForService(service)
	}
}

func (s *Service) tryToFlushCacheForService(service runtime.Service) {
	if _, exists := s.caches[service.Name()]; exists {
		return
	}

	candidate, ok := service.(HasCache)
	if !ok {
		return
	}

	s.logger.Info().Msgf("flushing cache for service: %v", service.Name())

	newCache := cache.New(candidate.CacheBudget())
	candidate.FlushCache(newCache)

	s.caches[service.Name()] = newCache
}

type CacheStats struct {
	ServiceName string
	Size        string
	Budget      string
	Saturation  float64
}

func (s *Service) getCacheStats() []CacheStats {
	cacheStats := make([]CacheStats, 0)

	var keys []string
	for name := range s.caches {
		keys = append(keys, name)
	}

	sort.Strings(keys)

	if len(s.caches) < 1 {
		return nil
	}

	for _, name := range keys {
		c := s.caches[name]
		w := c.GetWeight()
		b := c.GetBudget()

		stats := CacheStats{
			ServiceName: name,
			Size:        formatBytes(w),
			Budget:      formatBytes(b),
			Saturation:  float64(w) / float64(b),
		}

		cacheStats = append(cacheStats, stats)
	}

	return cacheStats
}
