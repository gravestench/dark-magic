package assetLoader

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Service) Slug() string {
	return "mpq"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("*path", s.extractAndDownloadFromMpq)
}

func (s *Service) extractAndDownloadFromMpq(c *gin.Context) {
	path := c.Param("path")

	s.logger.Info().Msg(path)

	stream, err := s.Load(path)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error().Msgf("loading file: %v", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	filename := filepath.Base(path)

	if strings.HasSuffix(filename, ".txt") {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, string(data))
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/octet-stream", data)
}
