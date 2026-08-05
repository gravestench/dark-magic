package assetLoader

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"github.com/gravestench/dark-magic/pkg/assetinspect"
)

//go:embed internal/d2_uber_file_list.txt.gz
var uberFileList []byte

func (s *Service) Slug() string {
	return "asset"
}

func (s *Service) InitRoutes(group *gin.RouterGroup) {
	group.GET("", s.handleGetUberFileList)
	group.GET("raw/*path", s.extractAndDownloadFromMpq)
	group.GET("inspect/*path", s.inspectAsset)
	group.GET("preview/*path", s.previewAsset)
}

func (s *Service) previewAsset(c *gin.Context) {
	path := strings.TrimPrefix(c.Param("path"), "/")
	direction, err := queryIndex(c, "direction")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	frame, err := queryIndex(c, "frame")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stream, err := s.file.Load(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()
	data, err := io.ReadAll(stream)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	preview, err := assetinspect.Preview(fstest.MapFS{path: &fstest.MapFile{Data: data}}, path, direction, frame)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", preview)
}

func queryIndex(c *gin.Context, name string) (int, error) {
	raw := c.DefaultQuery(name, "0")
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return value, nil
}

func (s *Service) inspectAsset(c *gin.Context) {
	path := c.Param("path")
	stream, err := s.file.Load(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	report, err := assetinspect.InspectData(path, data)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (s *Service) handleGetUberFileList(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", s.extractGzip(uberFileList))
}

func (s *Service) extractAndDownloadFromMpq(c *gin.Context) {
	path := c.Param("path")

	stream, err := s.file.Load(path)
	if err != nil {
		s.logger.Error("loading file", "error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error("loading file", "error", err)
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
		s.logger.Error("ExtractTarGz: NewReader failed", "error", err)
		panic(err)
	}

	out := bytes.NewBuffer([]byte{})

	// Copy the decompressed content to the output file
	_, err = io.Copy(out, r)
	if err != nil {
		s.logger.Error("extracting file", "error", err)
		panic(err)
	}

	return out.Bytes()
}
