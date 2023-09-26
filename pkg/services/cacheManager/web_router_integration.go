package cacheManager

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) Slug() string {
	return "cache"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("stats", s.handleGetCacheStats)
	group.GET("flush", s.handleFlushAllCaches)
}

func (s *Service) handleGetCacheStats(c *gin.Context) {
	type CacheStats struct {
		ServiceName string
		Size        string
		Budget      string
		Saturation  string
	}

	cacheStats := make([]CacheStats, 0)

	for name, cache := range s.caches {
		w := cache.GetWeight()
		b := cache.GetBudget()
		stats := CacheStats{
			ServiceName: name,
			Size:        formatBytes(w),
			Budget:      formatBytes(b),
			Saturation:  fmt.Sprintf("%0.2f%%", float64(w)/float64(b)*100),
		}

		cacheStats = append(cacheStats, stats)
	}

	c.JSON(http.StatusOK, cacheStats)
}

func (s *Service) handleFlushAllCaches(c *gin.Context) {
	s.FlushAllCaches()
	c.JSON(http.StatusOK, "caches flushed")
}

func formatBytes(byteCount int) string {
	if byteCount < 0 {
		return "Invalid size"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	if byteCount == 0 {
		return "0 B"
	}

	// Calculate the appropriate unit and value
	unitIndex := int(math.Floor(math.Log(float64(byteCount)) / math.Log(1024)))
	value := float64(byteCount) / math.Pow(1024, float64(unitIndex))

	// Use %g to format the value and trim unnecessary zeros
	return fmt.Sprintf("%.2f %s", value, units[unitIndex])
}
