package assetLoader

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed d2_uber_file_list.txt.gz
var uberFileList []byte

func (s *Service) Slug() string {
	return "asset"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", s.handleGetUberFileList)
	group.GET("raw/*path", s.extractAndDownloadFromMpq)
}

func (s *Service) handleGetUberFileList(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", s.extractGzip(uberFileList))
}

func (s *Service) extractAndDownloadFromMpq(c *gin.Context) {
	path := c.Param("path")

	s.logger.Info().Msg(path)

	stream, err := s.mpq.Load(path)
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

func (s *Service) extractGzip(data []byte) []byte {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		s.logger.Fatal().Msg("ExtractTarGz: NewReader failed")
	}

	out := bytes.NewBuffer([]byte{})

	// Copy the decompressed content to the output file
	_, err = io.Copy(out, r)
	if err != nil {
		s.logger.Fatal().Msgf("extracting file: %v", err)
	}

	return out.Bytes()
}
