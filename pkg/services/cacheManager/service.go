package cacheManager

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
)

type cacheLookup = map[string]*cache.Cache

type Service struct {
	mesh   servicemesh.Mesh
	logger *slog.Logger
	caches cacheLookup
	mux    sync.Mutex
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh

	s.FlushAllCaches()
}

func (s *Service) Name() string {
	return "Cache Manager"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) FlushAllCaches() {
	s.mux.Lock()
	s.caches = make(cacheLookup)
	s.mux.Unlock()

	for _, service := range s.mesh.Services() {
		s.tryToFlushCacheForService(service)
	}
}

func (s *Service) tryToFlushAllCaches(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		s.tryToFlushCacheForService(service)
		if candidate, ok := service.(HasCaches); ok {
			for idx, destination := range candidate.Caches() {
				newCache := cache.New(destination.CacheBudget())
				destination.FlushCache(newCache)

				s.caches[fmt.Sprintf("%s:%d", service.Name(), idx)] = newCache
			}
		}
	}
}

func (s *Service) tryToFlushCacheForService(service servicemesh.Service) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.caches[service.Name()]; exists {
		return
	}

	candidate, ok := service.(HasCache)
	if !ok {
		return
	}

	s.logger.Info("flushing cache", "service", service.Name())

	if d, hasDeps := candidate.(servicemesh.HasDependencies); hasDeps {
		for !d.DependenciesResolved() {
			time.Sleep(time.Millisecond * 10)
		}
	}

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
